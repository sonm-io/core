package node

import (
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type marketAPI struct{}

func (m *marketAPI) GetOrders(context.Context, *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	return &pb.GetOrdersReply{}, nil
}

func (m *marketAPI) GetOrderByID(context.Context, *pb.ID) (*pb.Order, error) {
	return &pb.Order{}, nil
}

func (m *marketAPI) CreateOrder(context.Context, *pb.Order) (*pb.Order, error) {
	return &pb.Order{}, nil
}

func (m *marketAPI) CancelOrder(context.Context, *pb.Order) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func newMarketAPI() pb.MarketServer {
	return &marketAPI{}
}
