package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/logging"
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

	log.G(ctx).Info("starting with config", zap.Any("config", cfg))

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

	go metrics.NewPrometheusExporter(cfg.MetricsListenAddr, metrics.WithLogging(logger.Sugar())).Serve(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		p.Stop()
		w.Stop()
	}()

	log.G(ctx).Info("starting DWH service")
	log.G(ctx).Info("starting L1 events processor")

	wg := errgroup.Group{}
	wg.Go(p.Start)
	wg.Go(w.Serve)

	return wg.Wait()
}
