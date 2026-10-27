package launch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/auth"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/bmclapi"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/config"
)

// Progress reports status text.
type Progress func(msg string)

// PrepareSession resolves offline or authlib session from config.
func PrepareSession(ctx context.Context, cfg config.Config) (auth.Session, error) {
	switch cfg.AuthMode {
	case auth.ModeAuthlib:
		if cfg.AccessToken != "" && cfg.ProfileName != "" && cfg.ProfileUUID != "" {
			return auth.Session{
				Mode:        auth.ModeAuthlib,
				Username:    cfg.ProfileName,
				UUID:        cfg.ProfileUUID,
				AccessToken: cfg.AccessToken,
				ClientToken: cfg.ClientToken,
				UserType:    "mojang",
				AuthlibAPI:  cfg.AuthlibAPI,
			}, nil
		}
		if cfg.AuthlibAPI == "" || cfg.AuthlibUser == "" || cfg.AuthlibPass == "" {
			// fall back to offline rather than hard-fail
			return auth.OfflineSession(cfg.OfflineName), nil
		}
		cli := auth.NewAuthlibClient(cfg.AuthlibAPI)
		sess, err := cli.Login(ctx, cfg.AuthlibUser, cfg.AuthlibPass, cfg.ClientToken)
		if err != nil {
			return auth.Session{}, err
		}
		return sess, nil
	default:
		return auth.OfflineSession(cfg.OfflineName), nil
	}
}

// UpdateVersion installs/updates the selected version via BMCLAPI.
func UpdateVersion(ctx context.Context, cfg config.Config, versionID string, progress Progress) (*bmclapi.VersionJSON, error) {
	if versionID == "" {
		return nil, fmt.Errorf("未选择版本")
	}
	c := bmclapi.NewClient()
	m, err := c.FetchManifest(ctx)
	if err != nil {
		return nil, err
	}
	info, ok := bmclapi.FindVersionInfo(m, versionID)
	if !ok {
		info = bmclapi.VersionInfo{ID: versionID}
	}
	pf := Progress(func(msg string) {})
	if progress != nil {
		pf = progress
	}
	return c.InstallVersion(ctx, info, bmclapi.InstallOptions{
		GameDir:  cfg.GameDir,
		Version:  versionID,
		Progress: bmclapi.ProgressFunc(pf),
	})
}

// Launch starts the game process. If version files are missing, returns an error suggesting update.
func Launch(ctx context.Context, cfg config.Config, versionID string, progress Progress) (*exec.Cmd, error) {
	if progress == nil {
		progress = func(string) {}
	}
	if versionID == "" {
		versionID = cfg.SelectedVer
	}
	if versionID == "" {
		return nil, fmt.Errorf("未选择版本")
	}
	metaPath := filepath.Join(cfg.GameDir, "versions", versionID, versionID+".json")
	jarPath := filepath.Join(cfg.GameDir, "versions", versionID, versionID+".jar")
	if _, err := os.Stat(metaPath); err != nil {
		return nil, fmt.Errorf("版本 %s 未安装，请先点击「更新游戏」", versionID)
	}
	if _, err := os.Stat(jarPath); err != nil {
		return nil, fmt.Errorf("缺少客户端 jar，请先「更新游戏」")
	}

	raw, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}
	var vj bmclapi.VersionJSON
	if err := json.Unmarshal(raw, &vj); err != nil {
		return nil, err
	}

	progress("正在准备登录会话…")
	sess, err := PrepareSession(ctx, cfg)
	if err != nil {
		return nil, err
	}

	authlibJar := ""
	if sess.Mode == auth.ModeAuthlib {
		progress("正在检查 Authlib Injector…")
		c := bmclapi.NewClient()
		authlibJar, err = c.EnsureAuthlibInjector(ctx, filepath.Join(cfg.GameDir, ".mclauncher"))
		if err != nil {
			return nil, fmt.Errorf("authlib-injector: %w", err)
		}
	}

	progress("正在组装启动参数…")
	cmdSpec, err := Build(Options{
		GameDir:     cfg.GameDir,
		VersionID:   versionID,
		JavaPath:    cfg.JavaPath,
		MaxMemoryMB: cfg.MaxMemoryMB,
		Session:     sess,
		AuthlibJar:  authlibJar,
		Debug:       cfg.Debug,
		Version:     &vj,
	})
	if err != nil {
		return nil, err
	}

	progress(fmt.Sprintf("正在启动 %s（%s）…", versionID, sess.DisplayName()))
	cmd := exec.CommandContext(ctx, cmdSpec.Java, cmdSpec.Args...)
	cmd.Dir = cmdSpec.Dir
	if cfg.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		logPath := filepath.Join(cfg.GameDir, ".mclauncher", "game.log")
		_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
		f, err := os.Create(logPath)
		if err == nil {
			cmd.Stdout = f
			cmd.Stderr = f
		}
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动失败: %w", err)
	}
	return cmd, nil
}

