package qmlres

import "embed"

// QML sources for the main window and settings dialog.
//
//go:embed Main.qml SettingsDialog.qml
var Files embed.FS
