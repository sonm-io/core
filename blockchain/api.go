package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strings"
	"time"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/sonm-io/core/blockchain/market"
	marketAPI "github.com/sonm-io/core/blockchain/market/api"
)

type API interface {
	CertsAPI
	MarketAPI
	BlacklistAPI
	TokenAPI
	GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts
}

type CertsAPI interface{}

type MarketAPI interface {
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) (*types.Transaction, error)
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) (*types.Transaction, error)
	GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.MarketDeal, error)
	GetDealsAmount(ctx context.Context) (*big.Int, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.MarketOrder) (*types.Transaction, error)
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.MarketOrder, error)
	GetOrdersAmount(ctx context.Context) (*big.Int, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) (*types.Transaction, error)
	RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (*types.Transaction, error)
	ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (*types.Transaction, error)
	RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (*types.Transaction, error)
	GetMaster(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (common.Address, error)
	GetMarketEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error)
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
	// GetTokens - send 100 SNMT token for message caller
	// this function added for MVP purposes and has been deleted later
	GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error)
}

type BasicAPI struct {
	client            *ethclient.Client
	gasPrice          int64
	marketContract    *marketAPI.Market
	blacklistContract *marketAPI.Blacklist
	tokenContract     *marketAPI.SNMTToken
	logger            *zap.Logger
}

func NewAPI(ethEndpoint *string, gasPrice *int64) (API, error) {
	client, err := initEthClient(ethEndpoint)
	if err != nil {
		return nil, err
	}

	if gasPrice == nil {
		var p int64 = defaultGasPrice
		gasPrice = &p
	}

	blacklistContract, err := marketAPI.NewBlacklist(common.HexToAddress(market.BlacklistAddress), client)
	if err != nil {
		return nil, err
	}

	marketContract, err := marketAPI.NewMarket(common.HexToAddress(market.MarketAddress), client)
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewSNMTToken(common.HexToAddress(market.SNMTAddress), client)
	if err != nil {
		return nil, err
	}

	api := &BasicAPI{
		client:            client,
		gasPrice:          *gasPrice,
		marketContract:    marketContract,
		blacklistContract: blacklistContract,
		tokenContract:     tokenContract,
	}

	return api, nil
}

func (api *BasicAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.OpenDeal(opts, askID, bigID)
}

func (api *BasicAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.CloseDeal(opts, dealID, blacklisted)
}

func (api *BasicAPI) GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.MarketDeal, error) {

	deal1, err := api.marketContract.GetDealInfo(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	deal2, err := api.marketContract.GetDealParams(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	var benchmarks = make([]uint64, len(deal1.Benchmarks))
	for idx, benchmark := range deal1.Benchmarks {
		benchmarks[idx] = benchmark.Uint64()
	}

	return &pb.MarketDeal{
		Id:             dealID.String(),
		Benchmarks:     benchmarks,
		SupplierID:     deal1.SupplierID.String(),
		ConsumerID:     deal1.ConsumerID.String(),
		MasterID:       deal1.MasterID.String(),
		AskID:          deal1.AskID.String(),
		BidID:          deal1.BidID.String(),
		Duration:       deal2.Duration.Uint64(),
		Price:          pb.NewBigInt(deal2.Price),
		StartTime:      &pb.Timestamp{Seconds: deal1.StartTime.Int64()},
		EndTime:        &pb.Timestamp{Seconds: deal2.EndTime.Int64()},
		Status:         pb.MarketDealStatus(deal2.Status),
		BlockedBalance: pb.NewBigInt(deal2.BlockedBalance),
		TotalPayout:    pb.NewBigInt(deal2.TotalPayout),
		LastBillTS:     &pb.Timestamp{Seconds: deal2.LastBillTS.Int64()},
	}, nil
}

func (api *BasicAPI) GetDealsAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetDealsAmount(getCallOptions(ctx))
}

func (api *BasicAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.MarketOrder) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	var bigBenchmarks = make([]*big.Int, len(order.Benchmarks))
	for idx, benchmark := range order.Benchmarks {
		bigBenchmarks[idx] = big.NewInt(int64(benchmark))
	}
	var fixedTag [32]byte
	copy(fixedTag[:], order.Tag[:])
	var fixedNetflags [3]bool
	copy(fixedNetflags[:], order.Netflags[:])
	return api.marketContract.PlaceOrder(opts,
		uint8(order.OrderType),
		common.StringToAddress(order.Counterparty),
		order.Price.Unwrap(),
		big.NewInt(int64(order.Duration)),
		fixedNetflags,
		uint8(order.IdentityLevel),
		common.StringToAddress(order.Blacklist),
		fixedTag,
		bigBenchmarks,
	)
}

