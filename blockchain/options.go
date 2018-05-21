package blockchain

import "time"

const (
	defaultEthEndpoint        = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"
	defaultSidechainEndpoint  = "https://sidechain-dev.sonm.com"
	defaultGasPrice           = 20000000000 // 20 Gwei
	defaultGasPriceSidechain  = 0
	defaultBlockConfirmations = 5
)

type options struct {
	gasPrice             int64
	gasPriceSidechain    int64
	apiEndpoint          string
	apiSidechainEndpoint string
	logParsePeriod       time.Duration
	blockConfirmations   int64
}

func defaultOptions() *options {
	return &options{
		gasPrice:             defaultGasPrice,
		gasPriceSidechain:    defaultGasPriceSidechain,
		apiEndpoint:          defaultEthEndpoint,
		apiSidechainEndpoint: defaultSidechainEndpoint,
		logParsePeriod:       time.Second,
		blockConfirmations:   defaultBlockConfirmations,
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

func WithSidechainEndpoint(s string) Option {
	return func(o *options) {
		o.apiSidechainEndpoint = s
	}
}

func WithConfig(cfg *Config) Option {
	return func(o *options) {
		if cfg != nil {
			o.apiEndpoint = cfg.Endpoint.String()
			o.apiSidechainEndpoint = cfg.SidechainEndpoint.String()
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.logParsePeriod = d
	}
}

func WithBlockConfirmations(c int64) Option {
	return func(o *options) {
		o.blockConfirmations = c
	}
}
