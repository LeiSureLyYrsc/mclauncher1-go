// Self-extracting stub for Qt6/QML deployments.
//
// True single-file packers (Enigma Virtual Box, BoxedApp, UPX overlay FS) often
// break Qt Quick: QML modules and plugins need a real filesystem layout
// (platforms/, qml/QtQuick/, …). The reliable pattern is:
//
//  1. Build the app normally and deploy runtime next to it (windeployqt / bundle-linux-qt).
//  2. Zip that directory and append it to this pure-Go stub with an 8-byte size + magic trailer.
//  3. On first run, extract to a hashed cache under os.TempDir(), set Qt env vars, and exec
//     the real binary. Later runs reuse the cache.
//
// Layout of the final artifact:
//
//	[stub bytes][zip payload][uint64 LE payload size][8-byte magic "MCLSFX01"]
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const magic = "MCLSFX01"

func main() {
	payload, err := readTrailingZip()
	if err != nil {
		fatal("read payload: %v", err)
	}

	sum := sha256.Sum256(payload)
	hash := hex.EncodeToString(sum[:8])
	cacheRoot := filepath.Join(os.TempDir(), "mclauncher-sfx-"+hash)
	readyMarker := filepath.Join(cacheRoot, ".sfx-ready")

	if _, err := os.Stat(readyMarker); err != nil {
		if err := extractZip(payload, cacheRoot); err != nil {
			_ = os.RemoveAll(cacheRoot)
			fatal("extract runtime: %v", err)
		}
		if err := os.WriteFile(readyMarker, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644); err != nil {
			fatal("mark ready: %v", err)
		}
	}

	inner := innerBinaryName()
	innerPath := filepath.Join(cacheRoot, inner)
	if _, err := os.Stat(innerPath); err != nil {
		// Allow one nested folder (zip of dir/ vs zip of dir contents).
		matches, _ := filepath.Glob(filepath.Join(cacheRoot, "*", inner))
		if len(matches) == 1 {
			innerPath = matches[0]
			cacheRoot = filepath.Dir(innerPath)
		} else {
			fatal("inner binary not found: %s", innerPath)
		}
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(innerPath, 0o755)
	}

	env := os.Environ()
	env = withQtEnv(env, cacheRoot)

	cmd := exec.Command(innerPath, os.Args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Preserve caller CWD so the app's game-dir resolution (Getwd) stays correct.
	cmd.Dir = ""

	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fatal("run %s: %v", innerPath, err)
	}
}

func innerBinaryName() string {
	if runtime.GOOS == "windows" {
		return "mclauncher.exe"
	}
	return "mclauncher"
}

func withQtEnv(env []string, root string) []string {
	plugins := firstExisting(
		filepath.Join(root, "plugins"),
		filepath.Join(root, "qt6", "plugins"),
		root, // windeployqt often places platforms/ next to the exe
	)
	qml := firstExisting(
		filepath.Join(root, "qml"),
		filepath.Join(root, "qt6", "qml"),
	)

	env = setOrPrepend(env, "QT_PLUGIN_PATH", plugins)
	if qml != "" {
		env = setOrPrepend(env, "QML2_IMPORT_PATH", qml)
		env = setOrPrepend(env, "QML_IMPORT_PATH", qml)
	}

	switch runtime.GOOS {
	case "windows":
		env = setOrPrepend(env, "PATH", root)
	default:
		env = setOrPrepend(env, "LD_LIBRARY_PATH", root)
		// Some distro layouts put libs under lib/
		libDir := filepath.Join(root, "lib")
		if st, err := os.Stat(libDir); err == nil && st.IsDir() {
			env = setOrPrepend(env, "LD_LIBRARY_PATH", libDir)
		}
	}
	return env
}

func firstExisting(candidates ...string) string {
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func setOrPrepend(env []string, key, value string) []string {
	if value == "" {
		return env
	}
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			old := strings.TrimPrefix(e, prefix)
			if old == "" {
				env[i] = prefix + value
			} else if runtime.GOOS == "windows" {
				env[i] = prefix + value + ";" + old
			} else {
				env[i] = prefix + value + ":" + old
			}
			return env
		}
	}
	return append(env, prefix+value)
}

func readTrailingZip() ([]byte, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(self)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	if size < 16 {
		return nil, fmt.Errorf("file too small to contain SFX trailer")
	}

	trailer := make([]byte, 16)
	if _, err := f.ReadAt(trailer, size-16); err != nil {
		return nil, err
	}
	if string(trailer[8:]) != magic {
		return nil, fmt.Errorf("missing SFX magic (not a packed single-file build?)")
	}
	payloadSize := binary.LittleEndian.Uint64(trailer[:8])
	if payloadSize == 0 || int64(payloadSize) > size-16 {
		return nil, fmt.Errorf("invalid payload size %d", payloadSize)
	}
	start := size - 16 - int64(payloadSize)
	buf := make([]byte, payloadSize)
	if _, err := f.ReadAt(buf, start); err != nil {
		return nil, err
	}
	// Sanity: must look like a zip.
	if len(buf) < 4 || buf[0] != 'P' || buf[1] != 'K' {
		return nil, fmt.Errorf("payload is not a zip archive")
	}
	return buf, nil
}

func extractZip(payload []byte, dest string) error {
	tmp := dest + ".extracting"
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o755); err != nil {
		return err
	}

	r, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}
	for _, zf := range r.File {
		if err := writeZipFile(tmp, zf); err != nil {
			return err
		}
	}

	_ = os.RemoveAll(dest)
	if err := os.Rename(tmp, dest); err != nil {
		// Cross-device fallback.
		if err2 := copyDir(tmp, dest); err2 != nil {
			return err
		}
		_ = os.RemoveAll(tmp)
	}
	return nil
}

func writeZipFile(dest string, zf *zip.File) error {
	name := filepath.Clean(zf.Name)
	if name == "." || strings.HasPrefix(name, "..") {
		return fmt.Errorf("unsafe path in zip: %q", zf.Name)
	}
	target := filepath.Join(dest, name)
	if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
		return fmt.Errorf("zip path escapes dest: %q", zf.Name)
	}

	if zf.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	rc, err := zf.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	mode := zf.Mode()
	if mode == 0 {
		mode = 0o644
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "mclauncher-sfx: "+format+"\n", args...)
	os.Exit(1)
}
