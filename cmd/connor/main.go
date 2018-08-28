package main

import (
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/connor"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util/metrics"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var (
	appVersion  string
	configFlag  string
	versionFlag bool
)

func main() {
	cmd.NewCmd("connor", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := connor.NewConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	log, err := logging.BuildLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := ctxlog.WithLogger(context.Background(), log)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})
	wg.Go(func() error {
		server, err := connor.New(ctx, cfg, log)
		if err != nil {
			return fmt.Errorf("failed to build Connor instance: %s", err)
		}

		return server.Serve(ctx)
	})
	wg.Go(func() error {
		return metrics.NewPrometheusExporter(cfg.Metrics, metrics.WithLogging(log.Sugar())).Serve(ctx)
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("connor termination: %s", err)
	}

	return nil
}
