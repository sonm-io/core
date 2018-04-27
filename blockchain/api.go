package blockchain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain/market"
	marketAPI "github.com/sonm-io/core/blockchain/market/api"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type API interface {
	CertsAPI
	EventsAPI
	MarketAPI
	BlacklistAPI
	TokenAPI
	GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit uint64) *bind.TransactOpts
}

type CertsAPI interface {
	GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error)
	GetCertificate(ctx context.Context, certificateID *big.Int) (*pb.Certificate, error)
}

type EventsAPI interface {
	GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error)
}

type DealOrError struct {
	Deal *pb.Deal
	Err  error
}

type OrderOrError struct {
	Order *pb.Order
	Err   error
}

type MarketAPI interface {
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bigID *big.Int) chan DealOrError
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) (*types.Transaction, error)
	GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.Deal, error)
	GetDealsAmount(ctx context.Context) (*big.Int, error)
	PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) chan OrderOrError
	CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) chan error
	GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.Order, error)
	GetOrdersAmount(ctx context.Context) (*big.Int, error)
	Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) (*types.Transaction, error)
	RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (*types.Transaction, error)
	ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (*types.Transaction, error)
	RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (*types.Transaction, error)
	GetMaster(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (common.Address, error)
	GetDealChangeRequestInfo(ctx context.Context, dealID *big.Int) (*pb.DealChangeRequest, error)
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
	profilesContract  *marketAPI.ProfileRegistry
	tokenContract     *marketAPI.SNMTToken
	marketABI         abi.ABI
	profilesABI       abi.ABI
	logger            *zap.Logger
	logParsePeriod    time.Duration
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

	blacklistContract, err := marketAPI.NewBlacklist(common.HexToAddress(market.BlacklistAddress), client)
	if err != nil {
		return nil, err
	}

	marketContract, err := marketAPI.NewMarket(common.HexToAddress(market.MarketAddress), client)
	if err != nil {
		return nil, err
	}

	profilesContract, err := marketAPI.NewProfileRegistry(common.HexToAddress(market.ProfileRegistryAddress), client)
	if err != nil {
		return nil, err
	}

	tokenContract, err := marketAPI.NewSNMTToken(common.HexToAddress(market.SNMAddress), client)
	if err != nil {
		return nil, err
	}

	marketABI, err := abi.JSON(strings.NewReader(marketAPI.MarketABI))
	if err != nil {
		return nil, err
	}

	profilesABI, err := abi.JSON(strings.NewReader(marketAPI.ProfileRegistryABI))
	if err != nil {
		return nil, err
	}

	api := &BasicAPI{
		client:            client,
		gasPrice:          defaults.gasPrice,
		marketContract:    marketContract,
		blacklistContract: blacklistContract,
		profilesContract:  profilesContract,
		tokenContract:     tokenContract,
		marketABI:         marketABI,
		profilesABI:       profilesABI,
		logParsePeriod:    defaults.logParsePeriod,
		logger:            ctxlog.GetLogger(context.Background()),
	}

	return api, nil
}

func (api *BasicAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) chan DealOrError {
	ch := make(chan DealOrError, 0)
	go api.openDeal(ctx, key, askID, bidID, ch)
	return ch
}

func (api *BasicAPI) openDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int, ch chan DealOrError) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	tx, err := api.marketContract.OpenDeal(opts, askID, bidID)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	id, err := api.waitForTransactionResult(ctx, tx, DealOpenedTopic)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	deal, err := api.GetDealInfo(ctx, id)
	ch <- DealOrError{deal, err}
}

func (api *BasicAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.marketContract.CloseDeal(opts, dealID, blacklisted)
}

func (api *BasicAPI) GetDealInfo(ctx context.Context, dealID *big.Int) (*pb.Deal, error) {
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
		Status:         pb.DealStatus(deal2.Status),
		BlockedBalance: pb.NewBigInt(deal2.BlockedBalance),
		TotalPayout:    pb.NewBigInt(deal2.TotalPayout),
		LastBillTS:     &pb.Timestamp{Seconds: deal2.LastBillTS.Int64()},
	}, nil
}

