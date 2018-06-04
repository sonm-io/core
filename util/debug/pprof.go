package debug

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"
)

const (
	prefix = "/debug/pprof"
	ipAddr = "localhost"
)

type Config struct {
	Port uint16 `yaml:"port" default:"6060"`
}

// ServePProf starts pprof HTTP server on the specified port, blocking until
// the context is done.
func ServePProf(ctx context.Context, cfg Config, log *zap.Logger) error {
	addr := fmt.Sprintf("%s:%d", ipAddr, cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	go func() error {
		log.Sugar().Infof("starting pprof server on %s", addr)
		defer log.Sugar().Infof("stopped pprof server on %s", addr)

		return http.Serve(listener, newHandler(prefix))
	}()

	<-ctx.Done()
	return ctx.Err()
}

func newHandler(prefix string) http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc(prefix+"/", pprof.Index)
	handler.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
	handler.HandleFunc(prefix+"/profile", pprof.Profile)
	handler.HandleFunc(prefix+"/symbol", pprof.Symbol)
	handler.HandleFunc(prefix+"/trace", pprof.Trace)

	return handler
}
