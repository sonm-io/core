package xgrpc

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/stats"
)

var (
	connectionsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "grpc_server_connections_current",
		Help: "Number of currently running tasks",
	})
)

func init() {
	prometheus.MustRegister(connectionsGauge)
}

type Handler struct{}

func (h *Handler) TagRPC(ctx context.Context, rpcTagInfo *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes the RPC stats.
func (h *Handler) HandleRPC(context.Context, stats.RPCStats) {}

func (h *Handler) TagConn(ctx context.Context, connTagInfo *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes the Conn stats.
func (h *Handler) HandleConn(ctx context.Context, connStats stats.ConnStats) {
	if connStats.IsClient() {
		return
	}

	switch connStats.(type) {
	case *stats.ConnBegin:
		connectionsGauge.Inc()
	case *stats.ConnEnd:
		connectionsGauge.Dec()
	}
}