func (api *BasicAPI) GetDealsAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetDealsAmount(getCallOptions(ctx))
}

func (api *BasicAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) chan OrderOrError {
	ch := make(chan OrderOrError, 0)
	go api.placeOrder(ctx, key, order, ch)
	return ch
}

func (api *BasicAPI) placeOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order, ch chan OrderOrError) {
	opts := api.GetTxOpts(ctx, key, gasLimitForPlaceOrderMethod)

	fixedNetflags := pb.UintToNetflags(order.Netflags)
	var fixedTag [32]byte
	copy(fixedTag[:], order.Tag[:])

	tx, err := api.marketContract.PlaceOrder(opts,
		uint8(order.OrderType),
		common.StringToAddress(order.CounterpartyID),
		order.Price.Unwrap(),
		big.NewInt(int64(order.Duration)),
		fixedNetflags,
		uint8(order.IdentityLevel),
		common.StringToAddress(order.Blacklist),
		fixedTag,
		order.GetBenchmarks().ToArray(),
	)

	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}

	id, err := api.waitForTransactionResult(ctx, tx, OrderPlacedTopic)
	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}
	orderInfo, err := api.GetOrderInfo(ctx, id)
	ch <- OrderOrError{orderInfo, err}
}

func (api *BasicAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) chan error {
	ch := make(chan error, 0)
	go api.cancelOrder(ctx, key, id, ch)
	return ch
}

func (api *BasicAPI) cancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int, ch chan error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	tx, err := api.marketContract.CancelOrder(opts, id)
	if err != nil {
		ch <- err
		return
	}

	if _, err := api.waitForTransactionResult(ctx, tx, OrderUpdatedTopic); err != nil {
		ch <- err
		return
	}
	ch <- nil
}

