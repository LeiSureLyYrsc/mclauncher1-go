package main

import (
	"os"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/ui"
)

func main() {
	qt.NewQApplication(os.Args)
	qt.QCoreApplication_SetApplicationName("MClauncher")
	qt.QCoreApplication_SetOrganizationName("忘却的旋律")

	ui.ApplyAppFont(nil)

	win := ui.NewMainWindow()
	win.Show()

	os.Exit(qt.QApplication_Exec())
}
