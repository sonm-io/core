// This package is responsible for Client side for NAT Punching Protocol.

package npp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	prometheusIO "github.com/prometheus/client_model/go"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp/relay"
	"go.uber.org/zap"
)

// Dialer represents an NPP dialer.
//
// This structure acts like an usual dialer with an exception that the address
// must be an authenticated endpoint and the connection establishment process
// is done via NAT Punching Protocol.
type Dialer struct {
	log *zap.Logger

	puncherNew  func(ctx context.Context) (NATPuncher, error)
	relayDialer *relay.Dialer

	mu      sync.Mutex
	metrics map[string]*dialMetrics
}

// NewDialer constructs a new dialer that is aware of NAT Punching Protocol.
func NewDialer(options ...Option) (*Dialer, error) {
	opts := newOptions()

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	return &Dialer{
		log:         opts.log,
		puncherNew:  opts.puncherNew,
		relayDialer: opts.relayDialer,
		metrics:     map[string]*dialMetrics{},
	}, nil
}

// Dial dials the given verified address using NPP.
func (m *Dialer) Dial(addr auth.Addr) (net.Conn, error) {
	return m.DialContext(context.Background(), addr)
}

// DialContext connects to the given verified address using NPP and the
// provided context.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
func (m *Dialer) DialContext(ctx context.Context, addr auth.Addr) (net.Conn, error) {
	now := time.Now()
	metric := m.metricHandle(addr)
	metric.NumAttempts.Inc()
	metric.LastTimeActive.SetToCurrentTime()

	log := m.log.With(zap.Stringer("remote_addr", addr))

	conn, err := m.dialContextExt(ctx, addr, metric)
	if err != nil {
		log.Warn("failed to connect using NPP - all methods failed")
		metric.NumFailed.Inc()
		return nil, err
	}

	log = log.With(zap.Stringer("remote_peer", conn.RemoteAddr()))

	switch conn.Source {
	case sourceDirectConnection:
		metric.UsingTCPDirectHistogram.Observe(conn.Duration.Seconds())
		log.Debug("successfully connected using direct TCP")
	case sourceNPPConnection:
		metric.UsingNATHistogram.Observe(conn.Duration.Seconds())
		log.Debug("successfully connected using NPP")
	case sourceRelayedConnection:
		metric.UsingRelayHistogram.Observe(conn.Duration.Seconds())
		log.Debug("successfully connected using Relay")
	}

	metric.NumSuccess.Inc()
	metric.SummaryHistogram.Observe(time.Since(now).Seconds())
	metric.LastTimeSuccess.SetToCurrentTime()

	return conn, nil
}

