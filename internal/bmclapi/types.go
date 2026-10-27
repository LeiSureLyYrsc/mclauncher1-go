package bmclapi

// Manifest is Mojang version_manifest_v2.json (mirrored by BMCLAPI).
type Manifest struct {
	Latest   ManifestLatest `json:"latest"`
	Versions []VersionInfo  `json:"versions"`
}

type ManifestLatest struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type VersionInfo struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	Time            string `json:"time"`
	ReleaseTime     string `json:"releaseTime"`
	SHA1            string `json:"sha1"`
	ComplianceLevel int    `json:"complianceLevel"`
}

// VersionJSON is a Minecraft client version metadata document.
type VersionJSON struct {
	ID                     string          `json:"id"`
	Type                   string          `json:"type"`
	MainClass              string          `json:"mainClass"`
	MinecraftArguments     string          `json:"minecraftArguments"`
	Arguments              *Arguments      `json:"arguments"`
	AssetIndex             *AssetIndexRef  `json:"assetIndex"`
	Assets                 string          `json:"assets"`
	Downloads              map[string]File `json:"downloads"`
	Libraries              []Library       `json:"libraries"`
	JavaVersion            *JavaVersion    `json:"javaVersion"`
	ComplianceLevel        int             `json:"complianceLevel"`
	MinimumLauncherVersion int             `json:"minimumLauncherVersion"`
}

type JavaVersion struct {
	Component    string `json:"component"`
	MajorVersion int    `json:"majorVersion"`
}

type Arguments struct {
	Game []any `json:"game"`
	JVM  []any `json:"jvm"`
}

type AssetIndexRef struct {
	ID        string `json:"id"`
	SHA1      string `json:"sha1"`
	Size      int64  `json:"size"`
	TotalSize int64  `json:"totalSize"`
	URL       string `json:"url"`
}

type File struct {
	ID   string `json:"id,omitempty"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
	URL  string `json:"url"`
	Path string `json:"path,omitempty"`
}

type Library struct {
	Name      string             `json:"name"`
	Downloads *LibraryDownloads  `json:"downloads,omitempty"`
	Rules     []Rule             `json:"rules,omitempty"`
	Natives   map[string]string  `json:"natives,omitempty"`
	Extract   *LibraryExtract    `json:"extract,omitempty"`
	URL       string             `json:"url,omitempty"` // legacy
}

type LibraryDownloads struct {
	Artifact    *File            `json:"artifact,omitempty"`
	Classifiers map[string]*File `json:"classifiers,omitempty"`
}

type LibraryExtract struct {
	Exclude []string `json:"exclude"`
}

type Rule struct {
	Action string    `json:"action"`
	OS     *RuleOS   `json:"os,omitempty"`
	Features map[string]bool `json:"features,omitempty"`
}

type RuleOS struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Arch    string `json:"arch"`
}

// AssetIndex is assets/<id>.json
type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

// AuthlibLatest is BMCLAPI / mirrors/authlib-injector/artifact/latest.json
type AuthlibLatest struct {
	BuildNumber int               `json:"build_number"`
	Version     string            `json:"version"`
	DownloadURL string            `json:"download_url"`
	Checksums   map[string]string `json:"checksums"`
}
