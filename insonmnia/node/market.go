package node

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

const purgeConcurrency = 32

type marketAPI struct {
	remotes       *remoteOptions
	workerCreator workerClientCreator
	log           *zap.SugaredLogger
}

func (m *marketAPI) GetOrders(ctx context.Context, req *sonm.Count) (*sonm.GetOrdersReply, error) {
	filter := &sonm.OrdersRequest{
		Type:     sonm.OrderType_BID,
		Status:   sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID: sonm.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		Limit:    req.GetCount(),
	}

	orders, err := m.remotes.dwh.GetOrders(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not get orders from DWH: %s", err)
	}

	reply := &sonm.GetOrdersReply{Orders: []*sonm.Order{}}
	for _, order := range orders.GetOrders() {
		reply.Orders = append(reply.Orders, order.Order)
	}

	return reply, nil
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *sonm.ID) (*sonm.Order, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("could not get parse order id %s to BigInt: %s", req.GetId(), err)
	}

	return m.remotes.eth.Market().GetOrderInfo(ctx, id)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *sonm.BidOrder) (*sonm.Order, error) {
	knownBenchmarks := m.remotes.benchList.MapByCode()
	givenBenchmarks := req.GetResources().GetBenchmarks()

	if len(givenBenchmarks) > len(knownBenchmarks) {
		return nil, fmt.Errorf("benchmark list too large")
	}

	if req.GetIdentity() == sonm.IdentityLevel_UNKNOWN {
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

	benchStruct, err := sonm.NewBenchmarks(benchmarksValues)
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark values: %s", err)
	}

	var blacklist string
	if req.GetBlacklist() != nil {
		blacklist = req.GetBlacklist().Unwrap().Hex()
	}

	order := &sonm.Order{
		OrderType:      sonm.OrderType_BID,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		CounterpartyID: req.GetCounterparty(),
		Duration:       uint64(req.GetDuration().Unwrap().Seconds()),
		Price:          req.GetPrice().GetPerSecond(),
		Netflags:       &sonm.NetFlags{},
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

func (m *marketAPI) CancelOrder(ctx context.Context, req *sonm.ID) (*sonm.Empty, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("could not get parse order id %s to BigInt: %s", req.GetId(), err)
	}

	if err := m.remotes.eth.Market().CancelOrder(ctx, m.remotes.key, id); err != nil {
		return nil, fmt.Errorf("could not get cancel order %s on blockchain: %s", req.GetId(), err)
	}

	return &sonm.Empty{}, nil
}

func (m *marketAPI) Purge(ctx context.Context, req *sonm.Empty) (*sonm.Empty, error) {
	status, err := m.PurgeVerbose(ctx, req)
	if err != nil {
		return nil, err
	}
	if err = sonm.CombinedError(status); err != nil {
		return nil, err
	}
	return &sonm.Empty{}, nil
}

func (m *marketAPI) PurgeVerbose(ctx context.Context, _ *sonm.Empty) (*sonm.ErrorByID, error) {
	orders, err := m.remotes.dwh.GetOrders(ctx, &sonm.OrdersRequest{
		Type:     sonm.OrderType_BID,
		Status:   sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID: sonm.NewEthAddress(crypto.PubkeyToAddress(m.remotes.key.PublicKey)),
		Limit:    dwh.MaxLimit,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot get orders from dwh: %v", err)
	}

	ids := make([]*sonm.BigInt, 0, len(orders.GetOrders()))
	for _, order := range orders.GetOrders() {
		ids = append(ids, order.GetOrder().GetId())
	}
	return m.cancelOrders(ctx, ids)
}

func (m *marketAPI) CancelOrders(ctx context.Context, req *sonm.OrderIDs) (*sonm.ErrorByID, error) {
	return m.cancelOrders(ctx, req.GetIds())
}

func (m *marketAPI) cancelOrders(ctx context.Context, ids []*sonm.BigInt) (*sonm.ErrorByID, error) {
	concurrency := purgeConcurrency
	if len(ids) < concurrency {
		concurrency = len(ids)
	}
	status := sonm.NewTSErrorByID()
	ch := make(chan *sonm.BigInt)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for id := range ch {
				if id.IsZero() {
					status.Append(id, errors.New("nil order id specified"))
				} else {
					m.log.Debugw("cancelling order", zap.String("id", id.Unwrap().String()))
					err := m.remotes.eth.Market().CancelOrder(ctx, m.remotes.key, id.Unwrap())
					status.Append(id, err)
				}
			}
		}()
	}
	for _, id := range ids {
		ch <- id
	}
	close(ch)
	wg.Wait()

	return status.Unwrap(), nil
}

func newMarketAPI(opts *remoteOptions) sonm.MarketServer {
	return &marketAPI{
		remotes:       opts,
		workerCreator: opts.workerCreator,
		log:           opts.log,
	}
}
