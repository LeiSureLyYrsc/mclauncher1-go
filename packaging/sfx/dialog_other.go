//go:build !windows

package main

func showErrorDialog(msg string) {
	// stderr already printed by fatal(); no GUI dialog on Unix.
	_ = msg
}
