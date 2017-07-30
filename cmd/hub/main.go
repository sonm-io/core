package main

import (
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"github.com/sonm-io/core/insonmnia/hub"
)

var (
	configPath = flag.String("config", "hub.yaml", "Path to hub config file")
)

func main() {
	flag.Parse()

	conf, err := hub.NewConfig(*configPath)
	if err != nil {
		panic("Cannot load config file: " + err.Error())
	}

	ctx := context.Background()
	h, err := hub.New(ctx, conf)
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
