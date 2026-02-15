//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func init() {
	// Enable ANSI escape sequence processing on Windows 10+
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return
	}
	windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)

	// Also enable for stderr
	h, err = windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if err != nil {
		return
	}
	if windows.GetConsoleMode(h, &mode) != nil {
		return
	}
	windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	_ = os.Stderr // keep import used
}
