package main

import (
	"context"
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/ogier/pflag"
	"go.uber.org/zap"

	"github.com/sonm-io/core/insonmnia/locator"
)

var (
	configPath  = pflag.String("config", "locator.yaml", "Path to locator config file")
	showVersion = pflag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	pflag.Parse()

	if *showVersion {
		fmt.Printf("SONM App %s\r\n", version)
		return
	}

	logger := ctxlog.GetLogger(context.Background())

	cfg, err := locator.NewConfig(*configPath)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		logger.Fatal("failed load private key", zap.Error(err))
	}

	app := locator.NewApp(cfg, key)
	if err := app.Init(); err != nil {
		logger.Fatal("cannot start App service", zap.Error(err))
	}

	logger.Info("starting App service", zap.String("bind_addr", cfg.ListenAddr))
	if err := app.Serve(); err != nil {
		logger.Fatal("cannot start App service", zap.Error(err))
	}
}
