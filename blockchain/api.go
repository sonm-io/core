package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain/market"
	marketAPI "github.com/sonm-io/core/blockchain/market/api"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	NumCurrentBenchmarks = 12
)

type API interface {
	ProfileRegistry() ProfileRegistryAPI
	Events() EventsAPI
	Market() MarketAPI
	Blacklist() BlacklistAPI
	LiveToken() TokenAPI
	SideToken() TokenAPI
	TestToken() TestTokenAPI
	LiveGatekeeper() SimpleGatekeeperAPI
	SideGatekeeper() SimpleGatekeeperAPI
	OracleUSD() OracleUSDAPI
}

type ProfileRegistryAPI interface {
	GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error)
	GetCertificate(ctx context.Context, certificateID *big.Int) (*pb.Certificate, error)
}

type EventsAPI interface {
	GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error)
}

type MarketAPI interface {
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) <-chan DealOrError
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) <-chan error
	GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.Deal, error)
	GetDealsAmount(ctx context.Context) (*big.Int, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) <-chan OrderOrError
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) <-chan error
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.Order, error)
	GetOrdersAmount(ctx context.Context) (*big.Int, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) <-chan error
	RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (*types.Transaction, error)
	ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (*types.Transaction, error)
	RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (*types.Transaction, error)
	GetMaster(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (common.Address, error)
	GetDealChangeRequestInfo(ctx context.Context, dealID *big.Int) (*pb.DealChangeRequest, error)
	GetNumBenchmarks(ctx context.Context) (int, error)
}

type BlacklistAPI interface {
	Check(ctx context.Context, who, whom common.Address) (bool, error)
	Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error)
	Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) (*types.Transaction, error)
	AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error)
	RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error)
	SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error)
}

// TokenAPI is a go implementation of ERC20-compatibility token with full functionality high-level interface
// standard description with placed: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20-token-standard.md
type TokenAPI interface {
	// Approve - add allowance from caller to other contract to spend tokens
	Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// Transfer token from caller
	Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// TransferFrom fallback function for contracts to transfer you allowance
	TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error)
	// BalanceOf returns balance of given address
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
	// AllowanceOf returns allowance of given address to spender account
	AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error)
	// TotalSupply - all amount of emitted token
	TotalSupply(ctx context.Context) (*big.Int, error)
}

type TestTokenAPI interface {
	// GetTokens - send 100 SNMT token for message caller
	// this function added for MVP purposes and has been deleted later
	GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error)
}

type SimpleGatekeeperAPI interface {
	PayIn(ctx context.Context, key *ecdsa.PrivateKey, value *big.Int) <-chan error
	Payout(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, value, txNumber *big.Int) <-chan error
}

type OracleUSDAPI interface {
	SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error)

	GetCurrentPrice(ctx context.Context) (*big.Int, error)
}

type BasicAPI struct {
	market          MarketAPI
	liveToken       TokenAPI
	sideToken       TokenAPI
	testToken       TestTokenAPI
	blacklist       BlacklistAPI
	profileRegistry ProfileRegistryAPI
	events          EventsAPI
	liveGatekeeper  SimpleGatekeeperAPI
	sideGatekeeper  SimpleGatekeeperAPI
	oracle          OracleUSDAPI
}

