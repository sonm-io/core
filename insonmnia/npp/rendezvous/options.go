package rendezvous

import (
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type options struct {
	log         *zap.Logger
	credentials credentials.TransportCredentials
}

func newOptions() *options {
	return &options{
		log:         zap.NewNop(),
		credentials: nil,
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
		options.log = log
	}
}

// WithCredentials is an option that specifies transport credentials used for
// establishing secure connections between peer and the server.
// Nil value is also supported, but discouraged, because it disables the
// authentication. Clients that require TLS will fail the handshake process
// and disconnect from such server.
// Note that we use ETH based TLS credentials.
func WithCredentials(credentials credentials.TransportCredentials) Option {
	return func(options *options) {
		options.credentials = credentials
	}
}
