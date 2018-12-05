package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"

	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/oracle"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := oracle.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("failed to build logger insance: %s", err)
	}

	// ctx := log.WithLogger(context.Background(), logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	o, err := oracle.NewOracle(ctx, logger, cfg)
	if err != nil {
		return fmt.Errorf("failed to build Oracle instance: %s", err)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})

	wg.Go(func() error {
		return o.Serve(ctx)
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("termination: %s", err)
	}

	return nil
}
