package tensor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/tensor"
	"github.com/sonm-io/core/util/metrics"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := tensor.NewConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := log.WithLogger(context.Background(), logger)

	log.G(ctx).Info("starting with config", zap.Any("config", cfg))

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %s", err)
	}

	w, err := tensor.NewServer(ctx, cfg, key)
	if err != nil {
		return fmt.Errorf("failed to create new Tensor service: %s", err)
	}

	go metrics.NewPrometheusExporter(cfg.MetricsListenAddr, metrics.WithLogging(logger.Sugar())).Serve(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		w.Stop()
	}()

	log.G(ctx).Info("starting Tensor service")

	wg := errgroup.Group{}
	wg.Go(w.Serve)

	return wg.Wait()
}
