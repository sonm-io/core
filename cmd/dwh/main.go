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
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("dwh", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := dwh.NewConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger := logging.BuildLogger(cfg.Logging.Level)
	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %s", err)
	}

	w, err := dwh.NewDWH(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("failed to create new DWH service: %s", err)
	}

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		w.Stop()
	}()

	log.G(ctx).Info("starting DWH service", zap.String("grpc_addr", cfg.GRPCListenAddr), zap.String("http_addr", cfg.HTTPListenAddr))
	if err := w.Serve(); err != nil {
		return fmt.Errorf("failed to serve DWH: %s", err)
	}

	return nil
}