func (api *BasicAPI) GetOrderInfo(ctx context.Context, orderID *big.Int) (*pb.Order, error) {
	order1, err := api.marketContract.GetOrderInfo(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	order2, err := api.marketContract.GetOrderParams(getCallOptions(ctx), orderID)
	if err != nil {
		return nil, err
	}

	netflags := pb.NetflagsToUint(order1.Netflags)

	dwhBenchmarks, err := pb.NewBenchmarks(order1.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &pb.Order{
		Id:             orderID.String(),
		DealID:         order2.DealID.String(),
		OrderType:      pb.OrderType(order1.OrderType),
		OrderStatus:    pb.OrderStatus(order2.OrderStatus),
		AuthorID:       order1.Author.String(),
		CounterpartyID: order1.Counterparty.String(),
		Duration:       order1.Duration.Uint64(),
		Price:          pb.NewBigInt(order1.Price),
		Netflags:       netflags,
		IdentityLevel:  pb.IdentityLevel(order1.IdentityLevel),
		Blacklist:      order1.Blacklist.String(),
		Tag:            order1.Tag[:],
		Benchmarks:     dwhBenchmarks,
		FrozenSum:      pb.NewBigInt(order1.FrozenSum),
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

func (api *BasicAPI) GetValidator(ctx context.Context, validatorID common.Address) (*pb.Validator, error) {
	level, err := api.profilesContract.GetValidatorLevel(getCallOptions(ctx), validatorID)
	if err != nil {
		return nil, err
	}

	return &pb.Validator{
		Id:    validatorID.String(),
		Level: uint64(level),
	}, nil
}

func (api *BasicAPI) GetCertificate(ctx context.Context, certificateID *big.Int) (
	*pb.Certificate, error) {

	validatorID, ownerID, attribute, value, err := api.profilesContract.GetCertificate(getCallOptions(ctx), certificateID)
	if err != nil {
		return nil, err
	}

	return &pb.Certificate{
		ValidatorID: validatorID.String(),
		OwnerID:     ownerID.String(),
		Attribute:   attribute.Uint64(),
		Value:       value,
	}, nil
}

func (api *BasicAPI) GetLastBlockNumber() (uint64, error) {
	p, err := api.client.SyncProgress(context.Background())

	if err != nil {
		return 0, errors.Wrap(err, "failed to GetLastBlockNumber")
	}

	if p == nil {
		return 0, errors.New("node is still syncing")
	}

	return p.CurrentBlock, nil
}

func (api *BasicAPI) GetEvents(ctx context.Context, fromBlockInitial *big.Int) (chan *Event, error) {
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
			fromBlock       = fromBlockInitial
			lastBlockNumber = fromBlockInitial
			tk              = time.NewTicker(time.Second)
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				logs, err := api.client.FilterLogs(ctx, ethereum.FilterQuery{
					Topics:    topics,
					FromBlock: fromBlock,
					Addresses: []common.Address{
						common.HexToAddress(market.MarketAddress),
						common.HexToAddress(market.BlacklistAddress),
						common.HexToAddress(market.ProfileRegistryAddress),
					},
				})

				if err != nil {
					out <- &Event{
						Data:        &ErrorData{Err: errors.Wrap(err, "failed to FilterLogs")},
						BlockNumber: fromBlock.Uint64(),
					}
				}

				numLogs := len(logs)
				if numLogs < 1 {
					api.logger.Info("no logs, skipping")
					continue
				}
				fromBlock = big.NewInt(int64(logs[numLogs-1].BlockNumber))

				var eventTS uint64
				for _, log := range logs {
					logBlockNumber := big.NewInt(int64(log.BlockNumber))
					if lastBlockNumber.Cmp(logBlockNumber) != 0 {
						lastBlockNumber = logBlockNumber
						block, err := api.client.BlockByNumber(context.Background(), lastBlockNumber)
						if err != nil {
							api.logger.Warn("failed to get event timestamp", zap.Error(err),
								zap.Uint64("blockNumber", lastBlockNumber.Uint64()))
						}
						eventTS = block.Time().Uint64()
					}
					api.processLog(log, eventTS, out)
				}
			}
		}
	}()

	return out, nil
}

func (api *BasicAPI) processLog(log types.Log, eventTS uint64, out chan *Event) {
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
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealOpenedData{ID: id})
	case DealUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealUpdatedData{ID: id})
	case DealChangeRequestSentTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestSentData{ID: id})
	case DealChangeRequestUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&DealChangeRequestUpdatedData{ID: id})
	case BilledTopic:
		var billedData = &BilledData{}
		if err := api.marketABI.Unpack(billedData, "Billed", log.Data); err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(billedData)
	case OrderPlacedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderPlacedData{ID: id})
	case OrderUpdatedTopic:
		id, err := extractBig(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&OrderUpdatedData{ID: id})
	case WorkerAnnouncedTopic:
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
	case WorkerConfirmedTopic:
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
	case WorkerRemovedTopic:
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
	case AddedToBlacklistTopic:
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
	case RemovedFromBlacklistTopic:
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
	case ValidatorCreatedTopic:
		id, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorCreatedData{ID: id})
	case ValidatorDeletedTopic:
		id, err := extractAddress(log, 1)
		if err != nil {
			sendErr(out, err, topic)
			return
		}
		sendData(&ValidatorDeletedData{ID: id})
	case CertificateCreatedTopic:
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

func (api *BasicAPI) GetDealChangeRequestInfo(ctx context.Context, dealID *big.Int) (*pb.DealChangeRequest, error) {
	changeRequest, err := api.marketContract.GetChangeRequestInfo(getCallOptions(ctx), dealID)
	if err != nil {
		return nil, err
	}

	return &pb.DealChangeRequest{
		DealID:      changeRequest.DealID.String(),
		RequestType: pb.OrderType(changeRequest.RequestType),
		Duration:    changeRequest.Duration.Uint64(),
		Price:       pb.NewBigInt(changeRequest.Price),
		Status:      pb.ChangeRequestStatus(changeRequest.Status),
	}, nil
}

