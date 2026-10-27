package bmclapi

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ProgressFunc reports install progress messages.
type ProgressFunc func(msg string)

// InstallOptions controls where game files are placed.
type InstallOptions struct {
	GameDir  string // .minecraft root
	Version  string // version id
	Progress ProgressFunc
}

func (o InstallOptions) progress(msg string) {
	if o.Progress != nil {
		o.Progress(msg)
	}
}

// InstallVersion downloads client jar, libraries, asset index + objects, and extracts natives.
func (c *Client) InstallVersion(ctx context.Context, info VersionInfo, opt InstallOptions) (*VersionJSON, error) {
	if opt.GameDir == "" {
		return nil, fmt.Errorf("game dir empty")
	}
	id := info.ID
	if opt.Version != "" {
		id = opt.Version
	}
	versionsDir := filepath.Join(opt.GameDir, "versions", id)
	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		return nil, err
	}

	opt.progress("正在获取版本元数据 " + id)
	vj, err := c.FetchVersionJSON(ctx, info)
	if err != nil {
		return nil, err
	}
	if vj.ID == "" {
		vj.ID = id
	}

	// Save version JSON locally
	metaPath := filepath.Join(versionsDir, id+".json")
	metaBytes, err := json.MarshalIndent(vj, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(metaPath, metaBytes, 0o644); err != nil {
		return nil, err
	}

	// Client jar
	client := vj.Downloads["client"]
	jarPath := filepath.Join(versionsDir, id+".jar")
	if client.URL != "" {
		opt.progress("正在下载客户端 " + id + ".jar")
		if err := c.DownloadFile(ctx, client.URL, jarPath, client.SHA1, client.Size); err != nil {
			return nil, fmt.Errorf("client jar: %w", err)
		}
	}

	libsDir := filepath.Join(opt.GameDir, "libraries")
	nativesDir := filepath.Join(versionsDir, "natives")
	_ = os.MkdirAll(nativesDir, 0o755)

	for i, lib := range vj.Libraries {
		if !LibraryAllowed(lib.Rules) {
			continue
		}
		opt.progress(fmt.Sprintf("正在下载依赖 (%d/%d) %s", i+1, len(vj.Libraries), lib.Name))
		if err := c.installLibrary(ctx, lib, libsDir, nativesDir); err != nil {
			return nil, fmt.Errorf("library %s: %w", lib.Name, err)
		}
	}

	if vj.AssetIndex != nil {
		opt.progress("正在下载资源索引 " + vj.AssetIndex.ID)
		if err := c.installAssets(ctx, vj, opt.GameDir, opt.Progress); err != nil {
			return nil, err
		}
	}

	opt.progress("版本 " + id + " 安装完成")
	return vj, nil
}

func (c *Client) installLibrary(ctx context.Context, lib Library, libsDir, nativesDir string) error {
	if lib.Downloads != nil && lib.Downloads.Artifact != nil {
		art := lib.Downloads.Artifact
		path := art.Path
		if path == "" {
			path = mavenPath(lib.Name)
		}
		dest := filepath.Join(libsDir, filepath.FromSlash(path))
		url := art.URL
		if url == "" {
			url = LibraryMavenURL(path)
		}
		if err := c.DownloadFile(ctx, url, dest, art.SHA1, art.Size); err != nil {
			return err
		}
	} else if lib.Name != "" && (lib.Downloads == nil || lib.Downloads.Artifact == nil) {
		// legacy: derive path from name
		path := mavenPath(lib.Name)
		dest := filepath.Join(libsDir, filepath.FromSlash(path))
		url := lib.URL
		if url == "" {
			url = LibraryMavenURL(path)
		} else if !strings.HasSuffix(url, ".jar") {
			url = strings.TrimSuffix(url, "/") + "/" + path
		}
		if err := c.DownloadFile(ctx, url, dest, "", 0); err != nil {
			return err
		}
	}

	classifier := NativeClassifier(lib)
	if classifier == "" || lib.Downloads == nil || lib.Downloads.Classifiers == nil {
		return nil
	}
	nat := lib.Downloads.Classifiers[classifier]
	if nat == nil {
		return nil
	}
	path := nat.Path
	if path == "" {
		path = mavenPathWithClassifier(lib.Name, classifier)
	}
	dest := filepath.Join(libsDir, filepath.FromSlash(path))
	url := nat.URL
	if url == "" {
		url = LibraryMavenURL(path)
	}
	if err := c.DownloadFile(ctx, url, dest, nat.SHA1, nat.Size); err != nil {
		return err
	}
	exclude := []string{}
	if lib.Extract != nil {
		exclude = lib.Extract.Exclude
	}
	return extractZip(dest, nativesDir, exclude)
}

