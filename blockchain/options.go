package blockchain

import "time"

const (
	defaultLivechainEndpoint  = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"
	defaultSidechainEndpoint  = "https://sidechain-dev.sonm.com"
	defaultLivechainGasPrice  = 20000000000 // 20 Gwei
	defaultSidechainGasPrice  = 0
	defaultBlockConfirmations = 5
	defaultLogParsePeriod     = time.Second
)

// chainOpts describes common options
// for almost any geth-node connection
// (live Eth network, rinkeby, SONM sidechain
// or local geth-node for testing).
type chainOpts struct {
	gasPrice           int64
	endpoint           string
	logParsePeriod     time.Duration
	blockConfirmations int64
}

func (c *chainOpts) newClient() (CustomEthereumClient, error) {
	return NewClient(c.endpoint)
}

type options struct {
	livechain *chainOpts
	sidechain *chainOpts
}

func defaultOptions() *options {
	return &options{
		livechain: &chainOpts{
			gasPrice:       defaultLivechainGasPrice,
			endpoint:       defaultLivechainEndpoint,
			logParsePeriod: defaultLogParsePeriod,
		},
		sidechain: &chainOpts{
			gasPrice:           defaultSidechainGasPrice,
			endpoint:           defaultSidechainEndpoint,
			logParsePeriod:     defaultLogParsePeriod,
			blockConfirmations: defaultBlockConfirmations,
		},
	}
}

type Option func(options *options)

func WithLivechainGasPrice(p int64) Option {
	return func(o *options) {
		o.livechain.gasPrice = p
	}
}

func WithSidechainGasPrice(p int64) Option {
	return func(o *options) {
		o.sidechain.gasPrice = p
	}
}

func WithLivechainEndpoint(s string) Option {
	return func(o *options) {
		o.livechain.endpoint = s
	}
}

func WithSidechainEndpoint(s string) Option {
	return func(o *options) {
		o.sidechain.endpoint = s
	}
}

func WithConfig(cfg *Config) Option {
	return func(o *options) {
		if cfg != nil {
			o.livechain.endpoint = cfg.Endpoint.String()
			o.sidechain.endpoint = cfg.SidechainEndpoint.String()
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.livechain.logParsePeriod = d
		o.sidechain.logParsePeriod = d
	}
}

func WithBlockConfirmations(c int64) Option {
	return func(o *options) {
		o.sidechain.blockConfirmations = c
	}
}
