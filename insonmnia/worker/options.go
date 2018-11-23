package worker

import (
	"context"
)

type Option func(*options)

type options struct {
	ctx     context.Context
	version string
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
