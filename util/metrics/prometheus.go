package metrics

import (
	"context"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type options struct {
	log *zap.SugaredLogger
}

func newOptions() *options {
	return &options{
		log: zap.NewNop().Sugar(),
	}
}

type Option func(o *options)

func WithLogging(log *zap.SugaredLogger) Option {
	return func(o *options) {
		o.log = log
	}
}

type PrometheusExporter struct {
	addr string
	log  *zap.SugaredLogger
}

func NewPrometheusExporter(addr string, options ...Option) *PrometheusExporter {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	m := &PrometheusExporter{
		addr: addr,
		log:  opts.log,
	}

	return m
}

// Serve starts prometheus exporter HTTP server, blocking until the specified
// context is done or some critical error occurs.
func (m *PrometheusExporter) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", m.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{Handler: newHandler()}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		m.log.Infof("starting metrics server on %s", m.addr)
		defer m.log.Infof("stopped metrics server on %s", m.addr)

		return server.Serve(listener)
	})

	// Wait for either external context canceled or server fails to serve.
	<-ctx.Done()
	server.Close()

	return wg.Wait()
}

func newHandler() http.Handler {
	handler := http.NewServeMux()
	handler.Handle("/metrics", promhttp.Handler())

	return handler
}
