package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/worker"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("worker", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	ctx := context.Background()
	waiter, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	cfg, err := worker.NewConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger := logging.BuildLogger(cfg.Logging.LogLevel())
	ctx = log.WithLogger(ctx, logger)

	storage, err := state.NewState(ctx, &cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to create state storage: %s", err)
	}

	waiter.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		log.G(ctx).Info("closing worker by interrupt signal")
		cancel()

		return nil
	})

	w, err := worker.NewWorker(worker.WithConfig(cfg), worker.WithContext(ctx), worker.WithStateStorage(storage),
		worker.WithVersion(appVersion))
	if err != nil {
		return fmt.Errorf("failed to create Worker instance: %s", err)
	}

	//TODO: fixme dangling goroutine
	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	if err = w.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}
	waiter.Wait()

	return nil
}
