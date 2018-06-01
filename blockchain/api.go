package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	marketAPI "github.com/sonm-io/core/blockchain/source/api"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type API interface {
	ProfileRegistry() ProfileRegistryAPI
	Events() EventsAPI
	Market() MarketAPI
	Blacklist() BlacklistAPI
	LiveToken() TokenAPI
	SideToken() TokenAPI
	TestToken() TestTokenAPI
	OracleUSD() OracleAPI
}

type ProfileRegistryAPI interface {
	GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error)
	GetCertificate(ctx context.Context, certificateID *big.Int) (*pb.Certificate, error)
}

type EventsAPI interface {
	GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error)
}

type MarketAPI interface {
	QuickBuy(ctx context.Context, key *ecdsa.PrivateKey, askId *big.Int) (*types.Transaction, error)
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) <-chan DealOrError
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) <-chan error
	GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.Deal, error)
	GetDealsAmount(ctx context.Context) (*big.Int, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) <-chan OrderOrError
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) <-chan error
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.Order, error)
	GetOrdersAmount(ctx context.Context) (*big.Int, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) <-chan error
	RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) <-chan error
	ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) <-chan error
	RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) <-chan error
	GetMaster(ctx context.Context, slave common.Address) (common.Address, error)
	GetDealChangeRequestInfo(ctx context.Context, id *big.Int) (*pb.DealChangeRequest, error)
	CreateChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, request *pb.DealChangeRequest) (*big.Int, error)
	CancelChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error
	GetNumBenchmarks(ctx context.Context) (uint64, error)
}

type BlacklistAPI interface {
	Check(ctx context.Context, who, whom common.Address) (bool, error)
	Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error)
	Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) error
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

// OracleAPI manage price relation between some currency and SNM token
type OracleAPI interface {
	// SetCurrentPrice sets current price relation between some currency and SONM token
	SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error)
	// GetCurrentPrice returns current price relation between some currency and SONM token
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
	oracle          OracleAPI
}

func NewAPI(opts ...Option) (API, error) {
	defaults := defaultOptions()
	for _, o := range opts {
		o(defaults)
	}

	liveToken, err := NewStandardToken(SNMAddr(), defaults.livechain)
	if err != nil {
		return nil, err
	}

	testToken, err := NewTestToken(SNMAddr(), defaults.livechain)
	if err != nil {
		return nil, err
	}

	blacklist, err := NewBasicBlacklist(BlacklistAddr(), defaults.sidechain)
	if err != nil {
		return nil, err
	}

	marketApi, err := NewBasicMarket(MarketAddr(), defaults.sidechain)
	if err != nil {
		return nil, err
	}

	profileRegistry, err := NewProfileRegistry(ProfileRegistryAddr(), defaults.sidechain)
	if err != nil {
		return nil, err
	}

	sideToken, err := NewStandardToken(SNMSidechainAddr(), defaults.sidechain)
	if err != nil {
		return nil, err
	}

	// fixme: wtf? context.Background for logger?
	events, err := NewEventsAPI(defaults.sidechain, ctxlog.GetLogger(context.Background()))
	if err != nil {
		return nil, err
	}

	oracle, err := NewOracleUSDAPI(OracleUsdAddr(), defaults.sidechain)
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

func (api *BasicAPI) OracleUSD() OracleAPI {
	return api.oracle
}

type BasicMarketAPI struct {
	client         CustomEthereumClient
	marketContract *marketAPI.Market
	opts           *chainOpts
}

func NewBasicMarket(address common.Address, opts *chainOpts) (MarketAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	marketContract, err := marketAPI.NewMarket(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicMarketAPI{
		client:         client,
		marketContract: marketContract,
		opts:           opts,
	}, nil
}

func (api *BasicMarketAPI) QuickBuy(ctx context.Context, key *ecdsa.PrivateKey, askId *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.marketContract.QuickBuy(opts, askId)
}

func (api *BasicMarketAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) <-chan DealOrError {
	ch := make(chan DealOrError, 0)
	go api.openDeal(ctx, key, askID, bidID, ch)
	return ch
}

func (api *BasicMarketAPI) openDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int, ch chan DealOrError) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.OpenDeal(opts, askID, bidID)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	receipt, err := WaitTransactionReceipt(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	logs, err := FindLogByTopic(receipt, DealOpenedTopic)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	id, err := extractBig(logs.Topics, 1)
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
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.CloseDeal(opts, dealID, blacklisted)
	if err != nil {
		ch <- err
		return
	}

	_, err = waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, DealUpdatedTopic)
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

	noAsk := deal1.AskID.Cmp(big.NewInt(0)) == 0
	noBid := deal1.BidID.Cmp(big.NewInt(0)) == 0
	if noAsk && noBid {
		return nil, fmt.Errorf("no deal with id = %s", dealID.String())
	}

	deal2, err := api.marketContract.GetDealParams(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}
	if deal2.Status == 0 {
		return nil, fmt.Errorf("deal fetching inconsistency for deal %s", dealID.String())
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
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)

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

	receipt, err := WaitTransactionReceipt(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}

	logs, err := FindLogByTopic(receipt, OrderPlacedTopic)
	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}

	id, err := extractBig(logs.Topics, 1)
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
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.CancelOrder(opts, id)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, OrderUpdatedTopic); err != nil {
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

	noAuthor := order1.Author.Big().Cmp(big.NewInt(0)) == 0
	noType := pb.OrderType(order1.OrderType) == pb.OrderType_ANY

	if noAuthor && noType {
		return nil, fmt.Errorf("no order with id = %s", orderID.String())
	}

	order2, err := api.marketContract.GetOrderParams(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}
	if order2.OrderStatus == 0 {
		return nil, fmt.Errorf("order fetching inconsistency for order %s", orderID.String())
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
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.Bill(opts, dealID)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, BilledTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) <-chan error {
	ch := make(chan error, 0)
	go api.registerWorker(ctx, key, master, ch)
	return ch
}

func (api *BasicMarketAPI) registerWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.RegisterWorker(opts, master)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, WorkerAnnouncedTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) <-chan error {
	ch := make(chan error, 0)
	go api.confirmWorker(ctx, key, slave, ch)
	return ch
}

