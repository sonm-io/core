package relay

import (
	"go.uber.org/atomic"
)

type metrics struct {
	ConnCurrent *atomic.Uint64
}

func newMetrics() *metrics {
	return &metrics{
		ConnCurrent: atomic.NewUint64(0),
	}
}
