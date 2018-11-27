package main

import (
	"context"
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
	"github.com/sonm-io/core/insonmnia/version"
	"github.com/sonm-io/core/util/metrics"
	"golang.org/x/sync/errgroup"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := node.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	log, err := logging.BuildLogger(cfg.Log)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := ctxlog.WithLogger(context.Background(), log)
	version.ValidateVersion(ctx, version.NewLogObserver(log.Sugar()))

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
