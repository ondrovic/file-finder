package main

import (
	"file-finder/cli/cmd"
	"runtime"

	sharedCli "github.com/ondrovic/common/utils/cli"
)

func main() {
	sharedCli.ClearTerminalScreen(runtime.GOOS)
	cmd.Execute()
}
