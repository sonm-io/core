package node

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/dwh"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type marketAPI struct {
	remotes       *remoteOptions
	ctx           context.Context
	workerCreator workerClientCreator
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
		deal, err := m.remotes.orderMatcher.CreateDealByOrder(m.remotes.ctx, order)
		if err != nil {
			ctxlog.G(m.remotes.ctx).Warn("cannot open deal", zap.Error(err))
			return
		}

		ctxlog.G(m.remotes.ctx).Info("opened deal for order",
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
		return nil, errors.WithMessage(err, "cannot get orders from dwh")
	}

	merr := multierror.NewMultiError()
	for _, order := range orders.Orders {
		id := order.GetOrder().GetId().Unwrap()
		ctxlog.G(m.remotes.ctx).Debug("cancelling order", zap.String("id", id.String()))
		if err := m.remotes.eth.Market().CancelOrder(ctx, m.remotes.key, id); err != nil {
			multierror.Append(merr, fmt.Errorf("cannot cancel order with id %s: %v", id.String(), err))
		}
	}

	if len(merr.WrappedErrors()) > 0 {
		return nil, merr.ErrorOrNil()
	}

	return &pb.Empty{}, nil
}

func newMarketAPI(opts *remoteOptions) (pb.MarketServer, error) {
	return &marketAPI{
		remotes:       opts,
		ctx:           opts.ctx,
		workerCreator: opts.workerCreator,
	}, nil
}