func (m *Dialer) dialContextExt(ctx context.Context, addr auth.Addr, metric *dialMetrics) (*nppConn, error) {
	m.log.Debug("connecting to remote peer", zap.Stringer("remote_addr", addr))

	if conn := m.dialDirect(ctx, addr); conn != nil {
		return conn, nil
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	if conn := m.dialNPP(ctx, ethAddr); conn != nil {
		return conn, nil
	}

	return m.dialRelayed(ctx, ethAddr)
}

// Note, that this method acts as an optimization.
func (m *Dialer) dialDirect(ctx context.Context, addr auth.Addr) *nppConn {
	now := time.Now()

	log := m.log.With(zap.Stringer("remote_addr", addr))
	log.Debug("connecting using direct TCP")

	netAddr, err := addr.Addr()
	if err != nil {
		log.Debug("failed to connect using direct TCP", zap.Error(err))
		return nil
	}

	dial := net.Dialer{}
	conn, err := dial.DialContext(ctx, "tcp", netAddr)
	if err != nil {
		log.Debug("failed to connect using direct TCP", zap.Error(err))
		return nil
	}

	return newDirectNPPConn(conn, time.Since(now))
}

func (m *Dialer) dialNPP(ctx context.Context, addr common.Address) *nppConn {
	if m.puncherNew == nil {
		return nil
	}

	now := time.Now()

	timeout := 5 * time.Second
	log := m.log.With(zap.Stringer("remote_addr", addr))
	log.Debug("connecting using NPP", zap.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	nppChannel := make(chan connTuple)

	go func() {
		defer close(nppChannel)

		puncher, err := m.puncherNew(ctx)
		if err != nil {
			nppChannel <- newConnTuple(nil, err)
			return
		}
		defer puncher.Close()

		nppChannel <- newConnTuple(puncher.Dial(addr))
	}()

	select {
	case conn := <-nppChannel:
		err := conn.Error()
		if err == nil {
			return newPunchedNPPConn(conn.conn, time.Since(now))
		}

		log.Warn("failed to connect using NPP", zap.Error(err))
	case <-ctx.Done():
		go drainConnChannel(nppChannel)
		log.Warn("failed to connect using NPP", zap.Error(ctx.Err()))
	}

	return nil
}

func (m *Dialer) dialRelayed(ctx context.Context, addr common.Address) (*nppConn, error) {
	if m.relayDialer == nil {
		return nil, fmt.Errorf("failed to connect using NPP: no Relay configured")
	}

	now := time.Now()

	log := m.log.With(zap.Stringer("remote_addr", addr))
	log.Debug("connecting using Relay")

	channel := make(chan connTuple)
	go func() {
		defer close(channel)

		channel <- newConnTuple(m.relayDialer.Dial(addr))
	}()

	select {
	case conn := <-channel:
		if err := conn.Error(); err != nil {
			log.Warn("failed to connect using Relay", zap.Error(err))
			return nil, err
		}

		return newRelayedNPPConn(conn.conn, time.Since(now)), nil
	case <-ctx.Done():
		log.Warn("failed to connect using Relay", zap.Error(ctx.Err()))
		go drainConnChannel(channel)
		return nil, ctx.Err()
	}
}

func (m *Dialer) Metrics() (map[string][]*NamedMetric, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summary := map[string][]*NamedMetric{}
	for addr, metric := range m.metrics {
		metricNames := metric.MetricNames()
		metricsToCollect := [...]prometheus.Metric{
			metric.NumAttempts,
			metric.NumSuccess,
			metric.NumFailed,
			metric.UsingTCPDirectHistogram,
			metric.UsingNATHistogram,
			metric.UsingRelayHistogram,
			metric.SummaryHistogram,
			metric.LastTimeActive,
			metric.LastTimeSuccess,
		}
		metricsCollected := make([]*NamedMetric, 0, len(metricsToCollect))

		for id, metricToCollect := range metricsToCollect {
			value := &prometheusIO.Metric{}
			if err := metricToCollect.Write(value); err != nil {
				return nil, err
			}
			metricsCollected = append(metricsCollected, &NamedMetric{Name: metricNames[id], Metric: value})
		}

		summary[addr] = metricsCollected
	}

	return summary, nil
}

// Close closes the dialer.
//
// Any blocked operations will be unblocked and return errors.
func (m *Dialer) Close() error {
	return nil
}

func (m *Dialer) metricHandle(addr auth.Addr) *dialMetrics {
	m.mu.Lock()
	defer m.mu.Unlock()

	k := addr.String()
	if metric, ok := m.metrics[k]; ok {
		return metric
	}

	metric := newDialMetrics()
	m.metrics[k] = metric
	return metric
}

type nppConn struct {
	net.Conn
	Source   connSource
	Duration time.Duration
}

func newDirectNPPConn(conn net.Conn, duration time.Duration) *nppConn {
	return &nppConn{
		Conn:     conn,
		Source:   sourceDirectConnection,
		Duration: duration,
	}
}

func newPunchedNPPConn(conn net.Conn, duration time.Duration) *nppConn {
	return &nppConn{
		Conn:     conn,
		Source:   sourceNPPConnection,
		Duration: duration,
	}
}

func newRelayedNPPConn(conn net.Conn, duration time.Duration) *nppConn {
	return &nppConn{
		Conn:     conn,
		Source:   sourceRelayedConnection,
		Duration: duration,
	}
}

func drainConnChannel(channel <-chan connTuple) {
	for conn := range channel {
		conn.Close()
	}
}
