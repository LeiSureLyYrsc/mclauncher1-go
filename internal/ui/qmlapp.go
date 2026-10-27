package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"

	"github.com/LeiSureLyYrsc/mclauncher1-go/assets"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/auth"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/bmclapi"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/config"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/java"
	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/launch"
	qmlres "github.com/LeiSureLyYrsc/mclauncher1-go/qml"
)

// retainedBridge keeps Qt objects alive for the process lifetime.
var retainedBridge *bridge

// LoadMainQML writes embedded QML + background JPEG into a temp dir and loads
// them with miqt's QQmlApplicationEngine.
func LoadMainQML() error {
	dir, err := os.MkdirTemp("", "mclauncher-qml-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bg_main.jpg"), assets.BGMainJPG, 0o644); err != nil {
		return fmt.Errorf("write bg: %w", err)
	}

	for _, name := range []string{"Main.qml", "SettingsDialog.qml"} {
		b, err := qmlres.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("embed %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	gameDir := resolveGameDir()
	cfg, _ := config.Load(gameDir)
	cfg.GameDir = gameDir
	if cfg.JavaPath == "" {
		cfg.JavaPath = java.Best()
	}

	bridge := newBridge(cfg)

	engine := qml.NewQQmlApplicationEngine()
	ctx := engine.RootContext()

	bgURL := qt.QUrl_FromLocalFile(filepath.Join(dir, "bg_main.jpg"))
	ctx.SetContextProperty2("bgImageUrl", qt.NewQVariant11(bgURL.ToString()))
	ctx.SetContextProperty2("gameDirPath", qt.NewQVariant11(gameDir))

	// Expose command/status bridge and version list model.
	ctx.SetContextProperty("launcher", bridge.props.QObject)
	ctx.SetContextProperty("versionModel", bridge.versionModel.QObject)

	bridge.applyConfigToProps()
	bridge.refreshVersionsAsync()

	mainPath := filepath.Join(dir, "Main.qml")
	url := qt.QUrl_FromLocalFile(mainPath)

	loadFailed := false
	engine.OnObjectCreationFailed(func(_ *qt.QUrl) {
		loadFailed = true
	})
	engine.OnObjectCreated(func(obj *qt.QObject, _ *qt.QUrl) {
		if obj == nil {
			loadFailed = true
		}
	})

	engine.Load(url)

	if loadFailed || len(engine.RootObjects()) == 0 {
		return fmt.Errorf("failed to load QML: %s", mainPath)
	}

	retainedBridge = bridge
	return nil
}

func resolveGameDir() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	mc := filepath.Join(dir, ".minecraft")
	if st, err := os.Stat(mc); err == nil && st.IsDir() {
		return mc
	}
	// Prefer creating .minecraft beside the binary/cwd for installs.
	_ = os.MkdirAll(mc, 0o755)
	return mc
}

type bridge struct {
	props        *qml.QQmlPropertyMap
	versionModel *qt.QStringListModel
	mu           sync.Mutex
	cfg          config.Config
	busy         bool
	manifest     *bmclapi.Manifest
	statusCh     chan string
	timer        *qt.QTimer
}

func newBridge(cfg config.Config) *bridge {
	b := &bridge{
		props:        qml.NewQQmlPropertyMap(),
		versionModel: qt.NewQStringListModel2([]string{"(加载中…)"}),
		cfg:          cfg,
		statusCh:     make(chan string, 64),
	}
	b.props.OnValueChanged(func(key string, value *qt.QVariant) {
		b.onPropChanged(key, value)
	})

	b.timer = qt.NewQTimer()
	b.timer.OnTimeout(func() {
		b.drainStatus()
	})
	b.timer.Start(100)

	return b
}

func (b *bridge) set(key, val string) {
	b.props.Insert(key, qt.NewQVariant11(val))
}

func (b *bridge) setBool(key string, v bool) {
	b.props.Insert(key, qt.NewQVariant8(v))
}

func (b *bridge) setInt(key string, v int) {
	b.props.Insert(key, qt.NewQVariant4(v))
}

func (b *bridge) get(key string) string {
	return b.props.Value(key).ToString()
}

func (b *bridge) applyConfigToProps() {
	b.set("authMode", b.cfg.AuthMode)
	b.set("offlineName", b.cfg.OfflineName)
	b.set("authlibAPI", b.cfg.AuthlibAPI)
	b.set("authlibUser", b.cfg.AuthlibUser)
	b.set("authlibPass", b.cfg.AuthlibPass)
	b.set("javaPath", b.cfg.JavaPath)
	b.setInt("maxMemoryMB", b.cfg.MaxMemoryMB)
	b.set("selectedVersion", b.cfg.SelectedVer)
	b.set("versionFilter", b.cfg.VersionFilter)
	b.setBool("debug", b.cfg.Debug)
	b.set("gameDir", b.cfg.GameDir)
	b.set("status", "就绪")
	b.set("busy", "0")
	b.set("cmd", "")
	b.set("lastError", "")
}

