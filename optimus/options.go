package optimus

import "go.uber.org/zap"

type Option func(o *options)

type options struct {
	Version string
	Log     *zap.SugaredLogger
}

func newOptions() *options {
	return &options{
		Version: "unspecified",
		Log:     zap.NewNop().Sugar(),
	}
}

func WithVersion(version string) Option {
	return func(o *options) {
		o.Version = version
	}
}

func WithLog(log *zap.SugaredLogger) Option {
	return func(o *options) {
		o.Log = log
	}
}
