package bmclapi

import (
	"net/url"
	"strings"
)

// BaseURL is the preferred BMCLAPI mirror root.
// Docs: https://bmclapidoc.bangbang93.com/
const BaseURL = "https://bmclapi2.bangbang93.com"

const (
	pathManifestV2 = "/mc/game/version_manifest_v2.json"
	pathManifestV1 = "/mc/game/version_manifest.json"
	pathAuthlibAPI = "/mirrors/authlib-injector"
)

// ManifestURL returns the version_manifest_v2.json mirror URL.
func ManifestURL() string {
	return BaseURL + pathManifestV2
}

// AuthlibInjectorAPI returns the BMCLAPI mirror root for authlib-injector artifacts.
func AuthlibInjectorAPI() string {
	return BaseURL + pathAuthlibAPI
}

// AuthlibLatestJSON is the latest authlib-injector metadata endpoint (BMCLAPI).
func AuthlibLatestJSON() string {
	return AuthlibInjectorAPI() + "/artifact/latest.json"
}

// MirrorURL rewrites known Mojang / upstream hosts to BMCLAPI.
func MirrorURL(raw string) string {
	if raw == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return rewriteHostString(raw)
	}
	host := strings.ToLower(u.Host)
	switch {
	case host == "launchermeta.mojang.com" || host == "launcher.mojang.com" ||
		host == "piston-meta.mojang.com" || host == "piston-data.mojang.com":
		u.Scheme = "https"
		u.Host = "bmclapi2.bangbang93.com"
		return u.String()
	case host == "libraries.minecraft.net":
		u.Scheme = "https"
		u.Host = "bmclapi2.bangbang93.com"
		u.Path = "/maven" + u.Path
		return u.String()
	case host == "resources.download.minecraft.net":
		u.Scheme = "https"
		u.Host = "bmclapi2.bangbang93.com"
		u.Path = "/assets" + u.Path
		return u.String()
	case host == "authlib-injector.yushi.moe":
		u.Scheme = "https"
		u.Host = "bmclapi2.bangbang93.com"
		if !strings.HasPrefix(u.Path, "/mirrors/authlib-injector") {
			u.Path = "/mirrors/authlib-injector" + u.Path
		}
		return u.String()
	default:
		return raw
	}
}

func rewriteHostString(raw string) string {
	repls := []struct{ old, neu string }{
		{"https://launchermeta.mojang.com", BaseURL},
		{"http://launchermeta.mojang.com", BaseURL},
		{"https://launcher.mojang.com", BaseURL},
		{"http://launcher.mojang.com", BaseURL},
		{"https://piston-meta.mojang.com", BaseURL},
		{"https://piston-data.mojang.com", BaseURL},
		{"https://libraries.minecraft.net/", BaseURL + "/maven/"},
		{"http://libraries.minecraft.net/", BaseURL + "/maven/"},
		{"https://resources.download.minecraft.net/", BaseURL + "/assets/"},
		{"http://resources.download.minecraft.net/", BaseURL + "/assets/"},
		{"https://authlib-injector.yushi.moe/", BaseURL + "/mirrors/authlib-injector/"},
	}
	out := raw
	for _, r := range repls {
		out = strings.ReplaceAll(out, r.old, r.neu)
	}
	return out
}

// LibraryMavenURL builds a BMCLAPI maven URL from a relative path.
func LibraryMavenURL(path string) string {
	path = strings.TrimPrefix(path, "/")
	return BaseURL + "/maven/" + path
}

// AssetURL builds a BMCLAPI assets URL from hash prefix/hash.
func AssetURL(hash string) string {
	if len(hash) < 2 {
		return BaseURL + "/assets/" + hash
	}
	return BaseURL + "/assets/" + hash[:2] + "/" + hash
}
