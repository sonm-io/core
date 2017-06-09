package main

import (
	"os"
	"os/signal"

	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/insonmnia/insonmnia/hub"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	h, err := hub.New(ctx)
	if err != nil {
		ctxlog.GetLogger(ctx).Fatal("failed to create a new Hub", zap.Error(err))
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		h.Close()
	}()

	// TODO: check error type
	if err = h.Serve(); err != nil {
		ctxlog.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
