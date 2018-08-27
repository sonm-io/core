package main

import (
	"context"
	"fmt"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/oracle"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("oracle", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := oracle.NewConfig(configFlag)
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

	o, err := oracle.NewOracle(ctx, key, cfg)
	if err != nil {
		return fmt.Errorf("failed to build Oracle instance: %s", err)
	}

	if err := o.Serve(ctx); err != nil {
		return fmt.Errorf("termination: %s", err)
	}

	return nil
}
