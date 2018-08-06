package volume

import (
	"fmt"

	"go.uber.org/zap"
)

// Option specifies how a Docker plugin can be configured.
type Option func(options interface{}) error

// Options describes generic volume plugin options.
type options struct {
	socketDir string
	log       *zap.SugaredLogger
}

func newOptions() *options {
	return &options{
		socketDir: defaultPluginSockDir,
		log:       zap.NewNop().Sugar(),
	}
}

// WithPluginSocketDir constructs an option that specifies the plugin
// directory where Unix sockets live.
func WithPluginSocketDir(path string) Option {
	return func(o interface{}) error {
		option, ok := o.(*options)
		if !ok {
			return fmt.Errorf("invalid option type: %T", o)
		}

		option.socketDir = path
		return nil
	}
}

func WithLogger(log *zap.SugaredLogger) Option {
	return func(o interface{}) error {
		option, ok := o.(*options)
		if !ok {
			return fmt.Errorf("invalid option type: %T", o)
		}

		option.log = log
		return nil
	}
}

// WithOptions constructs an option that forwards the given generic options
// to the plugin.
func WithOptions(opts map[string]string) Option {
	return func(o interface{}) error {
		switch o.(type) {
		case *options:
			return nil
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}
	}
}
