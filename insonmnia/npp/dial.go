// This package is responsible for Client side for NAT Punching Protocol.

package npp

import (
	"context"
	"net"

	"github.com/ethereum/go-ethereum/common"
)

// Dialer represents an NPP dialer.
//
// This structure acts like an usual dialer with an exception that the address
// must be an authenticated endpoint and the connection establishment process
// is done via NAT Punching Protocol.
type Dialer struct {
	ctx     context.Context
	puncher NATPuncher
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
		ctx:     ctx,
		puncher: opts.puncher,
	}, nil
}

// Dial dials the given verified address using NPP.
func (m *Dialer) Dial(addr common.Address) (net.Conn, error) {
	return m.puncher.Dial(addr)
}

// Close closes the dialer.
//
// Any blocked operations will be unblocked and return errors.
func (m *Dialer) Close() error {
	return m.puncher.Close()
}
