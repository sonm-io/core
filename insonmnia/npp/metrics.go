package npp

import (
	"net"

	"go.uber.org/atomic"
)

type ListenerMetrics struct {
	RendezvousAddr       net.Addr
	NumConnectionsDirect uint64
	NumConnectionsNAT    uint64
	NumConnectionsRelay  uint64
}

type metrics struct {
	NumConnectionsDirect *atomic.Uint64
	NumConnectionsNAT    *atomic.Uint64
	NumConnectionsRelay  *atomic.Uint64
}

func newMetrics() *metrics {
	return &metrics{
		NumConnectionsDirect: atomic.NewUint64(0),
		NumConnectionsNAT:    atomic.NewUint64(0),
		NumConnectionsRelay:  atomic.NewUint64(0),
	}
}
