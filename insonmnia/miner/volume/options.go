package volume

import (
	"fmt"
)

// Option specifies how a Docker plugin can be configured.
type Option func(options interface{}) error

// Options describes generic volume plugin options.
type Options struct {
	SocketDir string
}

// WithPluginSocketDir constructs an option that specifies the plugin
// directory where Unix sockets live.
func WithPluginSocketDir(path string) Option {
	return func(o interface{}) error {
		option, ok := o.(*Options)
		if !ok {
			return fmt.Errorf("invalid option type: %T", o)
		}

		option.SocketDir = path
		return nil
	}
}

// WithOptions constructs an option that forwards the given generic options
// to the plugin.
func WithOptions(options map[string]string) Option {
	return func(o interface{}) error {
		switch o.(type) {
		case *Options:
			return nil
		default:
			return fmt.Errorf("invalid option type: %T", o)
		}
	}
}
