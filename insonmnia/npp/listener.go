// This package is responsible for server-side of NAT Punching Protocol.
// TODO: Check for reuseport available. If not - do not try to punch the NAT.

package npp

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
)

type connTuple struct {
	net.Conn
	error
}

func newConnTuple(conn net.Conn, err error) connTuple {
	return connTuple{conn, err}
}

func (m *connTuple) IsTransportError() bool {
	if m.error == nil {
		return false
	}

	_, ok := m.error.(TransportError)
	return ok
}

func (m *connTuple) unwrap() (net.Conn, error) {
	return m.Conn, m.error
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
	}

	go m.listen()

	return m, nil
}

func (m *Listener) listen() error {
	defer m.log.Info("finished listening")
	for {
		conn, err := m.listener.Accept()
		m.listenerChannel <- connTuple{conn, err}
		if err != nil {
			return err
		}
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
	if m.puncherNew == nil {
		conn := <-m.listenerChannel
		return conn.unwrap()
	}

	if m.puncher == nil {
		m.log.Debug("constructing new puncher")
		puncher, err := m.puncherNew()
		if err != nil {
			m.log.Error("failed to construct a puncher", zap.Error(err))
			return nil, TransportError{err}
		}

		m.log.Debug("puncher has been constructed", zap.Stringer("remote", puncher.RemoteAddr()))
		m.puncher = puncher
	}

	// Check for acceptor listenerChannel, if there is a connection - return immediately.
	select {
	case conn := <-m.listenerChannel:
		m.log.Info("received acceptor peer", zap.Any("conn", conn))
		return conn.unwrap()
	default:
	}

	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	// Otherwise block when either a new connection arrives or NPP does its job.
	nppChannel := make(chan connTuple, 1)

	go func() {
		nppChannel <- newConnTuple(m.puncher.AcceptContext(ctx))
	}()

	select {
	case conn := <-m.listenerChannel:
		m.log.Info("received acceptor peer", zap.Any("conn", conn))
		return conn.unwrap()
	case conn := <-nppChannel:
		m.log.Info("received NPP peer", zap.Any("conn", conn))
		if conn.IsTransportError() {
			m.puncher.Close()
			m.puncher = nil
		}
		return conn.unwrap()
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
