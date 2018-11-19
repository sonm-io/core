package rendezvous

import (
	"crypto/tls"

	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type options struct {
	Log         *zap.Logger
	Credentials *xgrpc.TransportCredentials
	EnableQUIC  bool
}

func newOptions() *options {
	return &options{
		Log:         zap.NewNop(),
		Credentials: nil,
	}
}

// Option is a function that configures the server.
type Option func(options *options)

// WithLogger is an option that specifies provided logger used for the internal
// logging.
// Nil value is supported and can be passed to deactivate the logging system
// entirely.
func WithLogger(log *zap.Logger) Option {
	return func(options *options) {
		options.Log = log
	}
}

// WithCredentials is an option that specifies transport credentials used for
// establishing secure connections between peer and the server.
// Nil value is also supported, but discouraged, because it disables the
// authentication. Clients that require TLS will fail the handshake process
// and disconnect from such server.
// Note that we use ETH based TLS credentials.
func WithCredentials(cfg *tls.Config) Option {
	return func(options *options) {
		options.Credentials = xgrpc.NewTransportCredentials(cfg)
	}
}

// WithQUIC activates QUIC support in the Rendezvous server, allowing to
// penetrate NAT for UDP.
// When using this option it is REQUIRED to specify transport credentials by
// passing "WithCredentials" option, because QUIC provides security protection
// equivalent to TLS.
func WithQUIC() Option {
	return func(options *options) {
		options.EnableQUIC = true
	}
}
