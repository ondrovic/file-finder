//go:build windows
// +build windows

package utils

import (
	"golang.org/x/sys/windows"
	"os"
)

func getTerminalSize() (int, int, error) {
	return getTerminalSizeWindows()
}

func getTerminalSizeWindows() (int, int, error) {
	handle := windows.Handle(os.Stdout.Fd())
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(handle, &info)
	if err != nil {
		return 0, 0, err
	}
	width := int(info.Window.Right - info.Window.Left + 1)
	height := int(info.Window.Bottom - info.Window.Top + 1)
	return width, height, nil
}
