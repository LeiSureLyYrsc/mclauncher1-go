package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is persisted launcher settings (JSON under gameDir/.mclauncher/config.json).
// Authlib scheme 1: never persist passwords — only tokens + profile + display username.
type Config struct {
	AuthMode      string `json:"authMode"` // offline | authlib
	OfflineName   string `json:"offlineName"`
	AuthlibAPI    string `json:"authlibAPI"`
	AuthlibUser   string `json:"authlibUser"` // email/username for display & re-login; not a secret
	AccessToken   string `json:"accessToken,omitempty"`
	ClientToken   string `json:"clientToken,omitempty"`
	RefreshToken  string `json:"refreshToken,omitempty"` // reserved; Yggdrasil refresh uses access+client
	ProfileUUID   string `json:"profileUUID,omitempty"`
	ProfileName   string `json:"profileName,omitempty"`
	JavaPath      string `json:"javaPath"`
	MaxMemoryMB   int    `json:"maxMemoryMB"`
	SelectedVer   string `json:"selectedVersion"`
	GameDir       string `json:"gameDir"`
	Debug         bool   `json:"debug"`
	VersionFilter string `json:"versionFilter"` // all | release | snapshot
}

func Default() Config {
	return Config{
		AuthMode:      "offline",
		OfflineName:   "Player",
		MaxMemoryMB:   2048,
		VersionFilter: "release",
	}
}

func Path(gameDir string) string {
	return filepath.Join(gameDir, ".mclauncher", "config.json")
}

func Load(gameDir string) (Config, error) {
	cfg := Default()
	cfg.GameDir = gameDir
	p := Path(gameDir)
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	// Decode into a map first so legacy authlibPass is never kept on the struct.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return cfg, err
	}
	delete(raw, "authlibPass")
	cleaned, err := json.Marshal(raw)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(cleaned, &cfg); err != nil {
		return cfg, err
	}
	if cfg.GameDir == "" {
		cfg.GameDir = gameDir
	}
	if cfg.MaxMemoryMB <= 0 {
		cfg.MaxMemoryMB = 2048
	}
	if cfg.AuthMode == "" {
		cfg.AuthMode = "offline"
	}
	if cfg.OfflineName == "" {
		cfg.OfflineName = "Player"
	}
	return cfg, nil
}

func Save(cfg Config) error {
	dir := filepath.Dir(Path(cfg.GameDir))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// Explicit omit: never write a password field, even if an older build left one on disk.
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(cfg.GameDir), b, 0o600)
}
