package java

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Installation describes a discovered Java runtime.
type Installation struct {
	Path    string // path to java / javaw binary
	Version string // raw version line
}

// Detect searches JAVA_HOME, PATH, and common install locations.
func Detect() []Installation {
	seen := map[string]bool{}
	var out []Installation

	add := func(bin string) {
		bin = strings.TrimSpace(bin)
		if bin == "" {
			return
		}
		abs, err := filepath.Abs(bin)
		if err == nil {
			bin = abs
		}
		if seen[bin] {
			return
		}
		if st, err := os.Stat(bin); err != nil || st.IsDir() {
			return
		}
		ver := probeVersion(bin)
		seen[bin] = true
		out = append(out, Installation{Path: bin, Version: ver})
	}

	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		add(filepath.Join(jh, "bin", javaExeName()))
	}
	if p, err := exec.LookPath(javaExeName()); err == nil {
		add(p)
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("javaw.exe"); err == nil {
			add(p)
		}
	}

	for _, dir := range commonRoots() {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				base := strings.ToLower(d.Name())
				// prune deep unrelated trees a bit
				if base == "node_modules" || base == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			name := d.Name()
			if name == javaExeName() || name == "javaw.exe" || name == "javaw" {
				// prefer .../bin/java
				if strings.EqualFold(filepath.Base(filepath.Dir(path)), "bin") {
					add(path)
				}
			}
			return nil
		})
	}
	return out
}

// Best returns the preferred Java binary path, or empty.
func Best() string {
	list := Detect()
	if len(list) == 0 {
		return ""
	}
	// Prefer non-empty version strings; first hit is fine.
	for _, j := range list {
		if j.Version != "" {
			return j.Path
		}
	}
	return list[0].Path
}

func javaExeName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func probeVersion(bin string) string {
	cmd := exec.Command(bin, "-version")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run()
	lines := strings.Split(buf.String(), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

func commonRoots() []string {
	var roots []string
	switch runtime.GOOS {
	case "windows":
		roots = []string{
			`C:\Program Files\Java`,
			`C:\Program Files\Eclipse Adoptium`,
			`C:\Program Files\Microsoft`,
			`C:\Program Files\Zulu`,
			`C:\Program Files\Amazon Corretto`,
			`C:\Program Files (x86)\Java`,
		}
	case "darwin":
		roots = []string{
			"/Library/Java/JavaVirtualMachines",
			"/opt/homebrew/opt",
			"/usr/local/opt",
		}
	default: // linux etc.
		roots = []string{
			"/usr/lib/jvm",
			"/usr/lib64/jvm",
			"/usr/local/lib/jvm",
			"/opt/java",
			"/opt/jdk",
			filepath.Join(os.Getenv("HOME"), ".sdkman/candidates/java"),
			filepath.Join(os.Getenv("HOME"), ".jdks"),
		}
	}
	var exist []string
	for _, r := range roots {
		if r == "" {
			continue
		}
		if st, err := os.Stat(r); err == nil && st.IsDir() {
			exist = append(exist, r)
		}
	}
	return exist
}