func (api *BasicAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.CancelOrder(opts, id)
}

func (api *BasicAPI) GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.MarketOrder, error) {
	order1, err := api.marketContract.GetOrderInfo(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	order2, err := api.marketContract.GetOrderParams(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	var benchmarks = make([]uint64, len(order1.Benchmarks))
	for idx, benchmark := range order1.Benchmarks {
		benchmarks[idx] = benchmark.Uint64()
	}

	return &pb.MarketOrder{
		Id:            orderID.String(),
		DealID:        order2.DealID.String(),
		OrderType:     pb.MarketOrderType(order1.OrderType),
		OrderStatus:   pb.MarketOrderStatus(order2.OrderStatus),
		Author:        order1.Author.String(),
		Counterparty:  order1.Counterparty.String(),
		Price:         pb.NewBigInt(order1.Price),
		Duration:      order1.Duration.Uint64(),
		Netflags:      order1.Netflags[:],
		IdentityLevel: pb.MarketIdentityLevel(order1.IdentityLevel),
		Blacklist:     order1.Blacklist.String(),
		Tag:           order1.Tag[:],
		Benchmarks:    benchmarks,
		FrozenSum:     pb.NewBigInt(order1.FrozenSum),
	}, nil
}

func (api *BasicAPI) GetOrdersAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetOrdersAmount(getCallOptions(ctx))
}

func (api *BasicAPI) Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.Bill(opts, dealID)
}

func (api *BasicAPI) RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.RegisterWorker(opts, master)
}

func (api *BasicAPI) ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.RegisterWorker(opts, slave)
}

func (api *BasicAPI) RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.RemoveWorker(opts, master, slave)
}

func (api *BasicAPI) GetMaster(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (
	common.Address, error) {
	return api.marketContract.GetMaster(getCallOptions(ctx), slave)
}

func (api *BasicAPI) GetMarketEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error) {
	var (
		topics     [][]common.Hash
		eventTopic = []common.Hash{
			DealOpenedTopic,
			DealUpdatedTopic,
			OrderPlacedTopic,
			OrderUpdatedTopic,
			DealChangeRequestSent,
			DealChangeRequestUpdated,
			WorkerAnnouncedTopic,
			WorkerConfirmedTopic,
			WorkerConfirmedTopic,
			WorkerRemovedTopic}
		out = make(chan *Event, 128)
	)
	topics = append(topics, eventTopic)

	marketABI, err := abi.JSON(strings.NewReader(string(marketAPI.MarketABI)))
	if err != nil {
		close(out)
		return nil, err
	}

	go func() {
		var (
			fromBlock = fromBlockInitial
			tk        = time.NewTicker(time.Second)
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				logs, err := api.client.FilterLogs(ctx, ethereum.FilterQuery{
					Topics:    topics,
					FromBlock: fromBlock,
				})
				if err != nil {
					out <- &Event{Data: err, BlockNumber: fromBlock.Uint64()}
				}

				fromBlock = big.NewInt(int64(logs[len(logs)-1].BlockNumber))

				for _, log := range logs {
					api.processLog(log, marketABI, out)
				}
			}
		}
	}()

	return out, nil
}

