package main

import (
	"context"

	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/optimus"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("optimus", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := optimus.LoadConfig(configFlag)
	if err != nil {
		return err
	}

	ctx := context.Background()
	bot, err := optimus.NewOptimus(*cfg, logging.BuildLogger(*cfg.Logging.Level))
	if err != nil {
		return err
	}

	return bot.Run(ctx)
}
