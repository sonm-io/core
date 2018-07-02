package main

import (
	"context"
	"fmt"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp/relay"
	"golang.org/x/sync/errgroup"
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

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return server.Serve(ctx)
	})

	// Passing the context tells about server shutdown if so.
	//
	// There are two possible cases here:
	// - User interrupts the server - then server is closed explicitly and
	// 	 everything inside is cleared.
	// - Server stops for some reason - then the context is notified about this
	//	 forcing interruption handler to return.
	cmd.WaitInterrupted(ctx)

	server.Close()
	wg.Wait()

	return nil
}

func main() {
	cmd.NewCmd("relay", appVersion, &cfgPath, &versionFlag, start).Execute()
}
