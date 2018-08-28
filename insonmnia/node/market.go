package node

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/dwh"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

const purgeConcurrency = 32

type marketAPI struct {
	remotes       *remoteOptions
	workerCreator workerClientCreator
	log           *zap.SugaredLogger
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.Count) (*pb.GetOrdersReply, error) {
	filter := &pb.OrdersRequest{
		Type:     pb.OrderType_BID,
		Status:   pb.OrderStatus_ORDER_ACTIVE,
		AuthorID: pb.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		Limit:    req.GetCount(),
	}

	orders, err := m.remotes.dwh.GetOrders(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not get orders from DWH: %s", err)
	}

	reply := &pb.GetOrdersReply{Orders: []*pb.Order{}}
	for _, order := range orders.GetOrders() {
		reply.Orders = append(reply.Orders, order.Order)
	}

	return reply, nil
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("could not get parse order id %s to BigInt: %s", req.GetId(), err)
	}

	return m.remotes.eth.Market().GetOrderInfo(ctx, id)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.BidOrder) (*pb.Order, error) {
	knownBenchmarks := m.remotes.benchList.MapByCode()
	givenBenchmarks := req.GetResources().GetBenchmarks()

	if len(givenBenchmarks) > len(knownBenchmarks) {
		return nil, fmt.Errorf("benchmark list too large")
	}

	if req.GetIdentity() == pb.IdentityLevel_UNKNOWN {
		return nil, errors.New("identity level is required and should not be 0")
	}

	benchmarksValues := make([]uint64, len(knownBenchmarks))
	for code, value := range givenBenchmarks {
		bench, ok := knownBenchmarks[code]
		if !ok {
			return nil, fmt.Errorf("unknown benchmark code \"%s\"", code)
		}

		benchmarksValues[bench.GetID()] = value
	}

	benchStruct, err := pb.NewBenchmarks(benchmarksValues)
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark values: %s", err)
	}

	var blacklist string
	if req.GetBlacklist() != nil {
		blacklist = req.GetBlacklist().Unwrap().Hex()
	}

	order := &pb.Order{
		OrderType:      pb.OrderType_BID,
		OrderStatus:    pb.OrderStatus_ORDER_ACTIVE,
		AuthorID:       pb.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		CounterpartyID: req.GetCounterparty(),
		Duration:       uint64(req.GetDuration().Unwrap().Seconds()),
		Price:          req.GetPrice().GetPerSecond(),
		Netflags:       &pb.NetFlags{},
		IdentityLevel:  req.GetIdentity(),
		Blacklist:      blacklist,
		Tag:            []byte(req.GetTag()),
		Benchmarks:     benchStruct,
	}
	net := req.GetResources().GetNetwork()
	order.Netflags.SetOverlay(net.GetOverlay())
	order.Netflags.SetIncoming(net.GetIncoming())
	order.Netflags.SetOutbound(net.GetOutbound())

	order, err = m.remotes.eth.Market().PlaceOrder(ctx, m.remotes.key, order)
	if err != nil {
		return nil, fmt.Errorf("could not place order on blockchain: %s", err)
	}

	go func() {
		deal, err := m.remotes.orderMatcher.CreateDealByOrder(context.Background(), order)
		if err != nil {
			m.log.Warnw("cannot open deal", zap.Error(err))
			return
		}

		m.log.Infow("opened deal for order",
			zap.String("orderID", order.Id.Unwrap().String()),
			zap.String("dealID", deal.Id.Unwrap().String()))
	}()

	return order, nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("could not get parse order id %s to BigInt: %s", req.GetId(), err)
	}

	if err := m.remotes.eth.Market().CancelOrder(ctx, m.remotes.key, id); err != nil {
		return nil, fmt.Errorf("could not get cancel order %s on blockchain: %s", req.GetId(), err)
	}

	return &pb.Empty{}, nil
}

func (m *marketAPI) Purge(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	orders, err := m.remotes.dwh.GetOrders(ctx, &pb.OrdersRequest{
		Type:     pb.OrderType_BID,
		Status:   pb.OrderStatus_ORDER_ACTIVE,
		AuthorID: pb.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		Limit:    dwh.MaxLimit,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot get orders from dwh: %v", err)
	}

	ids := make([]*pb.BigInt, 0, len(orders.GetOrders()))
	for _, order := range orders.GetOrders() {
		ids = append(ids, order.GetOrder().GetId())
	}
	return m.cancelOrders(ctx, ids)
}

func (m *marketAPI) CancelOrders(ctx context.Context, req *pb.OrderIDs) (*pb.Empty, error) {
	return m.cancelOrders(ctx, req.GetIds())
}

func (m *marketAPI) cancelOrders(ctx context.Context, ids []*pb.BigInt) (*pb.Empty, error) {
	concurrency := purgeConcurrency
	if len(ids) < concurrency {
		concurrency = len(ids)
	}
	merr := multierror.NewTSMultiError()
	ch := make(chan *big.Int)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for id := range ch {
				m.log.Debugw("cancelling order", zap.String("id", id.String()))
				if err := m.remotes.eth.Market().CancelOrder(ctx, m.remotes.key, id); err != nil {
					merr.Append(fmt.Errorf("cannot cancel order with id %s: %v", id.String(), err))
				}
			}
		}()
	}
	for _, id := range ids {
		if id.IsZero() {
			merr.Append(errors.New("nil order id specified"))
		} else {
			ch <- id.Unwrap()
		}
	}
	close(ch)
	wg.Wait()

	if len(merr.WrappedErrors()) > 0 {
		return nil, merr.ErrorOrNil()
	}

	return &pb.Empty{}, nil
}

func newMarketAPI(opts *remoteOptions) pb.MarketServer {
	return &marketAPI{
		remotes:       opts,
		workerCreator: opts.workerCreator,
		log:           opts.log,
	}
}
