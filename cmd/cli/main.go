package main

import (
	"os"

	"github.com/sonm-io/core/cmd/cli/commands"
)

var appVersion string

func main() {
	cmd := commands.Root(appVersion)
	if err := cmd.Execute(); err != nil {
		commands.ShowError(cmd, err.Error(), nil)
		os.Exit(1)
	}
}