func NewAPI(opts ...Option) (API, error) {
	defaults := defaultOptions()
	for _, o := range opts {
		o(defaults)
	}

	client, err := initEthClient(defaults.apiEndpoint)
	if err != nil {
		return nil, err
	}

	liveToken, err := NewStandardToken(client, market.SNMAddr(), defaults.gasPrice)
	if err != nil {
		return nil, err
	}

	testToken, err := NewTestToken(client, market.SNMAddr(), defaults.gasPrice)
	if err != nil {
		return nil, err
	}

	liveGate, err := NewBasicSimpleGatekeeperAPI(client, market.GatekeeperLiveAddr(), defaults.gasPrice, defaults.logParsePeriod)
	if err != nil {
		return nil, err
	}

	clientSidechain, err := initEthClient(defaults.apiSidechainEndpoint)
	if err != nil {
		return nil, err
	}

	blacklist, err := NewBasicBlacklist(clientSidechain, market.BlacklistAddr(), defaults.gasPriceSidechain)
	if err != nil {
		return nil, err
	}

	marketApi, err := NewBasicMarket(clientSidechain, market.MarketAddr(), defaults.gasPriceSidechain, defaults.logParsePeriod)
	if err != nil {
		return nil, err
	}

	profileRegistry, err := NewBasicProfileRegistry(clientSidechain, market.ProfileRegistryAddr(), defaults.gasPriceSidechain)
	if err != nil {
		return nil, err
	}

	sideToken, err := NewStandardToken(clientSidechain, market.SNMSidechainAddr(), defaults.gasPriceSidechain)
	if err != nil {
		return nil, err
	}

	events, err := NewEventsAPI(clientSidechain, ctxlog.GetLogger(context.Background()))
	if err != nil {
		return nil, err
	}

	sideGate, err := NewBasicSimpleGatekeeperAPI(clientSidechain, market.GatekeeperSidechainAddr(), defaults.gasPrice, defaults.logParsePeriod)
	if err != nil {
		return nil, err
	}

	oracle, err := NewBasicOracleUSDAPI(clientSidechain, market.OracleUsdAddr(), defaults.gasPrice)
	if err != nil {
		return nil, err
	}

	return &BasicAPI{
		market:          marketApi,
		blacklist:       blacklist,
		profileRegistry: profileRegistry,
		liveToken:       liveToken,
		sideToken:       sideToken,
		testToken:       testToken,
		events:          events,
		liveGatekeeper:  liveGate,
		sideGatekeeper:  sideGate,
		oracle:          oracle,
	}, nil
}

func (api *BasicAPI) Market() MarketAPI {
	return api.market
}

func (api *BasicAPI) LiveToken() TokenAPI {
	return api.liveToken
}

func (api *BasicAPI) SideToken() TokenAPI {
	return api.sideToken
}

func (api *BasicAPI) TestToken() TestTokenAPI {
	return api.testToken
}

func (api *BasicAPI) Blacklist() BlacklistAPI {
	return api.blacklist
}

func (api *BasicAPI) ProfileRegistry() ProfileRegistryAPI {
	return api.profileRegistry
}

func (api *BasicAPI) Events() EventsAPI {
	return api.events
}

func (api *BasicAPI) LiveGatekeeper() SimpleGatekeeperAPI {
	return api.liveGatekeeper
}

func (api *BasicAPI) SideGatekeeper() SimpleGatekeeperAPI {
	return api.sideGatekeeper
}

func (api *BasicAPI) OracleUSD() OracleUSDAPI {
	return api.oracle
}

type BasicMarketAPI struct {
	client         *ethclient.Client
	marketContract *marketAPI.Market
	gasPrice       int64
	logParsePeriod time.Duration
}

func NewBasicMarket(client *ethclient.Client, address common.Address, gasPrice int64, logParsePeriod time.Duration) (MarketAPI, error) {
	marketContract, err := marketAPI.NewMarket(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicMarketAPI{
		client:         client,
		marketContract: marketContract,
		gasPrice:       gasPrice,
		logParsePeriod: logParsePeriod,
	}, nil
}

func (api *BasicMarketAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) <-chan DealOrError {
	ch := make(chan DealOrError, 0)
	go api.openDeal(ctx, key, askID, bidID, ch)
	return ch
}

func (api *BasicMarketAPI) openDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int, ch chan DealOrError) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.marketContract.OpenDeal(opts, askID, bidID)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	id, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.DealOpenedTopic)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	deal, err := api.GetDealInfo(ctx, id)
	ch <- DealOrError{deal, err}
}

func (api *BasicMarketAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) <-chan error {
	ch := make(chan error, 0)
	go api.closeDeal(ctx, key, dealID, blacklisted, ch)
	return ch
}

