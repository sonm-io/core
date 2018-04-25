package main

import (
	"context"
	"fmt"

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

func start() error {
	cfg, err := relay.NewServerConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	ctx := log.WithLogger(context.Background(), logging.BuildLogger(cfg.Logging.LogLevel()))

	options := []relay.Option{
		relay.WithLogger(log.G(ctx)),
	}
	server, err := relay.NewServer(*cfg, options...)
	if err != nil {
		return fmt.Errorf("failed to construct a Relay server: %s", err)
	}

	go server.Serve()
	defer server.Close()

	cmd.WaitInterrupted()
	return nil
}

func main() {
	cmd.NewCmd("relay", appVersion, &cfgPath, &versionFlag, start).Execute()
}
