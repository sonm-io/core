package main

import (
	"context"
	"fmt"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := dwh.NewDWHConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := log.WithLogger(context.Background(), logger)
	logger.Info("starting with config", zap.Any("config", cfg))

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %s", err)
	}

	w, err := dwh.NewDWH(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("failed to create new DWH service: %s", err)
	}

	p, err := dwh.NewL1Processor(ctx, &dwh.L1ProcessorConfig{
		Storage:    cfg.Storage,
		Blockchain: cfg.Blockchain,
		NumWorkers: cfg.NumWorkers,
		ColdStart:  cfg.ColdStart,
	})
	if err != nil {
		return fmt.Errorf("failed to create L1 events processor instance: %v", err)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		err := cmd.WaitInterrupted(ctx)
		p.Stop()
		w.Stop()
		return err
	})
	wg.Go(func() error {
		return metrics.NewPrometheusExporter(cfg.MetricsListenAddr, metrics.WithLogging(logger.Sugar())).Serve(ctx)
	})
	wg.Go(func() error {
		return debug.ServePProf(ctx, debug.Config{Port: 6060}, logger)
	})
	wg.Go(func() error {
		logger.Info("starting L1 events processor")
		defer logger.Info("stopping L1 events processor")
		return p.Start()
	})
	wg.Go(func() error {
		logger.Info("starting DWH service")
		defer logger.Info("stopping DWH service")
		return w.Serve()
	})

	return wg.Wait()
}
