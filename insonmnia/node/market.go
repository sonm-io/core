package node

import (
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type marketAPI struct {
	conf   Config
	market pb.MarketClient
	ctx    context.Context
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	log.G(m.ctx).Info("handling GetOrders request")
	return m.market.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	log.G(m.ctx).Info("handling GetOrderByID request", zap.String("id", req.Id))
	return m.market.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	log.G(m.ctx).Info("handling CreateOrder request")

	if req.OrderType != pb.OrderType_BID {
		return nil, errors.New("can create only Orders with type BID")
	}

	req.ByuerID = m.conf.ClientID()
	return m.market.CreateOrder(ctx, req)
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	log.G(m.ctx).Info("handling CancelOrder request", zap.String("id", req.Id))
	return m.market.CancelOrder(ctx, req)
}

func newMarketAPI(ctx context.Context, conf Config) (pb.MarketServer, error) {
	cc, err := util.MakeGrpcClient(conf.MarketEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	return &marketAPI{
		conf:   conf,
		ctx:    ctx,
		market: pb.NewMarketClient(cc),
	}, nil
}
