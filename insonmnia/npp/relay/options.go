package relay

import "go.uber.org/zap"

// Option is a function that configures the server.
type Option func(options *options) error

type options struct {
	bufferSize int
	log        *zap.Logger
}

func newOptions() *options {
	return &options{
		bufferSize: 32 * 1024,
		log:        zap.NewNop(),
	}
}

// WithLogger is an option that specifies provided logger used for the internal
// logging.
// Nil value is supported and can be passed to deactivate the logging system
// entirely.
func WithLogger(log *zap.Logger) Option {
	return func(options *options) error {
		options.log = log
		return nil
	}
}
