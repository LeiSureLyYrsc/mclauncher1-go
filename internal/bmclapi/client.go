package bmclapi

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client talks to BMCLAPI (and follows CDN redirects).
type Client struct {
	HTTP    *http.Client
	BaseURL string
}

func NewClient() *Client {
	return &Client{
		HTTP: &http.Client{
			Timeout: 120 * time.Second,
		},
		BaseURL: BaseURL,
	}
}

func (c *Client) getJSON(ctx context.Context, rawURL string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mclauncher1-go/0.1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("GET %s: HTTP %d: %s", rawURL, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// FetchManifest downloads version_manifest_v2.json from BMCLAPI.
func (c *Client) FetchManifest(ctx context.Context) (*Manifest, error) {
	var m Manifest
	if err := c.getJSON(ctx, ManifestURL(), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// FetchVersionJSON downloads a version metadata document, preferring BMCLAPI mirrors.
func (c *Client) FetchVersionJSON(ctx context.Context, info VersionInfo) (*VersionJSON, error) {
	candidates := []string{}
	if info.URL != "" {
		candidates = append(candidates, MirrorURL(info.URL))
	}
	if info.ID != "" && info.SHA1 != "" {
		candidates = append(candidates, fmt.Sprintf("%s/v1/packages/%s/%s.json", BaseURL, info.SHA1, info.ID))
	}
	if info.ID != "" {
		candidates = append(candidates, fmt.Sprintf("%s/version/%s/json", BaseURL, info.ID))
	}
	var last error
	for _, u := range candidates {
		var v VersionJSON
		if err := c.getJSON(ctx, u, &v); err != nil {
			last = err
			continue
		}
		if v.ID == "" {
			v.ID = info.ID
		}
		return &v, nil
	}
	if last == nil {
		last = fmt.Errorf("no URL for version %s", info.ID)
	}
	return nil, last
}

// DownloadFile downloads url into destPath, optionally verifying sha1 hex.
// Skips download when an existing file already matches sha1/size.
// Retries a few times on transient network/CDN errors.
func (c *Client) DownloadFile(ctx context.Context, rawURL, destPath, sha1hex string, expectedSize int64) error {
	rawURL = MirrorURL(rawURL)
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if st, err := os.Stat(destPath); err == nil && !st.IsDir() {
		if expectedSize > 0 && st.Size() != expectedSize {
			// size mismatch — redownload
		} else if sha1hex != "" {
			ok, _ := fileSHA1Equal(destPath, sha1hex)
			if ok {
				return nil
			}
		} else if expectedSize <= 0 || st.Size() == expectedSize {
			return nil
		}
	}

	var last error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 400 * time.Millisecond):
			}
		}
		last = c.downloadOnce(ctx, rawURL, destPath, sha1hex)
		if last == nil {
			return nil
		}
	}
	return last
}

func (c *Client) downloadOnce(ctx context.Context, rawURL, destPath, sha1hex string) error {
	tmp := destPath + ".part"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "mclauncher1-go/0.1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("download %s: HTTP %d: %s", rawURL, resp.StatusCode, string(b))
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	h := sha1.New()
	w := io.MultiWriter(f, h)
	_, err = io.Copy(w, resp.Body)
	closeErr := f.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}
	if closeErr != nil {
		os.Remove(tmp)
		return closeErr
	}
	if sha1hex != "" {
		sum := hex.EncodeToString(h.Sum(nil))
		if !equalFoldHex(sum, sha1hex) {
			os.Remove(tmp)
			return fmt.Errorf("sha1 mismatch for %s: got %s want %s", destPath, sum, sha1hex)
		}
	}
	return os.Rename(tmp, destPath)
}

func fileSHA1Equal(path, want string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}
	return equalFoldHex(hex.EncodeToString(h.Sum(nil)), want), nil
}

func equalFoldHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'F' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'F' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// EnsureAuthlibInjector downloads authlib-injector.jar into destDir if missing.
func (c *Client) EnsureAuthlibInjector(ctx context.Context, destDir string) (jarPath string, err error) {
	jarPath = filepath.Join(destDir, "authlib-injector.jar")
	if st, err := os.Stat(jarPath); err == nil && st.Size() > 0 {
		return jarPath, nil
	}
	var meta AuthlibLatest
	if err := c.getJSON(ctx, AuthlibLatestJSON(), &meta); err != nil {
		return "", err
	}
	dl := MirrorURL(meta.DownloadURL)
	sha := ""
	if meta.Checksums != nil {
		sha = meta.Checksums["sha1"] // may be empty; sha256 not checked here
	}
	_ = sha
	if err := c.DownloadFile(ctx, dl, jarPath, "", 0); err != nil {
		return "", err
	}
	return jarPath, nil
}
