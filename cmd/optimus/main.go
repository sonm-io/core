package main

import (
	"context"
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/version"
	"github.com/sonm-io/core/optimus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg, err := optimus.LoadConfig(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	if cfg.Restrictions != nil {
		control, err := optimus.RestrictUsage(cfg.Restrictions)
		if err != nil {
			return err
		}
		defer control.Delete()
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(cfg.Logging.LogLevel().Zap()),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	log, err := zapConfig.Build()
	if err != nil {
		return err
	}

	ctx := ctxlog.WithLogger(context.Background(), log)
	version.ValidateVersion(ctx, version.NewLogObserver(log.Sugar()))
	bot, err := optimus.NewOptimus(cfg, optimus.WithVersion(app.Version), optimus.WithLog(log.Sugar()))
	if err != nil {
		return fmt.Errorf("failed to create Optimus: %v", err)
	}

	return bot.Run(ctx)
}
