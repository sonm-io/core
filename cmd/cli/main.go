package main

import (
	"os"

	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/cmd/cli/commands"
)

func main() {
	root := commands.Root(cmd.AppVersion)
	if err := root.Execute(); err != nil {
		commands.ShowError(root, err.Error(), nil)
		os.Exit(1)
	}
}
