package main

import (
	"context"
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/optimus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("optimus", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	cfg, err := optimus.LoadConfig(configFlag)
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
	bot, err := optimus.NewOptimus(cfg, optimus.WithVersion(appVersion), optimus.WithLog(log.Sugar()))
	if err != nil {
		return fmt.Errorf("failed to create Optimus: %v", err)
	}

	return bot.Run(ctx)
}
