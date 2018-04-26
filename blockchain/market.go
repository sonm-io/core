package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	marketAPI "github.com/sonm-io/core/blockchain/market/api"
	pb "github.com/sonm-io/core/proto"
)

type BasicMarketAPI struct {
	client         *ethclient.Client
	marketContract *marketAPI.Market
	gasPrice       int64
	logParsePeriod time.Duration
}

func NewBasicMarketAPI(client *ethclient.Client, address common.Address, gasPrice int64, logParsePeriod time.Duration) (MarketAPI, error) {
	marketContract, err := marketAPI.NewMarket(address, client)
	if err != nil {
		return nil, err
	}

	api := &BasicMarketAPI{
		client:         client,
		marketContract: marketContract,
		gasPrice:       gasPrice,
		logParsePeriod: logParsePeriod,
	}
	return api, nil
}

func (api *BasicMarketAPI) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int) chan DealOrError {
	ch := make(chan DealOrError, 0)
	go api.openDeal(ctx, key, askID, bidID, ch)
	return ch
}

func (api *BasicMarketAPI) openDeal(ctx context.Context, key *ecdsa.PrivateKey, askID, bidID *big.Int, ch chan DealOrError) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	tx, err := api.marketContract.OpenDeal(opts, askID, bidID)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	id, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, DealOpenedTopic)
	if err != nil {
		ch <- DealOrError{nil, err}
		return
	}

	deal, err := api.GetDealInfo(ctx, id)
	ch <- DealOrError{deal, err}
}

func (api *BasicMarketAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklisted bool) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.marketContract.CloseDeal(opts, dealID, blacklisted)
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

func (api *BasicMarketAPI) GetDealsAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetDealsAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order) chan OrderOrError {
	ch := make(chan OrderOrError, 0)
	go api.placeOrder(ctx, key, order, ch)
	return ch
}

func (api *BasicMarketAPI) placeOrder(ctx context.Context, key *ecdsa.PrivateKey, order *pb.Order, ch chan OrderOrError) {
	opts := getTxOpts(ctx, key, gasLimitForPlaceOrderMethod, api.gasPrice)

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

	id, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, OrderPlacedTopic)
	if err != nil {
		ch <- OrderOrError{nil, err}
		return
	}
	orderInfo, err := api.GetOrderInfo(ctx, id)
	ch <- OrderOrError{orderInfo, err}
}

func (api *BasicMarketAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) chan error {
	ch := make(chan error, 0)
	go api.cancelOrder(ctx, key, id, ch)
	return ch
}

func (api *BasicMarketAPI) cancelOrder(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int, ch chan error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	tx, err := api.marketContract.CancelOrder(opts, id)
	if err != nil {
		ch <- err
		return
	}

	if _, err := waitForTransactionResult(ctx, api.client, api.logParsePeriod, tx, OrderUpdatedTopic); err != nil {
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

func (api *BasicMarketAPI) GetOrdersAmount(ctx context.Context) (*big.Int, error) {
	return api.marketContract.GetOrdersAmount(getCallOptions(ctx))
}

func (api *BasicMarketAPI) Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.marketContract.Bill(opts, dealID)
}

func (api *BasicMarketAPI) RegisterWorker(ctx context.Context, key *ecdsa.PrivateKey, master common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.marketContract.RegisterWorker(opts, master)
}

func (api *BasicMarketAPI) ConfirmWorker(ctx context.Context, key *ecdsa.PrivateKey, slave common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
	return api.marketContract.RegisterWorker(opts, slave)
}

func (api *BasicMarketAPI) RemoveWorker(ctx context.Context, key *ecdsa.PrivateKey, master, slave common.Address) (*types.Transaction, error) {
	opts := getTxOpts(ctx, key, defaultGasLimit, api.gasPrice)
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
		DealID:      changeRequest.DealID.String(),
		RequestType: pb.OrderType(changeRequest.RequestType),
		Duration:    changeRequest.Duration.Uint64(),
		Price:       pb.NewBigInt(changeRequest.Price),
		Status:      pb.ChangeRequestStatus(changeRequest.Status),
	}, nil
}
