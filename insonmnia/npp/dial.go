// This package is responsible for Client side for NAT Punching Protocol.

package npp

import (
	"context"
	"fmt"
	"net"
	"time"

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
	log := m.log.With(zap.Stringer("remote_addr", addr))
	log.Debug("connecting to remote peer")

	if conn := m.dialDirect(ctx, addr); conn != nil {
		return conn, nil
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	timeout := 5 * time.Second
	log.Debug("connecting using NPP", zap.Duration("timeout", timeout))

	if m.puncherNew != nil {
		nppCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		nppChannel := make(chan connTuple)

		go func() {
			puncher, err := m.puncherNew(nppCtx)
			if err != nil {
				nppChannel <- newConnTuple(nil, err)
				return
			}
			defer puncher.Close()

			nppChannel <- newConnTuple(puncher.Dial(ethAddr))
		}()

		select {
		case conn := <-nppChannel:
			err := conn.Error()
			if err == nil {
				log.Debug("successfully connected using NPP", zap.Stringer("remote_peer", conn.RemoteAddr()))
				return conn.unwrap()
			}

			log.Warn("failed to connect using NPP", zap.Error(err))

			if m.relayDialer == nil {
				log.Debug("no relay configured - returning error", zap.Error(err))
				return conn.unwrap()
			}
		case <-nppCtx.Done():
			log.Warn("failed to connect using NPP", zap.Error(nppCtx.Err()))
		}
	}

	if m.relayDialer == nil {
		log.Warn("failed to connect using NPP - all methods failed")
		return nil, fmt.Errorf("failed to connect using NPP - all methods failed")
	}

	log.Debug("connecting using Relay")
	channel := make(chan connTuple)
	go func() {
		channel <- newConnTuple(m.relayDialer.Dial(ethAddr))
	}()

	select {
	case conn := <-channel:
		err := conn.Error()
		if err == nil {
			log.Debug("successfully connected using Relay", zap.Stringer("remote_peer", conn.RemoteAddr()))
		} else {
			log.Warn("failed to connect using Relay", zap.Error(err))
		}

		return conn.unwrap()
	case <-ctx.Done():
		log.Warn("failed to connect using Relay", zap.Error(ctx.Err()))
		return nil, ctx.Err()
	}
}

// Note, that this method acts as an optimization.
func (m *Dialer) dialDirect(ctx context.Context, addr auth.Addr) net.Conn {
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

	log.Debug("successfully connected using direct TCP")
	return conn
}

// Close closes the dialer.
//
// Any blocked operations will be unblocked and return errors.
func (m *Dialer) Close() error {
	return nil
}
