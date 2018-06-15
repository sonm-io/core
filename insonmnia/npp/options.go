package npp

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/util/netutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// Option is a function that configures the listener or dialer.
type Option func(o *options) error

type options struct {
	ctx                   context.Context
	log                   *zap.Logger
	puncher               NATPuncher
	puncherNew            func() (NATPuncher, error)
	nppBacklog            int
	nppMinBackoffInterval time.Duration
	nppMaxBackoffInterval time.Duration
	relayListen           func() (net.Conn, error)
	relayDial             func(target common.Address) (net.Conn, error)
}

func newOptions(ctx context.Context) *options {
	return &options{
		ctx:                   ctx,
		log:                   zap.NewNop(),
		nppBacklog:            128,
		nppMinBackoffInterval: 500 * time.Millisecond,
		nppMaxBackoffInterval: 8000 * time.Millisecond,
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

		o.puncherNew = func() (NATPuncher, error) {
			for _, addr := range cfg.Endpoints {
				client, err := newRendezvousClient(o.ctx, addr, credentials)
				if err == nil {
					return newNATPuncher(o.ctx, cfg, client)
				}
			}

			return nil, fmt.Errorf("failed to connect to %+v", cfg.Endpoints)
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

// WithNPPBackoff is an option that specifies NPP timeouts.
func WithNPPBackoff(min, max time.Duration) Option {
	return func(o *options) error {
		o.nppMinBackoffInterval = min
		o.nppMaxBackoffInterval = max
		return nil
	}
}

// WithRelay is an option that specifies Relay client settings.
//
// Without this option no intermediate server will be used for relaying
// TCP.
func WithRelay(addrs []netutil.TCPAddr, key *ecdsa.PrivateKey, log *zap.Logger) Option {
	return func(o *options) error {
		signedAddr, err := relay.NewSignedAddr(key)
		if err != nil {
			return err
		}

		o.relayListen = func() (net.Conn, error) {
			for _, addr := range addrs {
				conn, err := relay.ListenWithLog(&addr, signedAddr, log)
				if err == nil {
					return conn, nil
				}
			}

			return nil, fmt.Errorf("failed to connect to %+v", addrs)
		}

		return nil
	}
}

func WithRelayClient(addrs []netutil.TCPAddr, log *zap.Logger) Option {
	return func(o *options) error {
		o.relayDial = func(target common.Address) (net.Conn, error) {
			for _, addr := range addrs {
				conn, err := relay.DialWithLog(&addr, target, "", log)
				if err == nil {
					return conn, nil
				}
			}

			return nil, fmt.Errorf("failed to connect to %+v", addrs)
		}

		return nil
	}
}