func (api *BasicMarketAPI) closeDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.marketContract.CloseDeal(opts, dealID, blacklisted)
	if err != nil {
		ch <- err
		return
	}

	_, err = waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.DealUpdatedTopic)
	if err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.Deal, error) {
	deal1, err := api.marketContract.GetDealInfo(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	deal2, err := api.marketContract.GetDealParams(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	benchmarks, err := pb.NewBenchmarks(deal1.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &pb.Deal{
		Id:             pb.NewBigInt(dealID),
		Benchmarks:     benchmarks,
		SupplierID:     pb.NewEthAddress(deal1.SupplierID),
		ConsumerID:     pb.NewEthAddress(deal1.ConsumerID),
		MasterID:       pb.NewEthAddress(deal1.MasterID),
		AskID:          pb.NewBigInt(deal1.AskID),
		BidID:          pb.NewBigInt(deal1.BidID),
		Duration:       deal2.Duration.Uint64(),
		Price:          pb.NewBigInt(deal2.Price),
		StartTime:      &pb.Timestamp{Seconds: deal1.StartTime.Int64()},
		EndTime:        &pb.Timestamp{Seconds: deal2.EndTime.Int64()},
		Status:         pb.DealStatus(deal2.Status),
		BlockedBalance: pb.NewBigInt(deal2.BlockedBalance),
		TotalPayout:    pb.NewBigInt(deal2.TotalPayout),
		LastBillTS:     &pb.Timestamp{Seconds: deal2.LastBillTS.Int64()},
	}, nil
}

func (api *BasicMarketAPI) GetDealsAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetDealsAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) <-chan OrderOrError {
	ch := make(chan OrderOrError, 0)
	go api.placeOrder(ctx, key, order, ch)
	return ch
}

func (api *BasicMarketAPI) placeOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order, ch chan OrderOrError) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)

	fixedNetflags := pb.UintToNetflags(order.Netflags)
	var fixedTag [32]byte
	copy(fixedTag[:], order.Tag[:])

	tx, err := api.marketContract.PlaceOrder(opts,
		uint8(order.OrderType),
		order.CounterpartyID.Unwrap(),
		big.NewInt(int64(order.Duration)),
		order.Price.Unwrap(),
		fixedNetflags,
		uint8(order.IdentityLevel),
		common.HexToAddress(order.Blacklist),
		fixedTag,
		order.GetBenchmarks().ToArray(),
	)

	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}

	id, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.OrderPlacedTopic)
	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}
	orderInfo, err := api.GetOrderInfo(ctx, id)
	ch <- OrderOrError{orderInfo, err}
}

func (api *BasicMarketAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) <-chan error {
	ch := make(chan error, 0)
	go api.cancelOrder(ctx, key, id, ch)
	return ch
}

func (api *BasicMarketAPI) cancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.marketContract.CancelOrder(opts, id)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.OrderUpdatedTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.Order, error) {
	order1, err := api.marketContract.GetOrderInfo(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	order2, err := api.marketContract.GetOrderParams(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	netflags := pb.NetflagsToUint(order1.Netflags)

	benchmarks, err := pb.NewBenchmarks(order1.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &pb.Order{
		Id:             pb.NewBigInt(orderID),
		DealID:         pb.NewBigInt(order2.DealID),
		OrderType:      pb.OrderType(order1.OrderType),
		OrderStatus:    pb.OrderStatus(order2.OrderStatus),
		AuthorID:       pb.NewEthAddress(order1.Author),
		CounterpartyID: pb.NewEthAddress(order1.Counterparty),
		Duration:       order1.Duration.Uint64(),
		Price:          pb.NewBigInt(order1.Price),
		Netflags:       netflags,
		IdentityLevel:  pb.IdentityLevel(order1.IdentityLevel),
		Blacklist:      order1.Blacklist.String(),
		Tag:            order1.Tag[:],
		Benchmarks:     benchmarks,
		FrozenSum:      pb.NewBigInt(order1.FrozenSum),
	}, nil
}

func (api *BasicMarketAPI) GetOrdersAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetOrdersAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) <-chan error {
	ch := make(chan error, 0)
	go api.bill(ctx, key, dealID, ch)
	return ch
}

func (api *BasicMarketAPI) bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.marketContract.Bill(opts, dealID)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.BilledTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.marketContract.RegisterWorker(opts, master)
}

func (api *BasicMarketAPI) ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.marketContract.RegisterWorker(opts, slave)
}

func (api *BasicMarketAPI) RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.marketContract.RemoveWorker(opts, master, slave)
}