func (b *bridge) postStatus(msg string) {
	select {
	case b.statusCh <- msg:
	default:
	}
}

func (b *bridge) setBusy(v bool) {
	b.mu.Lock()
	b.busy = v
	b.mu.Unlock()
	if v {
		b.set("busy", "1")
	} else {
		b.set("busy", "0")
	}
}

func (b *bridge) onPropChanged(key string, value *qt.QVariant) {
	switch key {
	case "cmd":
		cmd := strings.TrimSpace(value.ToString())
		if cmd == "" {
			return
		}
		// clear cmd so QML can re-trigger same command
		b.set("cmd", "")
		b.handleCmd(cmd)
	case "selectedVersion":
		b.mu.Lock()
		b.cfg.SelectedVer = value.ToString()
		b.mu.Unlock()
	case "versionFilter":
		b.mu.Lock()
		b.cfg.VersionFilter = value.ToString()
		b.mu.Unlock()
		b.applyVersionFilter()
	}
}

func (b *bridge) handleCmd(cmd string) {
	switch cmd {
	case "saveSettings":
		b.saveSettingsFromProps()
	case "sniffJava":
		path := java.Best()
		if path == "" {
			b.postStatus("未找到 Java，请手动指定路径")
			b.set("lastError", "未找到 Java")
			return
		}
		b.set("javaPath", path)
		b.mu.Lock()
		b.cfg.JavaPath = path
		b.mu.Unlock()
		b.postStatus("已嗅探到 Java: " + path)
	case "testLogin":
		go b.cmdTestLogin()
	case "launch":
		go b.cmdLaunch()
	case "update":
		go b.cmdUpdate()
	case "refreshVersions":
		b.refreshVersionsAsync()
	case "clearAuth":
		b.set("authlibUser", "")
		b.set("authlibPass", "")
		b.set("authMode", auth.ModeOffline)
		b.mu.Lock()
		b.cfg.AuthlibUser = ""
		b.cfg.AuthlibPass = ""
		b.cfg.AccessToken = ""
		b.cfg.ProfileName = ""
		b.cfg.ProfileUUID = ""
		b.cfg.AuthMode = auth.ModeOffline
		_ = config.Save(b.cfg)
		b.mu.Unlock()
		b.postStatus("已清除登录信息，切换为离线")
	default:
		b.postStatus("未知命令: " + cmd)
	}
}

func (b *bridge) saveSettingsFromProps() {
	b.mu.Lock()
	b.cfg.AuthMode = b.get("authMode")
	if b.cfg.AuthMode == "" {
		b.cfg.AuthMode = auth.ModeOffline
	}
	b.cfg.OfflineName = b.get("offlineName")
	if b.cfg.OfflineName == "" {
		b.cfg.OfflineName = "Player"
	}
	b.cfg.AuthlibAPI = strings.TrimSpace(b.get("authlibAPI"))
	b.cfg.AuthlibUser = b.get("authlibUser")
	b.cfg.AuthlibPass = b.get("authlibPass")
	b.cfg.JavaPath = b.get("javaPath")
	if v := b.props.Value("maxMemoryMB").ToInt(); v > 0 {
		b.cfg.MaxMemoryMB = v
	}
	b.cfg.Debug = b.get("debug") == "1" || b.get("debug") == "true"
	// also accept bool variant
	if b.props.Value("debug").ToBool() {
		b.cfg.Debug = true
	}
	b.cfg.SelectedVer = b.get("selectedVersion")
	b.cfg.VersionFilter = b.get("versionFilter")
	cfg := b.cfg
	b.mu.Unlock()
	if err := config.Save(cfg); err != nil {
		b.postStatus("保存失败: " + err.Error())
		b.set("lastError", err.Error())
		return
	}
	b.postStatus("设置已保存")
}

func (b *bridge) refreshVersionsAsync() {
	go func() {
		b.postStatus("正在从 BMCLAPI 拉取版本列表…")
		c := bmclapi.NewClient()
		m, err := c.FetchManifest(context.Background())
		if err != nil {
			b.postStatus("版本列表失败: " + err.Error())
			b.set("lastError", err.Error())
			return
		}
		b.mu.Lock()
		b.manifest = m
		b.mu.Unlock()
		// Update model on next timer tick via queued string — SetStringList must be main thread.
		// Use status channel pattern: stash ids and apply in drain via special message.
		b.postStatus("__versions_ready__")
		b.postStatus(fmt.Sprintf("已加载 %d 个版本（BMCLAPI）", len(m.Versions)))
	}()
}

