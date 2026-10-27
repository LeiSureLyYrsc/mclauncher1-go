// Self-extracting stub for Qt6/QML deployments.
//
// True single-file packers (Enigma Virtual Box, BoxedApp, UPX overlay FS) often
// break Qt Quick: QML modules and plugins need a real filesystem layout
// (platforms/, qml/QtQuick/, …). The reliable pattern is:
//
//  1. Build the app normally and deploy runtime next to it (windeployqt / bundle-linux-qt).
//  2. Zip that directory and append it to this pure-Go stub with an 8-byte size + magic trailer.
//  3. On first run, extract to a hashed cache, set Qt env vars, and exec the real binary.
//     Later runs reuse the cache.
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
	cacheRoot := filepath.Join(cacheBaseDir(), "mclauncher-sfx-"+hash)
	readyMarker := filepath.Join(cacheRoot, ".sfx-ready")

	if _, err := os.Stat(readyMarker); err != nil {
		if err := extractZip(payload, cacheRoot); err != nil {
			_ = os.RemoveAll(cacheRoot)
			fatal("extract runtime: %v", err)
		}
		ensureQtConf(cacheRoot)
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
			fatal("inner binary not found: %s\n(cache=%s)", innerPath, cacheRoot)
		}
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(innerPath, 0o755)
	}

	// Preflight: common silent-failure causes on Windows.
	if runtime.GOOS == "windows" {
		platformDLL := filepath.Join(cacheRoot, "platforms", "qwindows.dll")
		if _, err := os.Stat(platformDLL); err != nil {
			fatal("missing platforms/qwindows.dll beside app\nlooked at: %s\ncache: %s", platformDLL, cacheRoot)
		}
	}

	env := os.Environ()
	// Remember the user's launch CWD for game-dir resolution inside the app.
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		env = setOrReplace(env, "MCLAUNCHER_CWD", cwd)
	}
	env = withQtEnv(env, cacheRoot)

	cmd := exec.Command(innerPath, os.Args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run with extract dir as CWD so relative Qt plugin lookups work; the app
	// should prefer MCLAUNCHER_CWD (or a stable game dir) over Getwd().
	cmd.Dir = cacheRoot

	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code := ee.ExitCode()
			// DLL-load / NTSTATUS style failures often look like huge unsigned codes.
			if code != 0 {
				fatal("child exited with code %d (%s)\ncache: %s\nIf a DLL is missing, re-download the latest artifact.", code, innerPath, cacheRoot)
			}
			os.Exit(code)
		}
		fatal("failed to start %s: %v\ncache: %s", innerPath, err, cacheRoot)
	}
}

func cacheBaseDir() string {
	// Prefer LocalAppData on Windows — TEMP is often wiped / AV-hostile for unpacked exes.
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "mclauncher")
		}
	}
	if runtime.GOOS != "windows" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, ".cache", "mclauncher")
		}
	}
	return os.TempDir()
}

func innerBinaryName() string {
	if runtime.GOOS == "windows" {
		return "mclauncher.exe"
	}
	return "mclauncher"
}

func withQtEnv(env []string, root string) []string {
	plugins := firstExistingDir(
		filepath.Join(root, "plugins"),
		filepath.Join(root, "qt6", "plugins"),
		root, // windeployqt places platforms/ next to the exe
	)
	qml := firstExistingDir(
		filepath.Join(root, "qml"),
		filepath.Join(root, "qt6", "qml"),
	)
	platforms := firstExistingDir(
		filepath.Join(root, "platforms"),
		filepath.Join(root, "plugins", "platforms"),
		filepath.Join(root, "qt6", "plugins", "platforms"),
	)

	if plugins != "" {
		env = setOrReplace(env, "QT_PLUGIN_PATH", plugins)
	}
	if platforms != "" {
		// Most important for "could not find platform plugin windows".
		env = setOrReplace(env, "QT_QPA_PLATFORM_PLUGIN_PATH", platforms)
	}
	if qml != "" {
		env = setOrReplace(env, "QML2_IMPORT_PATH", qml)
		env = setOrReplace(env, "QML_IMPORT_PATH", qml)
	}

	switch runtime.GOOS {
	case "windows":
		env = setOrPrepend(env, "PATH", root)
	default:
		env = setOrPrepend(env, "LD_LIBRARY_PATH", root)
		libDir := filepath.Join(root, "lib")
		if st, err := os.Stat(libDir); err == nil && st.IsDir() {
			env = setOrPrepend(env, "LD_LIBRARY_PATH", libDir)
		}
	}
	return env
}

func ensureQtConf(root string) {
	path := filepath.Join(root, "qt.conf")
	if _, err := os.Stat(path); err == nil {
		return
	}
	// windeployqt layout: platforms/ and qml/ sit next to the exe.
	content := "[Paths]\nPrefix=.\nPlugins=.\nQmlImports=qml\nLibraries=.\n"
	_ = os.WriteFile(path, []byte(content), 0o644)
}

func firstExistingDir(candidates ...string) string {
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return ""
}

func setOrReplace(env []string, key, value string) []string {
	if value == "" {
		return env
	}
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
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
	// EvalSymlinks can fail on some Windows paths; fall back to raw path.
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
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
	msg := fmt.Sprintf(format, args...)
	line := "mclauncher-sfx: " + msg + "\n"
	fmt.Fprint(os.Stderr, line)
	writeFailureLog(msg)
	showErrorDialog(msg)
	os.Exit(1)
}

func writeFailureLog(msg string) {
	stamp := time.Now().Format("2006-01-02 15:04:05")
	body := fmt.Sprintf("[%s] %s\n", stamp, msg)
	candidates := []string{}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "mclauncher-sfx-error.log"))
	}
	candidates = append(candidates, filepath.Join(os.TempDir(), "mclauncher-sfx-error.log"))
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			candidates = append(candidates, filepath.Join(base, "mclauncher", "mclauncher-sfx-error.log"))
		}
	}
	for _, path := range candidates {
		_ = os.MkdirAll(filepath.Dir(path), 0o755)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			continue
		}
		_, _ = f.WriteString(body)
		_ = f.Close()
	}
}
