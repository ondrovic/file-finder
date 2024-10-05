//go:build !windows
// +build !windows

package utils

import (
	"fmt"
	"golang.org/x/term"
	"os"
)

func getTerminalSize() (int, int, error) {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return 0, 0, fmt.Errorf("not a terminal")
	}
	return term.GetSize(int(os.Stdin.Fd()))
}
