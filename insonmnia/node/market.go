package node

import (
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type marketAPI struct {
	conf Config
	cc   pb.MarketClient
	ctx  context.Context
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

	if req.OrderType != pb.OrderType_BID {
		return nil, errors.New("can create only Orders with type BID")
	}

	req.ByuerID = m.conf.ClientID()
	return m.cc.CreateOrder(ctx, req)
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	log.G(m.ctx).Debug("handling CancelOrder request", zap.String("id", req.Id))
	return m.cc.CancelOrder(ctx, req)
}

func newMarketAPI(ctx context.Context, conf Config) (pb.MarketServer, error) {
	// TODO(sshaman1101): enable compression into marketplace
	cc, err := grpc.Dial(conf.MarketEndpoint(), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &marketAPI{
		conf: conf,
		ctx:  ctx,
		cc:   pb.NewMarketClient(cc),
	}, nil
}