func (c *Client) installAssets(ctx context.Context, vj *VersionJSON, gameDir string, progress ProgressFunc) error {
	idx := vj.AssetIndex
	indexesDir := filepath.Join(gameDir, "assets", "indexes")
	objectsDir := filepath.Join(gameDir, "assets", "objects")
	_ = os.MkdirAll(indexesDir, 0o755)
	_ = os.MkdirAll(objectsDir, 0o755)

	indexPath := filepath.Join(indexesDir, idx.ID+".json")
	if err := c.DownloadFile(ctx, idx.URL, indexPath, idx.SHA1, idx.Size); err != nil {
		return fmt.Errorf("asset index: %w", err)
	}
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}
	var ai AssetIndex
	if err := json.Unmarshal(data, &ai); err != nil {
		return err
	}
	n := 0
	total := len(ai.Objects)
	for _, obj := range ai.Objects {
		n++
		if progress != nil && n%50 == 0 {
			progress(fmt.Sprintf("正在下载资源 %d/%d", n, total))
		}
		hash := obj.Hash
		if len(hash) < 2 {
			continue
		}
		dest := filepath.Join(objectsDir, hash[:2], hash)
		if err := c.DownloadFile(ctx, AssetURL(hash), dest, hash, obj.Size); err != nil {
			return fmt.Errorf("asset %s: %w", hash, err)
		}
	}
	return nil
}

func mavenPath(name string) string {
	// group:artifact:version[:classifier]
	parts := strings.Split(name, ":")
	if len(parts) < 3 {
		return name
	}
	group := strings.ReplaceAll(parts[0], ".", "/")
	artifact, version := parts[1], parts[2]
	classifier := ""
	ext := "jar"
	if len(parts) >= 4 {
		classifier = parts[3]
	}
	file := artifact + "-" + version
	if classifier != "" {
		file += "-" + classifier
	}
	file += "." + ext
	return group + "/" + artifact + "/" + version + "/" + file
}

func mavenPathWithClassifier(name, classifier string) string {
	parts := strings.Split(name, ":")
	if len(parts) < 3 {
		return name
	}
	base := parts[0] + ":" + parts[1] + ":" + parts[2] + ":" + classifier
	return mavenPath(base)
}

func extractZip(zipPath, destDir string, exclude []string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		name := f.Name
		skip := false
		for _, ex := range exclude {
			if strings.HasPrefix(name, ex) {
				skip = true
				break
			}
		}
		if skip || strings.HasSuffix(name, "/") {
			continue
		}
		// prevent zip slip
		target := filepath.Join(destDir, filepath.FromSlash(name))
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// FindVersionInfo looks up a version id in the manifest.
func FindVersionInfo(m *Manifest, id string) (VersionInfo, bool) {
	for _, v := range m.Versions {
		if v.ID == id {
			return v, true
		}
	}
	return VersionInfo{}, false
}

// VersionIDs returns all version ids (optionally filtered by type: release/snapshot/old_beta/old_alpha/""=all).
func VersionIDs(m *Manifest, typeFilter string) []string {
	out := make([]string, 0, len(m.Versions))
	for _, v := range m.Versions {
		if typeFilter == "" || typeFilter == "all" || v.Type == typeFilter {
			out = append(out, v.ID)
		}
	}
	return out
}
