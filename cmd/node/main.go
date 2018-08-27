package main

import (
	"context"
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
	"github.com/sonm-io/core/util/metrics"
	"golang.org/x/sync/errgroup"
)

var (
	appVersion  string
	configFlag  string
	versionFlag bool
)

func main() {
	cmd.NewCmd("node", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := node.NewConfig(configFlag)
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
		server, err := node.New(ctx, cfg, node.WithLog(log))
		if err != nil {
			return fmt.Errorf("failed to build Node instance: %s", err)
		}

		return server.Serve(ctx)
	})
	wg.Go(func() error {
		return metrics.NewPrometheusExporter(cfg.MetricsListenAddr, metrics.WithLogging(log.Sugar())).Serve(ctx)
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("node termination: %s", err)
	}

	return nil
}
