package main

import (
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"

	log "github.com/noxiouz/zapctx/ctxlog"
)

var (
	configPath = flag.String("config", "miner.yaml", "Path to miner config file")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg, err := miner.NewConfig(*configPath)
	if err != nil {
		ctxlog.GetLogger(ctx).Error("Cannot load config", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Logging().Level, common.DevelopmentMode)
	ctx = log.WithLogger(ctx, logger)

	builder := miner.MinerBuilder{}
	m, err := builder.Build()
	if err != nil {
		ctxlog.GetLogger(ctx).Fatal("failed to create a new Miner", zap.Error(err))
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		m.Close()
	}()

	// TODO: check error type
	if err = m.Serve(); err != nil {
		ctxlog.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
