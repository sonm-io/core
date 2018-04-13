package blockchain

import "time"

const (
	defaultEthEndpoint = "https://private-dev.sonm.io"
	defaultGasPrice    = 0
)

type options struct {
	gasPrice       int64
	apiEndpoint    string
	logParsePeriod time.Duration
}

func defaultOptions() *options {
	return &options{
		gasPrice:       defaultGasPrice,
		apiEndpoint:    defaultEthEndpoint,
		logParsePeriod: time.Second,
	}
}

type Option func(options *options)

func WithGasPrice(p int64) Option {
	return func(o *options) {
		o.gasPrice = p
	}
}

func WithEthEndpoint(s string) Option {
	return func(o *options) {
		o.apiEndpoint = s
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.logParsePeriod = d
	}
}