func (api *BasicMarketAPI) confirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.ConfirmWorker(opts, slave)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, WorkerConfirmedTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) <-chan error {
	ch := make(chan error, 0)
	go api.removeWorker(ctx, key, master, slave, ch)
	return ch
}

func (api *BasicMarketAPI) removeWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.RemoveWorker(opts, master, slave)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, WorkerRemovedTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicMarketAPI) GetMaster(ctx context.Context, slave common.Address) (common.Address, error) {
	return api.marketContract.GetMaster(getCallOptions(ctx), slave)
}

func (api *BasicMarketAPI) GetDealChangeRequestInfo(ctx context.Context, changeRequestID *big.Int) (*pb.DealChangeRequest, error) {
	changeRequest, err := api.marketContract.GetChangeRequestInfo(getCallOptions(ctx), changeRequestID)
	if err != nil {
		return nil, err
	}

	return &pb.DealChangeRequest{
		Id:          pb.NewBigInt(changeRequestID),
		DealID:      pb.NewBigInt(changeRequest.DealID),
		RequestType: pb.OrderType(changeRequest.RequestType),
		Duration:    changeRequest.Duration.Uint64(),
		Price:       pb.NewBigInt(changeRequest.Price),
		Status:      pb.ChangeRequestStatus(changeRequest.Status),
	}, nil
}

func (api *BasicMarketAPI) CreateChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, req *pb.DealChangeRequest) (*big.Int, error) {
	duration := big.NewInt(int64(req.GetDuration()))
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.CreateChangeRequest(opts, req.GetDealID().Unwrap(), req.GetPrice().Unwrap(), duration)
	if err != nil {
		return nil, err
	}

	log, err := waitForTransactionResult(ctx, api.client, api.opts.logParsePeriod, tx, DealChangeRequestSentTopic)
	if err != nil {
		return nil, err
	}

	id, err := extractBig(log.Topics, 1)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot extract change request id from transaction logs")
	}

	return id, nil
}

func (api *BasicMarketAPI) CancelChangeRequest(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) error {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.marketContract.CancelChangeRequest(opts, id)
	if err != nil {
		return err
	}

	_, err = WaitTransactionReceipt(ctx, api.client, defaultBlockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		return err
	}

	return nil
}

