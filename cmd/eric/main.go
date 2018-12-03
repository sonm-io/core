package main

import (
	"context"
	"fmt"
	"github.com/sonm-io/core/insonmnia/eric"
	"golang.org/x/sync/errgroup"

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

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})
	wg.Go(func() error {
		e, err := eric.NewEric(ctx, cfg)
		if err != nil {
			return fmt.Errorf("failed to build Eric instance: %s", err)
		}

		err = e.Start(ctx)
		if err != nil {
			return fmt.Errorf("eric execution failed: %s", err)
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("termination: %s", err)
	}

	return nil
}
