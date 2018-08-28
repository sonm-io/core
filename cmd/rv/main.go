package main

import (
	"context"
	"fmt"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/util"
	"golang.org/x/sync/errgroup"
)

var (
	cfgPath     string
	versionFlag bool
	appVersion  string
)

func start() error {
	cfg, err := rendezvous.NewServerConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := log.WithLogger(context.Background(), logger)

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, cfg.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate rotator: %s", err)
	}
	defer certRotator.Close()

	credentials := util.NewTLS(TLSConfig)

	options := []rendezvous.Option{
		rendezvous.WithLogger(log.G(ctx)),
		rendezvous.WithCredentials(credentials),
	}
	server, err := rendezvous.NewServer(*cfg, options...)
	if err != nil {
		return fmt.Errorf("failed to create Rendezvous server: %s", err)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return server.Run(ctx)
	})
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})

	if err := wg.Wait(); err != nil {
		log.S(ctx).Infof("rendezvous server is stopped: %v", err)
	}

	return nil
}

func main() {
	cmd.NewCmd("rendezvous", appVersion, &cfgPath, &versionFlag, start).Execute()
}
