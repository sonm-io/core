package main

import (
	"fmt"
	"os"

	"github.com/sonm-io/core/cmd/cli/commands"
	"github.com/sonm-io/core/cmd/cli/config"
)

var appVersion string

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Cannot load config: %s\r\n", err)
		return
	}

	cmd := commands.Root(appVersion, cfg)
	if err := cmd.Execute(); err != nil {
		commands.ShowError(cmd, err.Error(), nil)
		os.Exit(1)
	}
}
