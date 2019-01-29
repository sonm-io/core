package main

import (
	"fmt"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/gatekeeper"
	"github.com/sonm-io/core/insonmnia/logging"
	"golang.org/x/net/context"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := gatekeeper.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Log)
	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load Ethereum keys: %s", err)
	}

	g, err := gatekeeper.NewGatekeeper(ctx, key, cfg)
	if err != nil {
		return fmt.Errorf("failed to build Gatekeeper instance: %s", err)
	}

	if err := g.Serve(ctx); err != nil {
		return fmt.Errorf("termination: %s", err)
	}

	return nil
}
