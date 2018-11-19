// This package is responsible for server-side of NAT Punching Protocol.
// TODO: Check for reuseport available. If not - do not try to punch the NAT.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sonm-io/core/insonmnia/npp/relay"
	"go.uber.org/zap"
)

type connTuple struct {
	conn net.Conn
	err  error
}

func newConnTuple(conn net.Conn, err error) connTuple {
	return connTuple{conn, err}
}

func (m *connTuple) RemoteAddr() net.Addr {
	if m == nil || m.conn == nil {
		return nil
	}
	return m.conn.RemoteAddr()
}

func (m *connTuple) Close() error {
	if m == nil || m.conn == nil {
		return nil
	}
	return m.conn.Close()
}

func (m *connTuple) Error() error {
	return m.err
}

func (m *connTuple) IsRendezvousError() bool {
	if m.err == nil {
		return false
	}

	_, ok := m.err.(*rendezvousError)
	return ok
}

func (m *connTuple) unwrap() (net.Conn, error) {
	return m.conn, m.err
}

func (m *connTuple) unwrapWithSource(source connSource) (net.Conn, connSource, error) {
	return m.conn, source, m.err
}

// Listener specifies a net.Listener wrapper that is aware of NAT Punching
// Protocol and can switch to it when it's required to establish a connection.
//
// Options are: rendezvous server, private IPs usage, relay server(s) if any.
type Listener struct {
	metrics *metrics
	ctx     context.Context // Required here, because of gRPC server, which can't stop properly even if "Stop" called.
	cancel  context.CancelFunc
	log     *zap.Logger

	listener        net.Listener
	listenerChannel chan connTuple

	puncher    NATPuncher
	puncherNew func(ctx context.Context) (NATPuncher, error)
	nppChannel chan connTuple

	puncherQUIC    NATPuncher
	puncherNewQUIC func(ctx context.Context) (NATPuncher, error)

	relayListener *relay.Listener
	relayChannel  chan connTuple

	minBackoffInterval time.Duration
	maxBackoffInterval time.Duration
}

// NewListener constructs a new NPP listener that will listen the specified
// network address with TCP protocol, switching to the NPP when there is no
// pending connections available.
func NewListener(ctx context.Context, addr string, options ...Option) (*Listener, error) {
	opts := newOptions()

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	channel := make(chan connTuple, 1)

	listener, err := net.Listen(protocol, addr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	m := &Listener{
		metrics:         newMetrics(),
		ctx:             ctx,
		cancel:          cancel,
		log:             opts.log,
		listenerChannel: channel,
		listener:        listener,
		puncher:         nil,
		puncherNew:      opts.puncherNew,
		nppChannel:      make(chan connTuple, opts.nppBacklog),

		puncherQUIC:    nil,
		puncherNewQUIC: opts.puncherNewQUIC,

		relayListener: opts.relayListener,
		relayChannel:  make(chan connTuple, opts.nppBacklog),

		minBackoffInterval: opts.nppMinBackoffInterval,
		maxBackoffInterval: opts.nppMaxBackoffInterval,
	}

	go m.listen(ctx)
	go m.listenQUIC(ctx)
	go m.listenPuncher(ctx)
	go m.listenRelay(ctx)

	return m, nil
}

func (m *Listener) listen(ctx context.Context) {
	for {
		conn, err := m.listener.Accept()
		select {
		case m.listenerChannel <- newConnTuple(conn, err):
		case <-ctx.Done():
			m.log.Info("finished listening due to cancellation", zap.Error(ctx.Err()))
			return
		}
		if err != nil {
			m.log.Info("finished listening on Accept error", zap.Error(err))
			return
		}
	}
}

func (m *Listener) listenQUIC(ctx context.Context) error {
	if m.puncherNewQUIC == nil {
		return nil
	}

	defer m.log.Info("finished listening QUIC NPP")

	timeout := m.minBackoffInterval

	for {
		timer := time.NewTimer(timeout)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Okay, let's go.
		}

		if m.puncherQUIC == nil {
			m.log.Debug("constructing new QUIC puncher")
			puncher, err := m.puncherNewQUIC(ctx)
			if err != nil {
				m.log.Warn("failed to construct a QUIC puncher", zap.Error(err))
				if timeout < m.maxBackoffInterval {
					timeout = 2 * timeout
				}
				continue
			}

			m.log.Debug("QUIC puncher has been constructed", zap.Stringer("remote", puncher.RemoteAddr()))
			m.puncherQUIC = puncher

			timeout = m.minBackoffInterval
		}

		connTuple := newConnTuple(m.puncherQUIC.AcceptContext(ctx))
		if connTuple.IsRendezvousError() {
			// In case of any rendezvous errors it's better to reconnect.
			// Just in case.
			// todo: reconnect only if error is on network level.
			m.puncherQUIC.Close()
			m.puncherQUIC = nil
		}

		m.nppChannel <- connTuple
	}
}

