package bmclapi

import (
	"runtime"
	"strings"
)

// OSName returns the Mojang library rule OS name for the current platform.
func OSName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}

// LibraryAllowed evaluates Mojang library rules for the current OS.
// Features (is_demo_user / has_custom_resolution) default to false.
func LibraryAllowed(rules []Rule) bool {
	if len(rules) == 0 {
		return true
	}
	allowed := false
	for _, r := range rules {
		action := strings.ToLower(r.Action)
		match := ruleMatches(r)
		if match {
			allowed = action == "allow"
		} else if action == "allow" && r.OS == nil && len(r.Features) == 0 {
			allowed = true
		}
	}
	// Mojang semantics: start disallowed if any rules present; apply in order.
	allowed = false
	for _, r := range rules {
		if !ruleMatches(r) {
			continue
		}
		if strings.EqualFold(r.Action, "allow") {
			allowed = true
		} else {
			allowed = false
		}
	}
	return allowed
}

func ruleMatches(r Rule) bool {
	if len(r.Features) > 0 {
		for _, v := range r.Features {
			if v {
				// we don't enable demo / custom resolution by default
				return false
			}
		}
	}
	if r.OS == nil {
		return true
	}
	if r.OS.Name != "" && r.OS.Name != OSName() {
		return false
	}
	if r.OS.Arch != "" {
		arch := runtime.GOARCH
		if r.OS.Arch == "x86" && arch != "386" {
			return false
		}
	}
	return true
}

// NativeClassifier returns the natives classifier key for current OS, if any.
func NativeClassifier(lib Library) string {
	if lib.Natives == nil {
		return ""
	}
	key, ok := lib.Natives[OSName()]
	if !ok {
		return ""
	}
	// replace ${arch}
	arch := "64"
	if runtime.GOARCH == "386" {
		arch = "32"
	}
	return strings.ReplaceAll(key, "${arch}", arch)
}
