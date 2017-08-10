package main

import (
	"fmt"

	"github.com/sonm-io/core/cmd/cli/commands"
	"github.com/sonm-io/core/cmd/cli/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Cannot load config: %s\r\n", err)
		return
	}

	commands.Root(cfg).Execute()
}
