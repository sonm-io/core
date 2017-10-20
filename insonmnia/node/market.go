package node

import (
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type marketAPI struct {
	endpoint string
	cc       pb.MarketClient
	ctx      context.Context
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	log.G(m.ctx).Debug("handling GetOrders request")
	return m.cc.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	log.G(m.ctx).Debug("handling GetOrderByID request", zap.String("id", req.Id))
	return m.cc.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	log.G(m.ctx).Debug("handling CreateOrder request")
	return m.cc.CreateOrder(ctx, req)
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	log.G(m.ctx).Debug("handling CancelOrder request", zap.String("id", req.Id))
	return m.cc.CancelOrder(ctx, req)
}

func newMarketAPI(ctx context.Context, endpoint string) (pb.MarketServer, error) {
	cc, err := initGrpcClient(endpoint, nil)
	if err != nil {
		return nil, err
	}

	return &marketAPI{
		endpoint: endpoint,
		cc:       pb.NewMarketClient(cc),
		ctx:      ctx,
	}, nil
}
