package ui

import (
	"os"
	"path/filepath"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LeiSureLyYrsc/mclauncher1-go/assets"
)

const (
	winW = 720
	winH = 450
)

// Layout measured from original 720×450 Wine screenshot (right control cluster).
var (
	geoLaunch   = rect{400, 283, 88, 30}
	geoUpdate   = rect{492, 283, 52, 14}
	geoSettings = rect{492, 299, 52, 14}
	geoCombo    = rect{401, 320, 100, 16}
	geoOpenDir  = rect{510, 318, 72, 18}
	geoFooter   = rect{8, 428, 360, 18}
	geoClose    = rect{692, 4, 22, 20}
)

type rect struct{ x, y, w, h int }

// MainWindow is the frameless 720×450 skin using QLabel + real widgets (no full-window QPainter).
type MainWindow struct {
	*qt.QWidget

	bg        *qt.QLabel
	btnLaunch *qt.QPushButton
	btnUpdate *qt.QPushButton
	btnSettings *qt.QPushButton
	btnOpenDir  *qt.QPushButton
	btnClose    *qt.QPushButton
	combo       *qt.QComboBox
	footer      *qt.QLabel

	dragPos *qt.QPoint
	dragging bool
}

// NewMainWindow builds the 1:1 main UI shell.
func NewMainWindow() *MainWindow {
	w := qt.NewQWidget2()
	mw := &MainWindow{QWidget: w}

	w.SetWindowTitle("MClauncher")
	w.SetFixedSize2(winW, winH)
	// Frameless chrome matching original custom skin window.
	w.SetWindowFlags(qt.Window | qt.FramelessWindowHint)
	w.SetAttribute2(qt.WA_StyledBackground, true)

	// Background via QLabel + setPixmap (NOT paintEvent skin).
	mw.bg = qt.NewQLabel(w)
	mw.bg.SetGeometry(0, 0, winW, winH)
	mw.bg.SetScaledContents(true)
	pix := qt.NewQPixmap()
	if !pix.LoadFromData4(assets.BGMainJPG, "JPG") {
		// Fallback empty pixmap keeps window usable if asset missing.
		pix = qt.NewQPixmap2(winW, winH)
	}
	mw.bg.SetPixmap(pix)
	mw.bg.Lower()

	mw.btnLaunch = qt.NewQPushButton5("启动游戏", w)
	place(mw.btnLaunch.QWidget, geoLaunch)
	mw.btnLaunch.SetStyleSheet(styleLaunch)
	mw.btnLaunch.SetFont(uiFont(13, true))
	mw.btnLaunch.OnClicked(func() {
		qt.QMessageBox_Information(w, "启动游戏", "启动逻辑尚未实现。\n下一步将接入旧版 Mojang Yggdrasil（邮箱/密码）登录。")
	})

	mw.btnUpdate = qt.NewQPushButton5("更新游戏", w)
	place(mw.btnUpdate.QWidget, geoUpdate)
	mw.btnUpdate.SetStyleSheet(styleSmall)
	mw.btnUpdate.SetFont(uiFont(10, false))
	mw.btnUpdate.OnClicked(func() {
		qt.QMessageBox_Information(w, "更新游戏", "更新功能占位（未实现网络下载）。")
	})

	mw.btnSettings = qt.NewQPushButton5("设置", w)
	place(mw.btnSettings.QWidget, geoSettings)
	mw.btnSettings.SetStyleSheet(styleSmall)
	mw.btnSettings.SetFont(uiFont(10, false))
	mw.btnSettings.OnClicked(func() {
		OpenSettingsDialog(w)
	})

	mw.combo = qt.NewQComboBox(w)
	place(mw.combo.QWidget, geoCombo)
	mw.combo.SetStyleSheet(styleCombo)
	mw.combo.SetFont(uiFont(10, false))
	mw.combo.AddItem("(默认)")
	mw.combo.AddItem("正式版")
	mw.combo.AddItem("快照预览版")
	mw.combo.SetEditable(false)

	mw.btnOpenDir = qt.NewQPushButton5("打开目录", w)
	place(mw.btnOpenDir.QWidget, geoOpenDir)
	mw.btnOpenDir.SetStyleSheet(styleOpenDir)
	mw.btnOpenDir.SetFont(uiFont(10, false))
	mw.btnOpenDir.SetCursor(qt.NewQCursor2(qt.PointingHandCursor))
	mw.btnOpenDir.SetFlat(true)
	mw.btnOpenDir.OnClicked(func() {
		dir, err := os.Getwd()
		if err != nil {
			dir = "."
		}
		// Prefer .minecraft next to cwd if present.
		mc := filepath.Join(dir, ".minecraft")
		if st, err := os.Stat(mc); err == nil && st.IsDir() {
			dir = mc
		}
		url := qt.QUrl_FromLocalFile(dir)
		qt.QDesktopServices_OpenUrl(url)
	})

	mw.footer = qt.NewQLabel5("MineCraft Launcher 5.31 by 忘却的旋律", w)
	place(mw.footer.QWidget, geoFooter)
	mw.footer.SetStyleSheet(styleFooter)
	mw.footer.SetFont(uiFont(9, false))
	mw.footer.SetAlignment(qt.AlignLeft | qt.AlignVCenter)

	mw.btnClose = qt.NewQPushButton5("×", w)
	place(mw.btnClose.QWidget, geoClose)
	mw.btnClose.SetStyleSheet(styleClose)
	mw.btnClose.OnClicked(func() {
		w.Close()
	})

	// Raise controls above background label.
	mw.btnLaunch.Raise()
	mw.btnUpdate.Raise()
	mw.btnSettings.Raise()
	mw.combo.Raise()
	mw.btnOpenDir.Raise()
	mw.footer.Raise()
	mw.btnClose.Raise()

	mw.enableDrag(w)
	return mw
}

func place(w *qt.QWidget, r rect) {
	w.SetGeometry(r.x, r.y, r.w, r.h)
}

func (mw *MainWindow) enableDrag(w *qt.QWidget) {
	w.SetMouseTracking(true)
	w.OnMousePressEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		if e.Button() == qt.LeftButton {
			gp := e.GlobalPos()
			wp := w.Pos()
			mw.dragPos = qt.NewQPoint2(gp.X()-wp.X(), gp.Y()-wp.Y())
			mw.dragging = true
		}
		super(e)
	})
	w.OnMouseMoveEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		if mw.dragging && mw.dragPos != nil && (e.Buttons()&qt.LeftButton) != 0 {
			gp := e.GlobalPos()
			w.Move(gp.X()-mw.dragPos.X(), gp.Y()-mw.dragPos.Y())
		}
		super(e)
	})
	w.OnMouseReleaseEvent(func(super func(e *qt.QMouseEvent), e *qt.QMouseEvent) {
		mw.dragging = false
		super(e)
	})
}