func (api *BasicMarketAPI) GetMaster(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (common.Address, error) {
	return api.marketContract.GetMaster(getCallOptions(ctx), slave)
}

func (api *BasicMarketAPI) GetDealChangeRequestInfo(ctx context.Context, dealID *big.Int) (*pb.DealChangeRequest, error) {
	changeRequest, err := api.marketContract.GetChangeRequestInfo(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	return &pb.DealChangeRequest{
		DealID:      pb.NewBigInt(changeRequest.DealID),
		RequestType: pb.OrderType(changeRequest.RequestType),
		Duration:    changeRequest.Duration.Uint64(),
		Price:       pb.NewBigInt(changeRequest.Price),
		Status:      pb.ChangeRequestStatus(changeRequest.Status),
	}, nil
}

func (api *BasicMarketAPI) GetNumBenchmarks(ctx context.Context) (int, error) {
	return NumCurrentBenchmarks, nil
}

type BasicProfileRegistryAPI struct {
	client                  *ethclient.Client
	profileRegistryContract *marketAPI.ProfileRegistry
	gasPrice                int64
}

func NewBasicProfileRegistry(client *ethclient.Client, address common.Address, gasPrice int64) (ProfileRegistryAPI, error) {
	profileRegistryContract, err := marketAPI.NewProfileRegistry(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicProfileRegistryAPI{
		client:                  client,
		profileRegistryContract: profileRegistryContract,
		gasPrice:                gasPrice,
	}, nil
}

func (api *BasicProfileRegistryAPI) GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error) {
	level, err := api.profileRegistryContract.GetValidatorLevel(getCallOptions(ctx), validatorID)
	if err != nil {
		return nil, err
	}

	return &pb.Validator{
		Id:    pb.NewEthAddress(validatorID),
		Level: uint64(level),
	}, nil
}

func (api *BasicProfileRegistryAPI) GetCertificate(ctx context.Context, certificateID *big.Int) (*pb.Certificate, error) {
	validatorID, ownerID, attribute, value, err := api.profileRegistryContract.GetCertificate(getCallOptions(ctx), certificateID)
	if err != nil {
		return nil, err
	}

	return &pb.Certificate{
		ValidatorID: pb.NewEthAddress(validatorID),
		OwnerID:     pb.NewEthAddress(ownerID),
		Attribute:   attribute.Uint64(),
		Value:       value,
	}, nil
}

type BasicBlacklistAPI struct {
	client            *ethclient.Client
	blacklistContract *marketAPI.Blacklist
	gasPrice          int64
}

func NewBasicBlacklist(client *ethclient.Client, address common.Address, gasPrice int64) (BlacklistAPI, error) {
	blacklistContract, err := marketAPI.NewBlacklist(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicBlacklistAPI{
		client:            client,
		blacklistContract: blacklistContract,
		gasPrice:          gasPrice,
	}, nil
}

func (api *BasicBlacklistAPI) Check(ctx context.Context, who, whom common.Address) (bool, error) {
	return api.blacklistContract.Check(getCallOptions(ctx), who, whom)
}

func (api *BasicBlacklistAPI) Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.blacklistContract.Add(opts, who, whom)
}

func (api *BasicBlacklistAPI) Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.blacklistContract.Remove(opts, whom)
}

func (api *BasicBlacklistAPI) AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.blacklistContract.AddMaster(opts, root)
}

func (api *BasicBlacklistAPI) RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.blacklistContract.RemoveMaster(opts, root)
}

func (api *BasicBlacklistAPI) SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.blacklistContract.SetMarketAddress(opts, market)
}

type StandardTokenAPI struct {
	client        *ethclient.Client
	tokenContract *marketAPI.StandardToken
	gasPrice      int64
}

func NewStandardToken(client *ethclient.Client, address common.Address, gasPrice int64) (TokenAPI, error) {
	tokenContract, err := marketAPI.NewStandardToken(address, client)
	if err != nil {
		return nil, err
	}

	return &StandardTokenAPI{
		client:        client,
		tokenContract: tokenContract,
		gasPrice:      gasPrice,
	}, nil
}

func (api *StandardTokenAPI) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	return api.tokenContract.BalanceOf(getCallOptions(ctx), common.HexToAddress(address))
}

func (api *StandardTokenAPI) AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error) {
	return api.tokenContract.Allowance(getCallOptions(ctx), common.HexToAddress(from), common.HexToAddress(to))
}

func (api *StandardTokenAPI) Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.tokenContract.Approve(opts, common.HexToAddress(to), amount)
}

func (api *StandardTokenAPI) Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
}

func (api *StandardTokenAPI) TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
}

func (api *StandardTokenAPI) TotalSupply(ctx context.Context) (*big.Int, error) {
	return api.tokenContract.TotalSupply(getCallOptions(ctx))
}

type TestTokenApi struct {
	client        *ethclient.Client
	tokenContract *marketAPI.SNMTToken
	gasPrice      int64
}

