package main

import (
	"fmt"
	"os"

	qt "github.com/mappu/miqt/qt6"

	"github.com/LeiSureLyYrsc/mclauncher1-go/internal/ui"
)

func main() {
	qt.NewQApplication(os.Args)
	qt.QCoreApplication_SetApplicationName("MClauncher")
	qt.QCoreApplication_SetOrganizationName("忘却的旋律")

	if err := ui.LoadMainQML(); err != nil {
		fmt.Fprintf(os.Stderr, "QML load failed: %v\n", err)
		os.Exit(1)
	}

	os.Exit(qt.QApplication_Exec())
}
