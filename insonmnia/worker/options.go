package worker

import (
	"context"

	"github.com/sonm-io/core/insonmnia/logging"
)

type Option func(*options)

type options struct {
	ctx        context.Context
	version    string
	logWatcher *logging.WatcherCore
}

func newOptions() *options {
	return &options{
		ctx:     context.Background(),
		version: "unspecified",
	}
}

func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

func WithVersion(v string) Option {
	return func(o *options) {
		o.version = v
	}
}

func WithLogWatcher(watcher *logging.WatcherCore) Option {
	return func(o *options) {
		o.logWatcher = watcher
	}
}
