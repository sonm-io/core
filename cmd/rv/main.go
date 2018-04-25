package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/util"
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

	ctx := log.WithLogger(context.Background(), logging.BuildLogger(cfg.Logging.LogLevel()))

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

	go server.Run()
	defer server.Stop()

	waitInterrupted()
	return nil
}

func waitInterrupted() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func main() {
	cmd.NewCmd("rendezvous", appVersion, &cfgPath, &versionFlag, start).Execute()
}
