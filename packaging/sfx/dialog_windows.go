//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func showErrorDialog(msg string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	title, err1 := syscall.UTF16PtrFromString("MClauncher")
	text, err2 := syscall.UTF16PtrFromString(
		"启动失败 / Failed to start\n\n" + msg +
			"\n\n日志请查看:\n%TEMP%\\mclauncher-sfx-error.log\n或 %LOCALAPPDATA%\\mclauncher\\mclauncher-sfx-error.log",
	)
	if err1 != nil || err2 != nil {
		return
	}
	// MB_OK | MB_ICONERROR = 0x10
	_, _, _ = proc.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}
