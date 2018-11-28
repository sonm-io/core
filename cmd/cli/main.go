package main

import (
	"os"

	"github.com/sonm-io/core/cmd/cli/commands"
	"github.com/sonm-io/core/insonmnia/version"
)

func main() {
	root := commands.Root(version.Version)
	if err := root.Execute(); err != nil {
		commands.ShowError(root, err.Error(), nil)
		os.Exit(1)
	}
}
