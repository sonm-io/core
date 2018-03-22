package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp/relay"
)

var (
	cfgPath     string
	versionFlag bool
	appVersion  string
)

func start() {
	cfg, err := relay.NewConfig(cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config file: %s\r\n", err)
		os.Exit(1)
	}

	ctx := log.WithLogger(context.Background(), logging.BuildLogger(cfg.LogLevel()))

	options := []relay.Option{
		relay.WithLogger(log.G(ctx)),
	}
	server, err := relay.NewServer(*cfg, options...)
	if err != nil {
		log.S(ctx).Errorf("failed to construct a Relay server: %s", err)
		os.Exit(1)
	}

	go server.Serve()
	defer server.Close()

	cmd.WaitInterrupted()
}

func main() {
	cmd.NewCmd("relay", appVersion, &cfgPath, &versionFlag, start).Execute()
}
