package npp

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// Option is a function that configures the listener or dialer.
type Option func(o *options) error

type options struct {
	log                   *zap.Logger
	puncher               NATPuncher
	puncherNew            func(ctx context.Context) (NATPuncher, error)
	nppBacklog            int
	nppMinBackoffInterval time.Duration
	nppMaxBackoffInterval time.Duration
	relayListener         *relay.Listener
	relayDialer           *relay.Dialer
	protocol              string
}

func newOptions() *options {
	return &options{
		log:                   zap.NewNop(),
		nppBacklog:            128,
		nppMinBackoffInterval: 500 * time.Millisecond,
		nppMaxBackoffInterval: 8000 * time.Millisecond,
		protocol:              sonm.DefaultNPPProtocol,
	}
}

// WithRendezvous is an option that specifies Rendezvous client settings.
//
// Without this option no intermediate server will be used for obtaining
// peer's endpoints and the entire connection establishment process will fall
// back to the old good plain TCP connection.
func WithRendezvous(cfg rendezvous.Config, credentials credentials.TransportCredentials) Option {
	return func(o *options) error {
		if len(cfg.Endpoints) == 0 {
			return nil
		}

		o.puncherNew = func(ctx context.Context) (NATPuncher, error) {
			for _, addr := range cfg.Endpoints {
				client, err := newRendezvousClient(ctx, addr, credentials)
				if err == nil {
					return newNATPuncher(ctx, cfg, client, o.protocol, o.log)
				}
			}

			return nil, fmt.Errorf("failed to connect to %+v", cfg.Endpoints)
		}

		return nil
	}
}

// WithRelay is an option that activates Relay fallback on a NPP dialer and
// listener.
//
// Without this option no intermediate server will be used for relaying
// TCP.
// One or more Relay TCP addresses must be specified in "cfg.Endpoints"
// argument. Hostname resolution is performed for each of them for environments
// with dynamic DNS addition/removal. Thus, a single Relay endpoint as a
// hostname should fit the best.
// The "credentials" argument is used both for extracting the ETH address of
// a server and for request signing to ensure that the published server
// actually owns the ETH address is publishes. When dialing this argument is
// currently ignored and can be "nil".
func WithRelay(cfg relay.Config, credentials *ecdsa.PrivateKey) Option {
	return func(o *options) error {
		dialer := &relay.Dialer{
			Addrs: cfg.Endpoints,
			Log:   o.log,
		}

		listener, err := relay.NewListener(cfg.Endpoints, credentials, o.log)
		if err != nil {
			return err
		}

		o.relayDialer = dialer
		o.relayListener = listener
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

// WithNPPBackoff is an option that specifies NPP timeouts.
func WithNPPBackoff(min, max time.Duration) Option {
	return func(o *options) error {
		o.nppMinBackoffInterval = min
		o.nppMaxBackoffInterval = max
		return nil
	}
}

func WithProtocol(protocol string) Option {
	return func(o *options) error {
		o.protocol = protocol
		return nil
	}
}
