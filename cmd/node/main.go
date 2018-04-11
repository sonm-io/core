package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("node", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() {
	cfg, err := node.NewConfig(configFlag)
	if err != nil {
		fmt.Printf("cannot load config file: %s\r\n", err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Log.LogLevel())
	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.G(ctx).Error("cannot load Ethereum keys", zap.Error(err))
		os.Exit(1)
	}

	n, err := node.New(ctx, cfg, key)
	if err != nil {
		log.G(ctx).Error("cannot build node instance", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		n.Close()
	}()

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	if err := n.Serve(); err != nil {
		log.G(ctx).Error("node termination", zap.Error(err))
		os.Exit(1)
	}
}
