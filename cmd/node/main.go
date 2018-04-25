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

func run() error {
	cfg, err := node.NewConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger := logging.BuildLogger(cfg.Log.LogLevel())
	ctx := log.WithLogger(context.Background(), logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load Ethereum keys: %s", err)
	}

	n, err := node.New(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("failed to build Node instance: %s", err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		n.Close()
	}()

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	if err := n.Serve(); err != nil {
		return fmt.Errorf("node termination: %s", err)
	}

	return nil
}
