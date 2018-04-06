package blockchain

const (
	defaultEthEndpoint = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"
	defaultGasPrice    = 20 * 1000000000
)

type options struct {
	gasPrice    int64
	apiEndpoint string
}

func defaultOptions() *options {
	return &options{
		gasPrice:    defaultGasPrice,
		apiEndpoint: defaultEthEndpoint,
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
