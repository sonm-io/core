package npp

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

// Option is a function that configures the listener or dialer.
type Option func(o *options) error

type puncherFactory func(ctx context.Context) (NATPuncher, error)

type options struct {
	log                   *zap.Logger
	puncherNew            puncherFactory
	puncherNewQUIC        puncherFactory
	nppBacklog            int
	nppMinBackoffInterval time.Duration
	nppMaxBackoffInterval time.Duration
	relayListener         *relay.Listener
	relayDialer           *relay.Dialer
	Protocol              string
}

func newOptions() *options {
	return &options{
		log:                   zap.NewNop(),
		nppBacklog:            128,
		nppMinBackoffInterval: 500 * time.Millisecond,
		nppMaxBackoffInterval: 8000 * time.Millisecond,
		Protocol:              sonm.DefaultNPPProtocol,
	}
}

// WithRendezvous is an option that specifies Rendezvous client settings and
// activates NAT punching protocol.
//
// Without this option no intermediate server will be used for obtaining
// peer's endpoints and the entire connection establishment process will fall
// back to the old good plain TCP connection.
func WithRendezvous(cfg rendezvous.Config, credentials *xgrpc.TransportCredentials) Option {
	return func(o *options) error {
		if len(cfg.Endpoints) == 0 {
			return nil
		}

		o.puncherNew = newTCPPuncherFactory(cfg, credentials, o)

		if credentials.TLSConfig != nil {
			// Preliminary create and save UDP socket for QUIC communication.
			//
			// We chose the port automatically here. However, the UDP socket is
			// reused for ALL connections to be able to keep NAT mapping
			// unchanged. This increases successful connection establishing
			// probability after the hole has been punched at least once.
			//
			// IPv4 restriction is required, because in case of dual-stack
			// remote network with global IPv6 address NAT isn't a problem anymore.
			conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
			if err != nil {
				return err
			}

			o.puncherNewQUIC = newQUICPuncherFactory(cfg, credentials, conn, o)
		}

		return nil
	}
}

func newTCPPuncherFactory(cfg rendezvous.Config, credentials *xgrpc.TransportCredentials, options *options) puncherFactory {
	return func(ctx context.Context) (NATPuncher, error) {
		errs := multierror.NewMultiError()

		for _, addr := range cfg.Endpoints {
			client, err := newRendezvousClient(ctx, addr, credentials)
			if err != nil {
				errs = multierror.AppendUnique(errs, err)
				continue
			}

			return newNATPuncher(ctx, cfg, client, options.Protocol, options.log)
		}

		return nil, fmt.Errorf("failed to connect to %+v: %v", cfg.Endpoints, errs.Error())
	}
}

func newQUICPuncherFactory(cfg rendezvous.Config, credentials *xgrpc.TransportCredentials, conn net.PacketConn, options *options) puncherFactory {
	return func(ctx context.Context) (NATPuncher, error) {
		errs := multierror.NewMultiError()

		for _, addr := range cfg.Endpoints {
			client, err := newRendezvousQUICClient(ctx, conn, addr, credentials)
			if err != nil {
				errs = multierror.AppendUnique(errs, err)
				continue
			}

			return newQUICPuncher(client, credentials.TLSConfig, options.Protocol, options.log)
		}

		return nil, fmt.Errorf("failed to connect to %+v: %v", cfg.Endpoints, errs.Error())
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

// WithProtocol is an option that specifies application level protocol.
//
// In case of servers it will publish itself with a connection ID "PROTOCOL://ETH_ADDRESS".
// In case of clients this option helps to distinguish whether the destination
// peer supports such protocol.
// For example this option is used for punching NAT for SSH connections.
func WithProtocol(protocol string) Option {
	return func(o *options) error {
		o.Protocol = protocol
		return nil
	}
}