func (m *Listener) listenPuncher(ctx context.Context) error {
	if m.puncherNew == nil {
		return nil
	}

	defer m.log.Info("finished listening NPP")

	timeout := m.minBackoffInterval
	for {
		timer := time.NewTimer(timeout)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Okay, let's go.
		}

		if m.puncher == nil {
			m.log.Debug("constructing new puncher")
			puncher, err := m.puncherNew(ctx)
			if err != nil {
				m.log.Warn("failed to construct a puncher", zap.Error(err))
				if timeout < m.maxBackoffInterval {
					timeout = 2 * timeout
				}
				continue
			}

			m.log.Debug("puncher has been constructed", zap.Stringer("remote", puncher.RemoteAddr()))
			m.puncher = puncher

			timeout = m.minBackoffInterval
		}

		connTuple := newConnTuple(m.puncher.AcceptContext(ctx))
		if connTuple.IsRendezvousError() {
			// In case of any rendezvous errors it's better to reconnect.
			// Just in case.
			m.puncher.Close()
			m.puncher = nil
		}

		m.nppChannel <- connTuple
	}
}

func (m *Listener) listenRelay(ctx context.Context) error {
	if m.relayListener == nil {
		return nil
	}

	defer m.log.Info("finished listening Relay")

	timeout := m.minBackoffInterval

	for {
		timer := time.NewTimer(timeout)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		m.log.Debug("connecting using relay")

		conn, err := m.relayListener.Accept()
		if err != nil {
			m.log.Warn("failed to relay", zap.Error(err))
			if timeout < m.maxBackoffInterval {
				timeout = 2 * timeout
			}
		} else {
			timeout = m.minBackoffInterval
		}

		m.relayChannel <- newConnTuple(conn, newRelayError(err))
	}
}

// Accept waits for and returns the next connection to the listener.
//
// This method will firstly check whether there are pending sockets in the
// listener, returning immediately if so.
// Then an attempt to communicate with the Rendezvous server occurs by
// publishing server's ID to check if there are someone wanted to connect with
// us.
// Simultaneously additional sockets are constructed after resolution to make
// punching mechanism work. This can consume a meaningful amount of file
// descriptors, so be prepared to enlarge your limits.
func (m *Listener) Accept() (net.Conn, error) {
	return m.AcceptContext(m.ctx)
}

func (m *Listener) AcceptContext(ctx context.Context) (net.Conn, error) {
	conn, source, err := m.accept(ctx)
	if err != nil {
		m.log.Warn("failed to accept peer", zap.Error(err))
		return nil, err
	}

	m.log.Info("accepted peer", zap.Stringer("source", source), zap.Stringer("remote", conn.RemoteAddr()))

	switch source {
	case sourceDirectConnection:
		m.metrics.NumConnectionsDirect.Inc()
	case sourceNPPConnection:
		m.metrics.NumConnectionsNAT.Inc()
	case sourceRelayedConnection:
		m.metrics.NumConnectionsRelay.Inc()
	default:
		return nil, fmt.Errorf("unknown connection source")
	}

	return conn, nil
}

// Note: this function only listens for multiple channels and transforms the
// result from a single-value to multiple values, due to weird Go type system.
func (m *Listener) accept(ctx context.Context) (net.Conn, connSource, error) {
	// Act as a listener if there is no puncher specified.
	// Check for acceptor listenerChannel, if there is a connection - return immediately.
	select {
	case conn := <-m.listenerChannel:
		return conn.unwrapWithSource(sourceDirectConnection)
	default:
	}

	// Otherwise block when either a new connection arrives or NPP does its job.
	for {
		select {
		case <-ctx.Done():
			return nil, sourceError, ctx.Err()
		case conn := <-m.listenerChannel:
			return conn.unwrapWithSource(sourceDirectConnection)
		case conn := <-m.nppChannel:
			if !conn.IsRendezvousError() {
				return conn.unwrapWithSource(sourceNPPConnection)
			}
		case conn := <-m.relayChannel:
			return conn.unwrapWithSource(sourceRelayedConnection)
		}
	}
}

func (m *Listener) Close() error {
	m.cancel()

	var errs []error

	if err := m.listener.Close(); err != nil {
		errs = append(errs, err)
	}
	if m.puncher != nil {
		if err := m.puncher.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close listener: %+v", errs)
	} else {
		return nil
	}
}

func (m *Listener) Addr() net.Addr {
	return m.listener.Addr()
}

func (m *Listener) Metrics() ListenerMetrics {
	var rendezvousAddr net.Addr
	if m.puncher != nil {
		rendezvousAddr = m.puncher.RemoteAddr()
	}

	return ListenerMetrics{
		RendezvousAddr:       rendezvousAddr,
		NumConnectionsDirect: m.metrics.NumConnectionsDirect.Load(),
		NumConnectionsNAT:    m.metrics.NumConnectionsNAT.Load(),
		NumConnectionsRelay:  m.metrics.NumConnectionsRelay.Load(),
	}
}
