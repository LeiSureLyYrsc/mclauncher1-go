package ui

import (
	"fmt"
	"os"
	"path/filepath"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"

	"github.com/LeiSureLyYrsc/mclauncher1-go/assets"
	qmlres "github.com/LeiSureLyYrsc/mclauncher1-go/qml"
)

// LoadMainQML writes embedded QML + background JPEG into a temp dir and loads
// them with miqt's QQmlApplicationEngine. Prefer this over Widgets for the
// main window (see widgets_legacy/ for the prior QWidget implementation).
func LoadMainQML() error {
	dir, err := os.MkdirTemp("", "mclauncher-qml-*")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	// Keep the temp dir for the process lifetime; OS cleans on reboot.
	// (Deleting early would break live QML Image / import resolution.)

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

	engine := qml.NewQQmlApplicationEngine()

	bgURL := qt.QUrl_FromLocalFile(filepath.Join(dir, "bg_main.jpg"))
	engine.RootContext().SetContextProperty2("bgImageUrl", qt.NewQVariant11(bgURL.ToString()))

	gameDir := resolveGameDir()
	engine.RootContext().SetContextProperty2("gameDirPath", qt.NewQVariant11(gameDir))

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
	return dir
}