func (api *BasicMarketAPI) GetNumBenchmarks(ctx context.Context) (uint64, error) {
	num, err := api.marketContract.GetBenchmarksQuantity(getCallOptions(ctx))
	if err != nil {
		return 0, err
	}
	if !num.IsUint64() {
		return 0, errors.New("benchmarks quantity overflows int64")
	}
	return num.Uint64(), nil
}

type ProfileRegistry struct {
	client                  CustomEthereumClient
	profileRegistryContract *marketAPI.ProfileRegistry
	opts                    *chainOpts
}

func NewProfileRegistry(address common.Address, opts *chainOpts) (ProfileRegistryAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	profileRegistryContract, err := marketAPI.NewProfileRegistry(address, client)
	if err != nil {
		return nil, err
	}

	return &ProfileRegistry{
		client:                  client,
		profileRegistryContract: profileRegistryContract,
		opts: opts,
	}, nil
}

func (api *ProfileRegistry) GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error) {
	level, err := api.profileRegistryContract.GetValidatorLevel(getCallOptions(ctx), validatorID)
	if err != nil {
		return nil, err
	}

	return &pb.Validator{
		Id:    pb.NewEthAddress(validatorID),
		Level: uint64(level),
	}, nil
}

func (api *ProfileRegistry) GetCertificate(ctx context.Context, certificateID *big.Int) (*pb.Certificate, error) {
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
	client            CustomEthereumClient
	blacklistContract *marketAPI.Blacklist
	opts              *chainOpts
}

func NewBasicBlacklist(address common.Address, opts *chainOpts) (BlacklistAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	blacklistContract, err := marketAPI.NewBlacklist(address, client)
	if err != nil {
		return nil, err
	}

	return &BasicBlacklistAPI{
		client:            client,
		blacklistContract: blacklistContract,
		opts:              opts,
	}, nil
}

func (api *BasicBlacklistAPI) Check(ctx context.Context, who, whom common.Address) (bool, error) {
	return api.blacklistContract.Check(getCallOptions(ctx), who, whom)
}

func (api *BasicBlacklistAPI) Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.blacklistContract.Add(opts, who, whom)
}

func (api *BasicBlacklistAPI) Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) error {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	tx, err := api.blacklistContract.Remove(opts, whom)
	if err != nil {
		return err
	}

	rec, err := WaitTransactionReceipt(ctx, api.client, api.opts.blockConfirmations, api.opts.logParsePeriod, tx)
	if err != nil {
		return err
	}

	if _, err := FindLogByTopic(rec, RemovedFromBlacklistTopic); err != nil {
		return err
	}

	return nil
}

func (api *BasicBlacklistAPI) AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.blacklistContract.AddMaster(opts, root)
}

func (api *BasicBlacklistAPI) RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.blacklistContract.RemoveMaster(opts, root)
}

func (api *BasicBlacklistAPI) SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.blacklistContract.SetMarketAddress(opts, market)
}

type StandardTokenApi struct {
	client        CustomEthereumClient
	tokenContract *marketAPI.StandardToken
	opts          *chainOpts
}

func NewStandardToken(address common.Address, opts *chainOpts) (TokenAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewStandardToken(address, client)
	if err != nil {
		return nil, err
	}

	return &StandardTokenApi{
		client:        client,
		tokenContract: tokenContract,
		opts:          opts,
	}, nil
}

func (api *StandardTokenApi) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	return api.tokenContract.BalanceOf(getCallOptions(ctx), common.HexToAddress(address))
}

func (api *StandardTokenApi) AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error) {
	return api.tokenContract.Allowance(getCallOptions(ctx), common.HexToAddress(from), common.HexToAddress(to))
}

func (api *StandardTokenApi) Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.opts.gasPrice)
	return api.tokenContract.Approve(opts, common.HexToAddress(to), amount)
}

func (api *StandardTokenApi) Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.opts.gasPrice)
	return api.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
}

func (api *StandardTokenApi) TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.opts.gasPrice)
	return api.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
}