func NewTestToken(client *ethclient.Client, address common.Address, gasPrice int64) (TestTokenAPI, error) {
	tokenContract, err := marketAPI.NewSNMTToken(address, client)
	if err != nil {
		return nil, err
	}

	return &TestTokenApi{
		client:        client,
		tokenContract: tokenContract,
		gasPrice:      gasPrice,
	}, nil
}

func (api *TestTokenApi) GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.tokenContract.GetTokens(opts)
}

type BasicEventsAPI struct {
	client      *ethclient.Client
	logger      *zap.Logger
	marketABI   abi.ABI
	profilesABI abi.ABI
}

func NewEventsAPI(client *ethclient.Client, logger *zap.Logger) (EventsAPI, error) {
	marketABI, err := abi.JSON(strings.NewReader(marketAPI.MarketABI))
	if err != nil {
		return nil, err
	}

	profilesABI, err := abi.JSON(strings.NewReader(marketAPI.ProfileRegistryABI))
	if err != nil {
		return nil, err
	}

	return &BasicEventsAPI{
		client:      client,
		logger:      logger,
		marketABI:   marketABI,
		profilesABI: profilesABI,
	}, nil
}

func (api *BasicEventsAPI) GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error) {
	var (
		topics     [][]common.Hash
		eventTopic = []common.Hash{
			market.DealOpenedTopic,
			market.DealUpdatedTopic,
			market.OrderPlacedTopic,
			market.OrderUpdatedTopic,
			market.DealChangeRequestSentTopic,
			market.DealChangeRequestUpdatedTopic,
			market.BilledTopic,
			market.WorkerAnnouncedTopic,
			market.WorkerConfirmedTopic,
			market.WorkerConfirmedTopic,
			market.WorkerRemovedTopic,
			market.AddedToBlacklistTopic,
			market.RemovedFromBlacklistTopic,
			market.ValidatorCreatedTopic,
			market.ValidatorDeletedTopic,
			market.CertificateCreatedTopic,
		}
		out = make(chan *Event, 128)
	)
	topics = append(topics, eventTopic)

	go func() {
		var (
			lastLogBlockNumber = fromBlockInitial.Uint64()
			fromBlock          = fromBlockInitial.Uint64()
			tk                 = time.NewTicker(time.Second)
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				logs, err := api.client.FilterLogs(ctx, ethereum.FilterQuery{
					Topics:    topics,
					FromBlock: big.NewInt(0).SetUint64(fromBlock),
					Addresses: []common.Address{
						market.MarketAddr(),
						market.BlacklistAddr(),
						market.ProfileRegistryAddr(),
					},
				})

				if err != nil {
					out <- &Event{
						Data:        &ErrorData{Err: errors.Wrap(err, "failed to FilterLogs")},
						BlockNumber: fromBlock,
					}
				}

				numLogs := len(logs)
				if numLogs < 1 {
					api.logger.Info("no logs, skipping")
					continue
				}

				var eventTS uint64
				for _, log := range logs {
					// Skip logs from the last seen block.
					if log.BlockNumber == fromBlock {
						continue
					}
					// Update eventTS if we've got a new block.
					if lastLogBlockNumber != log.BlockNumber {
						lastLogBlockNumber = log.BlockNumber
						block, err := api.client.BlockByNumber(ctx, big.NewInt(0).SetUint64(lastLogBlockNumber))
						if err != nil {
							api.logger.Warn("failed to get event timestamp", zap.Error(err),
								zap.Uint64("blockNumber", lastLogBlockNumber))
						} else {
							eventTS = block.Time().Uint64()
						}
					}
					api.processLog(log, eventTS, out)
				}

				fromBlock = logs[numLogs-1].BlockNumber
			}
		}
	}()

	return out, nil
}

