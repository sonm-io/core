package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/worker"
	"github.com/sonm-io/core/util/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	ctx := context.Background()
	waiter, ctx := errgroup.WithContext(ctx)
	cfg, err := worker.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}
	ctx = log.WithLogger(ctx, logger)

	storage, err := state.NewState(ctx, &cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to create state storage: %s", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	waiter.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-c:
			log.G(ctx).Info("closing worker by interrupt signal")
		case <-ctx.Done():
		}

		cancel()

		return nil
	})

	w, err := worker.NewWorker(worker.WithConfig(cfg), worker.WithContext(ctx), worker.WithStateStorage(storage),
		worker.WithVersion(app.Version))
	if err != nil {
		return fmt.Errorf("failed to create Worker instance: %s", err)
	}

	go metrics.NewPrometheusExporter(cfg.MetricsListenAddr, metrics.WithLogging(logger.Sugar())).Serve(ctx)

	if err = w.Serve(); err != nil {
		cancel()
		log.G(ctx).Error("server stop", zap.Error(err))
	}

	return waiter.Wait()
}