func (api *BasicAPI) processLog(log types.Log, marketABI abi.ABI, out chan *Event) {
	// This should never happen, but it's ethereum, and things might happen.
	if len(log.Topics) < 1 {
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("malformed log entry"), Topic: "unknown"},
			BlockNumber: log.BlockNumber,
		}
		return
	}

	var topic = log.Topics[0]
	switch topic {
	case DealOpenedTopic:
		var dealOpenedData = &DealOpenedData{}
		if err := marketABI.Unpack(&dealOpenedData, "DealOpened", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: dealOpenedData, BlockNumber: log.BlockNumber}
		}
	case DealUpdatedTopic:
		var dealUpdatedData = &DealUpdatedData{}
		if err := marketABI.Unpack(&dealUpdatedData, "DealUpdated", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: dealUpdatedData, BlockNumber: log.BlockNumber}
		}
	case OrderPlacedTopic:
		var orderPlacedData = &OrderPlacedData{}
		if err := marketABI.Unpack(&orderPlacedData, "OrderPlaced", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: orderPlacedData, BlockNumber: log.BlockNumber}
		}
	case OrderUpdatedTopic:
		var orderUpdatedData = &OrderUpdatedData{}
		if err := marketABI.Unpack(&orderUpdatedData, "OrderUpdated", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: orderUpdatedData, BlockNumber: log.BlockNumber}
		}
	case DealChangeRequestSent:
		var dealChangeRequestSent = &DealChangeRequestSentData{}
		if err := marketABI.Unpack(&dealChangeRequestSent, "DealChangeRequestSent", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: dealChangeRequestSent, BlockNumber: log.BlockNumber}
		}
	case DealChangeRequestUpdated:
		var dealChangeRequestUpdated = &DealChangeRequestUpdatedData{}
		if err := marketABI.Unpack(&dealChangeRequestUpdated, "DealChangeRequestUpdated", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: dealChangeRequestUpdated, BlockNumber: log.BlockNumber}
		}
	case WorkerAnnouncedTopic:
		var workerAnnouncedData = &WorkerAnnouncedData{}
		if err := marketABI.Unpack(&workerAnnouncedData, "WorkerAnnounced", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: workerAnnouncedData, BlockNumber: log.BlockNumber}
		}
	case WorkerConfirmedTopic:
		var workerConfirmedData = &WorkerConfirmedData{}
		if err := marketABI.Unpack(&workerConfirmedData, "WorkerConfirmed", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: workerConfirmedData, BlockNumber: log.BlockNumber}
		}
	case WorkerRemovedTopic:
		var workerRemovedData = &WorkerRemovedData{}
		if err := marketABI.Unpack(&workerRemovedData, "WorkerRemoved", log.Data); err != nil {
			out <- &Event{Data: &ErrorData{Err: err, Topic: topic.String()}, BlockNumber: log.BlockNumber}
		} else {
			out <- &Event{Data: workerRemovedData, BlockNumber: log.BlockNumber}
		}
	default:
		out <- &Event{
			Data:        &ErrorData{Err: errors.New("unknown topic"), Topic: topic.String()},
			BlockNumber: log.BlockNumber,
		}
	}
}

func (api *BasicAPI) Check(ctx context.Context, who, whom common.Address) (bool, error) {
	return api.blacklistContract.Check(getCallOptions(ctx), who, whom)
}

func (api *BasicAPI) Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.Add(opts, who, whom)
}

func (api *BasicAPI) Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.Remove(opts, whom)
}

func (api *BasicAPI) AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.AddMaster(opts, root)
}

func (api *BasicAPI) RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.RemoveMaster(opts, root)
}

func (api *BasicAPI) SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.SetMarketAddress(opts, market)
}

func (api *BasicAPI) GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	opts.GasLimit = big.NewInt(gasLimit)
	opts.GasPrice = big.NewInt(api.gasPrice)
	return opts
}

func (api *BasicAPI) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	balance, err := api.tokenContract.BalanceOf(getCallOptions(ctx), common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (api *BasicAPI) AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error) {
	allowance, err := api.tokenContract.Allowance(getCallOptions(ctx), common.HexToAddress(from), common.HexToAddress(to))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (api *BasicAPI) Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, 50000)

	tx, err := api.tokenContract.Approve(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (api *BasicAPI) Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, 50000)

	tx, err := api.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (api *BasicAPI) TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, 50000)

	tx, err := api.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (api *BasicAPI) TotalSupply(ctx context.Context) (*big.Int, error) {
	supply, err := api.tokenContract.TotalSupply(getCallOptions(ctx))
	if err != nil {
		return nil, err
	}
	return supply, nil
}

func (api *BasicAPI) GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, 50000)

	tx, err := api.tokenContract.GetTokens(opts)
	if err != nil {
		return nil, err
	}
	return tx, err
}

type Event struct {
	Data        interface{}
	BlockNumber uint64
}

type DealOpenedData struct {
	ID *big.Int
}

type DealUpdatedData struct {
	ID *big.Int
}

type OrderPlacedData struct {
	ID *big.Int
}

type OrderUpdatedData struct {
	ID *big.Int
}

type DealChangeRequestSentData struct {
	ID *big.Int
}

type DealChangeRequestUpdatedData struct {
	ID *big.Int
}

type WorkerAnnouncedData struct {
	Slave  common.Address
	Master common.Address
}

type WorkerConfirmedData struct {
	Slave  common.Address
	Master common.Address
}

type WorkerRemovedData struct {
	Slave  common.Address
	Master common.Address
}

type ErrorData struct {
	Err   error
	Topic string
}
