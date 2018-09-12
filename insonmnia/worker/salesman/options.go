package salesman

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

type EthAPI interface {
	MarketAddress() common.Address
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklistType sonm.BlacklistType) error
	GetDealInfo(ctx context.Context, dealID *big.Int) (*sonm.Deal, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *sonm.Order) (*sonm.Order, error)
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*sonm.Order, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) error
}

type ethAPI struct {
	blockchain.MarketAPI
	blockchain.ContractRegistry
}

func NewEthAPI(api blockchain.API) *ethAPI {
	return &ethAPI{api.Market(), api.ContractRegistry()}
}

type options struct {
	log           *zap.SugaredLogger
	storage       state.SimpleStorage
	resources     *resource.Scheduler
	hardware      *hardware.Hardware
	eth           EthAPI
	cGroupManager cgroups.CGroupManager
	matcher       matcher.Matcher
	ethkey        *ecdsa.PrivateKey
	config        *YAMLConfig
}

func WithLogger(log *zap.SugaredLogger) Option {
	return func(opts *options) {
		opts.log = log
	}
}

func WithStorage(storage state.SimpleStorage) Option {
	return func(opts *options) {
		opts.storage = storage
	}
}
func WithResources(resources *resource.Scheduler) Option {
	return func(opts *options) {
		opts.resources = resources
	}
}
func WithHardware(hardware *hardware.Hardware) Option {
	return func(opts *options) {
		opts.hardware = hardware
	}
}
func WithEth(eth EthAPI) Option {
	return func(opts *options) {
		opts.eth = eth
	}
}
func WithCGroupManager(cGroupManager cgroups.CGroupManager) Option {
	return func(opts *options) {
		opts.cGroupManager = cGroupManager
	}
}
func WithMatcher(matcher matcher.Matcher) Option {
	return func(opts *options) {
		opts.matcher = matcher
	}
}
func WithEthkey(ethkey *ecdsa.PrivateKey) Option {
	return func(opts *options) {
		opts.ethkey = ethkey
	}
}
func WithConfig(config *YAMLConfig) Option {
	return func(opts *options) {
		opts.config = config
	}
}
func (m *options) Validate() error {
	err := multierror.NewMultiError()

	if m.log == nil {
		err = multierror.Append(err, errors.New("WithLogger option is required"))
	}

	if m.storage == nil {
		err = multierror.Append(err, errors.New("WithStorage option is required"))
	}

	if m.resources == nil {
		err = multierror.Append(err, errors.New("WithResources option is required"))
	}

	if m.hardware == nil {
		err = multierror.Append(err, errors.New("WithHardware option is required"))
	}

	if m.eth == nil {
		err = multierror.Append(err, errors.New("WithEth option is required"))
	}

	if m.cGroupManager == nil {
		err = multierror.Append(err, errors.New("WithCGroupManager option is required"))
	}

	if m.matcher == nil {
		err = multierror.Append(err, errors.New("WithMatcher option is required"))
	}

	if m.ethkey == nil {
		err = multierror.Append(err, errors.New("WithEthkey option is required"))
	}

	if m.config == nil {
		err = multierror.Append(err, errors.New("WithConfig option is required"))
	}

	return err.ErrorOrNil()
}

type Option func(*options)