func (b *bridge) drainStatus() {
	for {
		select {
		case msg := <-b.statusCh:
			if msg == "__versions_ready__" {
				b.applyVersionFilter()
				continue
			}
			b.set("status", msg)
		default:
			return
		}
	}
}

func (b *bridge) applyVersionFilter() {
	b.mu.Lock()
	m := b.manifest
	filter := b.cfg.VersionFilter
	selected := b.cfg.SelectedVer
	b.mu.Unlock()
	if m == nil {
		return
	}
	if filter == "" {
		filter = "release"
	}
	ids := bmclapi.VersionIDs(m, filter)
	if filter == "all" {
		ids = bmclapi.VersionIDs(m, "")
	}
	if len(ids) == 0 {
		ids = []string{"(无版本)"}
	}
	b.versionModel.SetStringList(ids)
	if selected == "" && m.Latest.Release != "" {
		selected = m.Latest.Release
		b.mu.Lock()
		b.cfg.SelectedVer = selected
		b.mu.Unlock()
	}
	if selected != "" {
		b.set("selectedVersion", selected)
	}
}

func (b *bridge) cmdTestLogin() {
	b.setBusy(true)
	defer b.setBusy(false)
	api := b.get("authlibAPI")
	user := b.get("authlibUser")
	pass := b.get("authlibPass")
	if api == "" || user == "" || pass == "" {
		b.postStatus("请填写 Authlib API 地址、用户名和密码")
		b.set("lastError", "缺少登录信息")
		return
	}
	b.postStatus("正在登录 Authlib Injector…")
	cli := auth.NewAuthlibClient(api)
	b.mu.Lock()
	ct := b.cfg.ClientToken
	b.mu.Unlock()
	sess, err := cli.Login(context.Background(), user, pass, ct)
	if err != nil {
		b.postStatus("登录失败: " + err.Error())
		b.set("lastError", err.Error())
		return
	}
	b.mu.Lock()
	b.cfg.AuthMode = auth.ModeAuthlib
	b.cfg.AuthlibAPI = api
	b.cfg.AuthlibUser = user
	b.cfg.AuthlibPass = pass
	b.cfg.AccessToken = sess.AccessToken
	b.cfg.ClientToken = sess.ClientToken
	b.cfg.ProfileUUID = sess.UUID
	b.cfg.ProfileName = sess.Username
	_ = config.Save(b.cfg)
	b.mu.Unlock()
	b.set("authMode", auth.ModeAuthlib)
	b.postStatus("登录成功: " + sess.Username)
}

func (b *bridge) cmdUpdate() {
	b.mu.Lock()
	if b.busy {
		b.mu.Unlock()
		b.postStatus("正在忙，请稍候…")
		return
	}
	b.mu.Unlock()
	b.setBusy(true)
	defer b.setBusy(false)

	b.saveSettingsFromProps()
	b.mu.Lock()
	cfg := b.cfg
	ver := cfg.SelectedVer
	b.mu.Unlock()
	ver = b.get("selectedVersion")
	if ver == "" || strings.HasPrefix(ver, "(") {
		b.postStatus("请先选择版本")
		return
	}
	b.postStatus("开始更新 " + ver + " …")
	_, err := launch.UpdateVersion(context.Background(), cfg, ver, func(msg string) {
		b.postStatus(msg)
	})
	if err != nil {
		b.postStatus("更新失败: " + err.Error())
		b.set("lastError", err.Error())
		return
	}
	b.postStatus("更新完成: " + ver)
}

func (b *bridge) cmdLaunch() {
	b.mu.Lock()
	if b.busy {
		b.mu.Unlock()
		b.postStatus("正在忙，请稍候…")
		return
	}
	b.mu.Unlock()
	b.setBusy(true)
	defer b.setBusy(false)

	b.saveSettingsFromProps()
	b.mu.Lock()
	cfg := b.cfg
	b.mu.Unlock()
	ver := b.get("selectedVersion")
	if ver == "" || strings.HasPrefix(ver, "(") {
		b.postStatus("请先选择版本")
		return
	}
	if cfg.JavaPath == "" {
		cfg.JavaPath = java.Best()
		b.set("javaPath", cfg.JavaPath)
	}
	cmd, err := launch.Launch(context.Background(), cfg, ver, func(msg string) {
		b.postStatus(msg)
	})
	if err != nil {
		b.postStatus("启动失败: " + err.Error())
		b.set("lastError", err.Error())
		return
	}
	b.postStatus(fmt.Sprintf("游戏已启动 (pid %d)", cmd.Process.Pid))
	go func() {
		_ = cmd.Wait()
		b.postStatus("游戏进程已退出")
	}()
}
