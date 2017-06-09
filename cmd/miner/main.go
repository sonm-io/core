package main

import (
	"os"
	"os/signal"

	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"github.com/sonm-io/insonmnia/insonmnia/miner"
)

func main() {
	ctx := context.Background()
	m, err := miner.New(ctx)
	if err != nil {
		ctxlog.GetLogger(ctx).Fatal("failed to create a new Miner", zap.Error(err))
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		m.Close()
	}()

	// TODO: check error type
	if err = m.Serve(); err != nil {
		ctxlog.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
