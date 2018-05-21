// This package is responsible for server-side of NAT Punching Protocol.
// TODO: Check for reuseport available. If not - do not try to punch the NAT.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
)

type connTuple struct {
	net.Conn
	err error
}

func newConnTuple(conn net.Conn, err error) connTuple {
	return connTuple{conn, err}
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
	return m.Conn, m.err
}

// Listener specifies a net.Listener wrapper that is aware of NAT Punching
// Protocol and can switch to it when it's required to establish a connection.
//
// Options are: rendezvous server, private IPs usage, relay server(s) if any.
type Listener struct {
	ctx context.Context
	log *zap.Logger

	listener        net.Listener
	listenerChannel chan connTuple

	puncher    NATPuncher
	puncherNew func() (NATPuncher, error)
	nppChannel chan connTuple

	relayListen  func() (net.Conn, error)
	relayChannel chan connTuple

	minBackoffInterval time.Duration
	maxBackoffInterval time.Duration
}

// NewListener constructs a new NPP listener that will listen the specified
// network address with TCP protocol, switching to the NPP when there is no
// pending connections available.
func NewListener(ctx context.Context, addr string, options ...Option) (net.Listener, error) {
	opts := newOptions(ctx)

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

	m := &Listener{
		ctx:             ctx,
		log:             opts.log,
		listenerChannel: channel,
		listener:        listener,
		puncher:         opts.puncher,
		puncherNew:      opts.puncherNew,
		nppChannel:      make(chan connTuple, opts.nppBacklog),

		relayListen:  opts.relayListen,
		relayChannel: make(chan connTuple, opts.nppBacklog),

		minBackoffInterval: 500 * time.Millisecond,
		maxBackoffInterval: 8000 * time.Millisecond,
	}

	go m.listen(ctx)
	go m.listenPuncher(ctx)
	go m.listenRelay(ctx)

	return m, nil
}

func (m *Listener) listen(ctx context.Context) {
	for {
		conn, err := m.listener.Accept()
		select {
		case m.listenerChannel <- connTuple{conn, err}:
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

func (m *Listener) listenPuncher(ctx context.Context) error {
	if m.puncherNew == nil {
		return nil
	}

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
			puncher, err := m.puncherNew()
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

		m.nppChannel <- newConnTuple(m.puncher.AcceptContext(ctx))
	}
}

func (m *Listener) listenRelay(ctx context.Context) error {
	if m.relayListen == nil {
		return nil
	}

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

		conn, err := m.relayListen()
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
	// Act as a listener if there is no puncher specified.
	// Check for acceptor listenerChannel, if there is a connection - return immediately.
	select {
	case conn := <-m.listenerChannel:
		m.log.Info("received acceptor peer", zap.Any("conn", conn))
		return conn.unwrap()
	default:
	}

	// Otherwise block when either a new connection arrives or NPP does its job.
	for {
		select {
		case <-m.ctx.Done():
			return nil, m.ctx.Err()
		case conn := <-m.listenerChannel:
			m.log.Info("received acceptor peer", zap.Any("conn", conn), zap.Error(conn.err))
			return conn.unwrap()
		case conn := <-m.nppChannel:
			m.log.Info("received NPP peer", zap.Any("conn", conn), zap.Error(conn.err))
			// In case of any rendezvous errors it's better to reconnect.
			// Just in case.
			if conn.IsRendezvousError() {
				m.puncher.Close()
				m.puncher = nil
			} else {
				return conn.unwrap()
			}
		case conn := <-m.relayChannel:
			m.log.Info("received relay peer", zap.Any("conn", conn), zap.Error(conn.err))
			return conn.unwrap()
		}
	}
}

func (m *Listener) Close() error {
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
