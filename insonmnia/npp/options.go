package npp

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/insonmnia/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// Option is a function that configures the listener or dialer.
type Option func(o *options) error

type options struct {
	ctx        context.Context
	log        *zap.Logger
	puncher    NATPuncher
	puncherNew func() (NATPuncher, error)
	nppBacklog int
}

func newOptions(ctx context.Context) *options {
	return &options{
		ctx:        ctx,
		nppBacklog: 128,
	}
}

// WithRendezvous is an option that specifies Rendezvous client settings.
//
// Without this option no intermediate server will be used for obtaining
// peer's endpoints and the entire connection establishment process will fall
// back to the old good plain TCP connection.
func WithRendezvous(addrs []auth.Endpoint, credentials credentials.TransportCredentials) Option {
	return func(o *options) error {
		o.puncherNew = func() (NATPuncher, error) {
			for _, addr := range addrs {
				client, err := newRendezvousClient(o.ctx, addr, credentials)
				if err == nil {
					return newNATPuncher(o.ctx, client)
				}
			}

			return nil, fmt.Errorf("failed to connect to %+v", addrs)
		}

		return nil
	}
}

// WithLogger is an option that specifies provided logger used for the internal
// logging.
// Nil value is supported and can be passed to deactivate the logging system
// entirely.
func WithLogger(log *zap.Logger) Option {
	return func(o *options) error {
		o.log = log
		return nil
	}
}

// WithNPPBacklog is an option that specifies NPP backlog size.
func WithNPPBacklog(backlog int) Option {
	return func(o *options) error {
		o.nppBacklog = backlog
		return nil
	}
}
