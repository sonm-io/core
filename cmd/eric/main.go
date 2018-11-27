package main

import (
	"context"
	"fmt"
	"github.com/sonm-io/core/insonmnia/eric"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := eric.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("failed to build logger insance: %s", err)
	}

	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load Ethereum keys: %s", err)
	}

	e, err := eric.NewEric(ctx, key, cfg)
	if err != nil {
		return fmt.Errorf("failed to build Oracle instance: %s", err)
	}

	if err := e.Start(ctx); err != nil {
		return fmt.Errorf("termination: %s", err)
	}

	return nil
}