func (api *StandardTokenApi) TotalSupply(ctx context.Context) (*big.Int, error) {
	return api.tokenContract.TotalSupply(getCallOptions(ctx))
}

type TestTokenApi struct {
	client        CustomEthereumClient
	tokenContract *marketAPI.SNMTToken
	opts          *chainOpts
}

func NewTestToken(address common.Address, opts *chainOpts) (TestTokenAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewSNMTToken(address, client)
	if err != nil {
		return nil, err
	}

	return &TestTokenApi{
		client:        client,
		tokenContract: tokenContract,
		opts:          opts,
	}, nil
}

func (api *TestTokenApi) GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.opts.gasPrice)
	return api.tokenContract.GetTokens(opts)
}

type BasicEventsAPI struct {
	client CustomEthereumClient
	logger *zap.Logger
}

func NewEventsAPI(opts *chainOpts, logger *zap.Logger) (EventsAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	return &BasicEventsAPI{
		client: client,
		logger: logger,
	}, nil
}

func (api *BasicEventsAPI) GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error) {
	var (
		topics     [][]common.Hash
		eventTopic = []common.Hash{
			DealOpenedTopic,
			DealUpdatedTopic,
			OrderPlacedTopic,
			OrderUpdatedTopic,
			DealChangeRequestSentTopic,
			DealChangeRequestUpdatedTopic,
			BilledTopic,
			WorkerAnnouncedTopic,
			WorkerConfirmedTopic,
			WorkerConfirmedTopic,
			WorkerRemovedTopic,
			AddedToBlacklistTopic,
			RemovedFromBlacklistTopic,
			ValidatorCreatedTopic,
			ValidatorDeletedTopic,
			CertificateCreatedTopic,
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
						MarketAddr(),
						BlacklistAddr(),
						ProfileRegistryAddr(),
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
	case DealOpenedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealOpenedData{ID: id})
	case DealUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealUpdatedData{ID: id})
	case DealChangeRequestSentTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestSentData{ID: id})
	case DealChangeRequestUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestUpdatedData{ID: id})
	case BilledTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		paidAmount, err := extractBig(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&BilledData{DealID: id, PaidAmount: paidAmount})
	case OrderPlacedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderPlacedData{ID: id})
	case OrderUpdatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderUpdatedData{ID: id})
	case WorkerAnnouncedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerAnnouncedData{WorkerID: slaveID, MasterID: masterID})
	case WorkerConfirmedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerConfirmedData{WorkerID: slaveID, MasterID: masterID})
	case WorkerRemovedTopic:
		slaveID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		masterID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&WorkerRemovedData{WorkerID: slaveID, MasterID: masterID})
	case AddedToBlacklistTopic:
		adderID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		addeeID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&AddedToBlacklistData{AdderID: adderID, AddeeID: addeeID})
	case RemovedFromBlacklistTopic:
		removerID, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		removeeID, err := extractAddress(log.Topics, 2)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&RemovedFromBlacklistData{RemoverID: removerID, RemoveeID: removeeID})
	case ValidatorCreatedTopic:
		id, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorCreatedData{ID: id})
	case ValidatorDeletedTopic:
		id, err := extractAddress(log.Topics, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorDeletedData{ID: id})
	case CertificateCreatedTopic:
		id, err := extractBig(log.Topics, 1)
		if err != nil {
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

type OracleUSDAPI struct {
	client         CustomEthereumClient
	oracleContract *marketAPI.OracleUSD
	opts           *chainOpts
}

func NewOracleUSDAPI(address common.Address, opts *chainOpts) (OracleAPI, error) {
	client, err := opts.getClient()
	if err != nil {
		return nil, err
	}

	oracleContract, err := marketAPI.NewOracleUSD(address, client)
	if err != nil {
		return nil, err
	}

	return &OracleUSDAPI{
		client:         client,
		oracleContract: oracleContract,
		opts:           opts,
	}, nil

}

func (api *OracleUSDAPI) SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, api.opts.gasPrice)
	return api.oracleContract.SetCurrentPrice(opts, price)
}

func (api *OracleUSDAPI) GetCurrentPrice(ctx context.Context) (*big.Int, error) {
	return api.oracleContract.GetCurrentPrice(getCallOptions(ctx))
}
