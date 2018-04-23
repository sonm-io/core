package node

import (
	"github.com/sonm-io/core/insonmnia/dwh"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type marketAPI struct {
	remotes    *remoteOptions
	ctx        context.Context
	hubCreator hubClientCreator
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	filter := dwh.OrderFilter{
		Type:  pb.OrderType_ANY,
		Count: req.Count,
	}

	orders, err := m.remotes.dwh.GetOrders(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrdersReply{Orders: orders}, nil
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, err
	}

	return m.remotes.eth.GetOrderInfo(ctx, id)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	ordOrErr := <-m.remotes.eth.PlaceOrder(ctx, m.remotes.key, req)
	return ordOrErr.Order, ordOrErr.Err
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	id, err := util.ParseBigInt(req.GetId())
	if err != nil {
		return nil, err
	}

	if err := <-m.remotes.eth.CancelOrder(ctx, m.remotes.key, id); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func newMarketAPI(opts *remoteOptions) (pb.MarketServer, error) {
	return &marketAPI{
		remotes:    opts,
		ctx:        opts.ctx,
		hubCreator: opts.hubCreator,
	}, nil
}
