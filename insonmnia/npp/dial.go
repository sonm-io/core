// This package is responsible for Client side for NAT Punching Protocol.

package npp

import (
	"context"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
)

// Dialer represents an NPP dialer.
//
// This structure acts like an usual dialer with an exception that the address
// must be an authenticated endpoint and the connection establishment process
// is done via NAT Punching Protocol.
type Dialer struct {
	ctx context.Context

	puncherNew func() (NATPuncher, error)
	relayDial  func(target common.Address) (net.Conn, error)
}

// NewDialer constructs a new dialer that is aware of NAT Punching Protocol.
func NewDialer(ctx context.Context, options ...Option) (*Dialer, error) {
	opts := newOptions(ctx)

	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	return &Dialer{
		ctx:        ctx,
		puncherNew: opts.puncherNew,
		relayDial:  opts.relayDial,
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
	if conn := m.dialDirect(ctx, addr); conn != nil {
		return conn, nil
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	nppCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	nppChannel := make(chan connTuple)

	go func() {
		puncher, err := m.puncherNew()
		if err != nil {
			nppChannel <- newConnTuple(nil, err)
			return
		}

		nppChannel <- newConnTuple(puncher.Dial(ethAddr))
	}()

	select {
	case conn := <-nppChannel:
		if conn.err == nil || m.relayDial == nil {
			return conn.unwrap()
		}
	case <-nppCtx.Done():
	}

	channel := make(chan connTuple)
	go func() {
		channel <- newConnTuple(m.relayDial(ethAddr))
	}()

	select {
	case conn := <-channel:
		return conn.unwrap()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *Dialer) dialDirect(ctx context.Context, addr auth.Addr) net.Conn {
	netAddr, err := addr.Addr()
	if err == nil {
		dialer := net.Dialer{}
		conn, err := dialer.DialContext(ctx, "tcp", netAddr)
		if err == nil {
			return conn
		} else {
			ctxlog.S(m.ctx).Warnf("failed to dial directly to %s: %s", netAddr, err)
		}
	}

	return nil
}

// Close closes the dialer.
//
// Any blocked operations will be unblocked and return errors.
func (m *Dialer) Close() error {
	return nil
}
