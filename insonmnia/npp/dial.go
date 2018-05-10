// This package is responsible for Client side for NAT Punching Protocol.

package npp

import (
	"context"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	if conn := m.dialDirect(addr); conn != nil {
		return conn, nil
	}

	ethAddr, err := addr.ETH()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel := make(chan connTuple)

	go func() {
		puncher, err := m.puncherNew()
		if err != nil {
			channel <- newConnTuple(nil, err)
			return
		}

		channel <- newConnTuple(puncher.Dial(ethAddr))
	}()

	select {
	case conn := <-channel:
		if conn.err != nil && m.relayDial != nil {
			return m.relayDial(ethAddr)
		} else {
			return conn.unwrap()
		}
	case <-ctx.Done():
		return m.relayDial(ethAddr)
	}
}

func (m *Dialer) dialDirect(addr auth.Addr) net.Conn {
	netAddr, err := addr.Addr()
	if err == nil {
		conn, err := net.Dial("tcp", netAddr)
		if err == nil {
			return conn
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
