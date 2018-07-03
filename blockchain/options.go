package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

const (
	defaultMasterchainEndpoint = "https://mainnet.infura.io/00iTrs5PIy0uGODwcsrb"
	defaultSidechainEndpoint   = "https://sidechain.livenet.sonm.com/"
	defaultMasterchainGasPrice = 20000000000 // 20 Gwei
	defaultSidechainGasPrice   = 0
	defaultBlockConfirmations  = 5
	defaultLogParsePeriod      = time.Second
	defaultMasterchainGasLimit = 500000
	defaultSidechainGasLimit   = 2000000

	approveGasLimit = 70000
	payinGasLimit   = 100000
	payoutGasLimit  = 100000
)

// chainOpts describes common options
// for almost any geth-node connection using JSON-RPC
// (live Eth network, rinkeby, SONM sidechain
// or local geth-node for testing).
type chainOpts struct {
	gasPrice           int64
	gasLimit           uint64
	endpoint           string
	logParsePeriod     time.Duration
	blockConfirmations int64
	client             CustomEthereumClient
}

func (c *chainOpts) getClient() (CustomEthereumClient, error) {
	var err error
	if c.client == nil {
		c.client, err = NewClient(c.endpoint)
	}

	return c.client, err
}

// getTxOpts returns options for transaction execution specified to chain
func (c *chainOpts) getTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit uint64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	opts.GasLimit = gasLimit
	opts.GasPrice = big.NewInt(c.gasPrice)

	return opts
}

type options struct {
	masterchain      *chainOpts
	sidechain        *chainOpts
	contractRegistry common.Address
	blocksBatchSize  int64
}

func defaultOptions() *options {
	return &options{
		masterchain: &chainOpts{
			gasPrice:           defaultMasterchainGasPrice,
			gasLimit:           defaultMasterchainGasLimit,
			endpoint:           defaultMasterchainEndpoint,
			logParsePeriod:     defaultLogParsePeriod,
			blockConfirmations: defaultBlockConfirmations,
		},
		sidechain: &chainOpts{
			gasPrice:           defaultSidechainGasPrice,
			gasLimit:           defaultSidechainGasLimit,
			endpoint:           defaultSidechainEndpoint,
			logParsePeriod:     defaultLogParsePeriod,
			blockConfirmations: defaultBlockConfirmations,
		},
		contractRegistry: common.HexToAddress(defaultContractRegistryAddr),
	}
}

type Option func(options *options)

func WithMasterchainGasPrice(p int64) Option {
	return func(o *options) {
		o.masterchain.gasPrice = p
	}
}

func WithSidechainGasPrice(p int64) Option {
	return func(o *options) {
		o.sidechain.gasPrice = p
	}
}

func WithMasterchainEndpoint(s string) Option {
	return func(o *options) {
		o.masterchain.endpoint = s
	}
}

func WithSidechainEndpoint(s string) Option {
	return func(o *options) {
		o.sidechain.endpoint = s
	}
}

func WithBlocksBatchSize(batchSize int64) Option {
	return func(o *options) {
		o.blocksBatchSize = batchSize
	}
}

func WithConfig(cfg *Config) Option {
	return func(o *options) {
		if cfg != nil {
			o.masterchain.endpoint = cfg.Endpoint.String()
			o.sidechain.endpoint = cfg.SidechainEndpoint.String()
			if cfg.ContractRegistryAddr.Big().Cmp(big.NewInt(0)) != 0 {
				o.contractRegistry = cfg.ContractRegistryAddr
			}
			o.blocksBatchSize = cfg.BlocksBatchSize
		}
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *options) {
		o.masterchain.logParsePeriod = d
		o.sidechain.logParsePeriod = d
	}
}

func WithBlockConfirmations(c int64) Option {
	return func(o *options) {
		o.sidechain.blockConfirmations = c
	}
}

func WithSidechainClient(c CustomEthereumClient) Option {
	return func(o *options) {
		o.sidechain.client = c
	}
}

func WithMasterchainClient(c CustomEthereumClient) Option {
	return func(o *options) {
		o.masterchain.client = c
	}
}

func WithContractRegistry(address common.Address) Option {
	return func(o *options) {
		o.contractRegistry = address
	}
}
