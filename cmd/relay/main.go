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

func start(app cmd.AppContext) error {
	cfg, err := relay.NewServerConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)

	}
	ctx := log.WithLogger(context.Background(), logger)

	options := []relay.Option{
		relay.WithLogger(log.G(ctx)),
	}
	server, err := relay.NewServer(*cfg, options...)
	if err != nil {
		return fmt.Errorf("failed to construct a Relay server: %s", err)
	}

	// Passing the context tells about server shutdown if so.
	//
	// There are two possible cases here:
	// - User interrupts the server - then server is closed explicitly and
	//   everything inside is cleared.
	// - Server stops for some reason - then the context is notified about this
	//   forcing interruption handler to return.
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return server.Serve(ctx)
	})
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})

	if err := wg.Wait(); err != nil {
		log.S(ctx).Infof("Relay server has been stopped: %v", err)
	}

	return nil
}

func main() {
	cmd.NewCmd(start).Execute()
}