func (api *BasicEventsAPI) processLog(log types.Log, eventTS uint64, out chan *Event) {
	// This should never happen, but it's ethereum, and things might happen.
	if len(log.Topics) < 1 {
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("malformed log entry"), Topic: "unknown"},
			BlockNumber: log.BlockNumber,
		}
		return
	}

	sendErr := func(out chan *Event, err error, topic common.Hash) {
		out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber, TS: eventTS}
	}

	sendData := func(data interface{}) {
		out <- &Event{Data: data, BlockNumber: log.BlockNumber, TS: eventTS}
	}

	var topic = log.Topics[0]
	switch topic {
	case market.DealOpenedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealOpenedData{ID: id})
	case market.DealUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealUpdatedData{ID: id})
	case market.DealChangeRequestSentTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestSentData{ID: id})
	case market.DealChangeRequestUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestUpdatedData{ID: id})
	case market.BilledTopic:
		var billedData = &BilledData{}
		if err := api.marketABI.Unpack(billedData, "Billed", log.Data); err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(billedData)
	case market.OrderPlacedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderPlacedData{ID: id})
	case market.OrderUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderUpdatedData{ID: id})
	case market.WorkerAnnouncedTopic:
		slaveID, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerAnnouncedData{SlaveID: slaveID, MasterID: masterID})
	case market.WorkerConfirmedTopic:
		slaveID, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerConfirmedData{SlaveID: slaveID, MasterID: masterID})
	case market.WorkerRemovedTopic:
		slaveID, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerRemovedData{SlaveID: slaveID, MasterID: masterID})
	case market.AddedToBlacklistTopic:
		adderID, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		addeeID, err := extractAddress(log, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&AddedToBlacklistData{AdderID: adderID, AddeeID: addeeID})
	case market.RemovedFromBlacklistTopic:
		removerID, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		removeeID, err := extractAddress(log, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&RemovedFromBlacklistData{RemoverID: removerID, RemoveeID: removeeID})
	case market.ValidatorCreatedTopic:
		id, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorCreatedData{ID: id})
	case market.ValidatorDeletedTopic:
		id, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorDeletedData{ID: id})
	case market.CertificateCreatedTopic:
		var id = big.NewInt(0)
		if err := api.profilesABI.Unpack(&id, "CertificateCreated", log.Data); err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&CertificateCreatedData{ID: id})
	default:
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("unknown topic"), Topic: topic.String()},
			BlockNumber: log.BlockNumber,
		}
	}
}

type BasicSimpleGatekeeperAPI struct {
	client             *ethclient.Client
	gatekeeperContract *marketAPI.Gatekeeper
	gasPrice           int64
	logParsePeriod     time.Duration
}

func NewBasicSimpleGatekeeperAPI(client *ethclient.Client, address common.Address, gasPrice int64, logParsePeriod time.Duration) (SimpleGatekeeperAPI, error) {
	gatekeeperContract, err := marketAPI.NewGatekeeper(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicSimpleGatekeeperAPI{
		gatekeeperContract: gatekeeperContract,
		gasPrice:           gasPrice,
		client:             client,
		logParsePeriod:     logParsePeriod,
	}, nil
}

func (api *BasicSimpleGatekeeperAPI) PayIn(ctx context.Context, key *ecdsa.PrivateKey, value *big.Int) <-chan error {
	ch := make(chan error, 0)
	go api.payin(ctx, key, value, ch)
	return ch
}

func (api *BasicSimpleGatekeeperAPI) payin(ctx context.Context, key *ecdsa.PrivateKey, value *big.Int, ch chan<- error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.gatekeeperContract.PayIn(opts, value)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.PayInTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicSimpleGatekeeperAPI) Payout(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, value, txNumber *big.Int) <-chan error {
	ch := make(chan error, 0)
	go api.payout(ctx, key, to, value, txNumber, ch)
	return ch
}

func (api *BasicSimpleGatekeeperAPI) payout(ctx context.Context, key *ecdsa.PrivateKey, to common.Address, value, txNumber *big.Int, ch chan<- error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	tx, err := api.gatekeeperContract.Payout(opts, to, value, txNumber)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, market.PayOutTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

type BasicOracleUSDAPI struct {
	client         *ethclient.Client
	oracleContract *marketAPI.OracleUSD
	gasPrice       int64
	logParsePeriod time.Duration
}

func NewBasicOracleUSDAPI(client *ethclient.Client, address common.Address, gasPrice int64) (OracleUSDAPI, error) {
	oracleContract, err := marketAPI.NewOracleUSD(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicOracleUSDAPI{
		oracleContract: oracleContract,
		gasPrice:       gasPrice,
		client:         client,
	}, nil
}

func (api *BasicOracleUSDAPI) SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.gasPrice)
	return api.oracleContract.SetCurrentPrice(opts, price)
}

func (api *BasicOracleUSDAPI) GetCurrentPrice(ctx context.Context) (*big.Int, error) {
	return api.oracleContract.GetCurrentPrice(getCallOptions(ctx))
}
