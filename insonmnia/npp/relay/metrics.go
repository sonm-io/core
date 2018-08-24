package relay

import (
	"sync"
	"time"

	"github.com/sonm-io/core/insonmnia/npp/nppc"
	"github.com/sonm-io/core/proto"
	"go.uber.org/atomic"
)

type metrics struct {
	ConnCurrent *atomic.Uint64

	mu          sync.Mutex
	net         map[nppc.ResourceID]*netMetrics
	meetingRoom *meetingRoom

	birthTime time.Time
}

func newMetrics(meetingRoom *meetingRoom) *metrics {
	return &metrics{
		ConnCurrent: atomic.NewUint64(0),
		net:         map[nppc.ResourceID]*netMetrics{},
		meetingRoom: meetingRoom,
		birthTime:   time.Now(),
	}
}

func (m *metrics) NetMetrics(addr nppc.ResourceID) *netMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics, ok := m.net[addr]
	if !ok {
		metrics = newNetMetrics()
		m.net[addr] = metrics
	}

	return metrics
}

func (m *metrics) Dump() *sonm.RelayMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	netMetrics := map[string]*sonm.NetMetrics{}

	for addr, metrics := range m.net {
		timestamp := &sonm.Timestamp{}
		if ts, err := m.meetingRoom.ServerActiveTime(addr); err == nil {
			timestamp.Seconds = ts.Unix()
		}
		netMetrics[addr.String()] = &sonm.NetMetrics{
			TxBytes:    metrics.TxBytes.Load(),
			RxBytes:    metrics.RxBytes.Load(),
			TimeActive: timestamp,
		}
	}

	return &sonm.RelayMetrics{
		ConnCurrent: m.ConnCurrent.Load(),
		Net:         netMetrics,
		Uptime:      uint64(time.Since(m.birthTime).Seconds()),
	}
}

type netMetrics struct {
	// TxBytes shows the number of bytes sent from a server.
	TxBytes *atomic.Uint64
	// RxBytes shows the number of bytes received by a server.
	RxBytes *atomic.Uint64
}

func newNetMetrics() *netMetrics {
	return &netMetrics{
		TxBytes: atomic.NewUint64(0),
		RxBytes: atomic.NewUint64(0),
	}
}