func (api *BasicAPI) Check(ctx context.Context, who, whom common.Address) (bool, error) {
	return api.blacklistContract.Check(getCallOptions(ctx), who, whom)
}

func (api *BasicAPI) Add(ctx context.Context, key *ecdsa.PrivateKey, who, whom common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.Add(opts, who, whom)
}

func (api *BasicAPI) Remove(ctx context.Context, key *ecdsa.PrivateKey, whom common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.Remove(opts, whom)
}

func (api *BasicAPI) AddMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.AddMaster(opts, root)
}

func (api *BasicAPI) RemoveMaster(ctx context.Context, key *ecdsa.PrivateKey, root common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.RemoveMaster(opts, root)
}

func (api *BasicAPI) SetMarketAddress(ctx context.Context, key *ecdsa.PrivateKey, market common.Address) (
	*types.Transaction, error) {
	opts := api.GetTxOpts(ctx, key, defaultGasLimit)
	return api.blacklistContract.SetMarketAddress(opts, market)
}

func (api *BasicAPI) GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit uint64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	opts.GasLimit = gasLimit
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
	allowance, err := api.tokenContract.Allowance(getCallOptions(ctx), common.HexToAddress(from),
		common.HexToAddress(to))
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

func (api *BasicAPI) waitForTransactionResult(ctx context.Context, tx *types.Transaction, topic common.Hash) (*big.Int, error) {
	tk := util.NewImmediateTicker(api.logParsePeriod)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			id, err := api.parseTransactionLogs(ctx, tx, topic)
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}

			return id, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (api *BasicAPI) parseTransactionLogs(ctx context.Context, tx *types.Transaction, topic common.Hash) (*big.Int, error) {
	txReceipt, err := api.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	if txReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction failed")
	}

	for _, l := range txReceipt.Logs {
		if len(l.Topics) < 1 {
			return nil, errors.New("transaction topics is malformed")
		}

		receivedTopic := l.Topics[0]
		topicCmp := bytes.Compare(receivedTopic.Bytes(), topic.Bytes())
		if topicCmp == 0 && len(l.Topics) > 1 {
			return l.Topics[1].Big(), nil
		}
	}

	// TODO(sshaman1101): not so user-friendly message leaved for debugging, remove before releasing.
	return nil, fmt.Errorf("cannot find topic \"%s\"in transaction", topic.Hex())
}

type Event struct {
	Data        interface{}
	BlockNumber uint64
	TS          uint64
}

type DealOpenedData struct {
	ID *big.Int
}

type DealUpdatedData struct {
	ID *big.Int
}

type DealChangeRequestSentData struct {
	ID *big.Int
}

type DealChangeRequestUpdatedData struct {
	ID *big.Int
}

type OrderPlacedData struct {
	ID *big.Int
}

type OrderUpdatedData struct {
	ID *big.Int
}

type BilledData struct {
	DealID     *big.Int `json:"dealID"`
	PaidAmount *big.Int `json:"paidAmount"`
}

type WorkerAnnouncedData struct {
	SlaveID  common.Address
	MasterID common.Address
}

type WorkerConfirmedData struct {
	SlaveID  common.Address
	MasterID common.Address
}

type WorkerRemovedData struct {
	SlaveID  common.Address
	MasterID common.Address
}

type ErrorData struct {
	Err   error
	Topic string
}

type AddedToBlacklistData struct {
	AdderID common.Address
	AddeeID common.Address
}

type RemovedFromBlacklistData struct {
	RemoverID common.Address
	RemoveeID common.Address
}

type ValidatorCreatedData struct {
	ID common.Address
}

type ValidatorDeletedData struct {
	ID common.Address
}

type CertificateCreatedData struct {
	ID *big.Int
}
