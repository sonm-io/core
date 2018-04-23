package node

import (
	"github.com/ethereum/go-ethereum/crypto"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type marketAPI struct {
	remotes    *remoteOptions
	ctx        context.Context
	hubCreator hubClientCreator
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.Count) (*pb.GetOrdersReply, error) {
	filter := &pb.OrdersRequest{
		Type:     pb.OrderType_BID,
		Status:   pb.OrderStatus_ORDER_ACTIVE,
		AuthorID: crypto.PubkeyToAddress(m.remotes.key.PublicKey).Hex(),
		Limit:    req.GetCount(),
	}

	orders, err := m.remotes.dwh.GetOrders(ctx, filter)
	if err != nil {
		return nil, err
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
