package node

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/dwh"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type marketAPI struct {
	remotes    *remoteOptions
	ctx        context.Context
	hubCreator hubClientCreator
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	filter := dwh.OrderFilter{
		Type:         req.Type,
		Count:        req.Count,
		Price:        req.Price.Unwrap(),
		Counterparty: req.Counterparty.Unwrap(),
	}

	orders, err := m.remotes.dwh.GetOrders(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &pb.GetOrdersReply{Orders: orders}, nil
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.MarketOrder, error) {
	id, ok := big.NewInt(0).SetString(req.GetId(), 10)
	if !ok {
		return nil, errors.New("cannot convert value to *big.Int")
	}

	return m.remotes.eth.GetOrderInfo(ctx, id)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.MarketOrder) (*pb.MarketOrder, error) {
	id, err := m.remotes.eth.PlaceOrderPending(ctx, m.remotes.key, req, m.remotes.blockchainTimeout)
	if err != nil {
		return nil, err
	}

	req.Id = id.String()
	return req, nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.ID) (*pb.Empty, error) {
	id, ok := big.NewInt(0).SetString(req.GetId(), 10)
	if !ok {
		return nil, errors.New("cannot convert value to *big.Int")
	}

	if _, err := m.remotes.eth.CancelOrder(ctx, m.remotes.key, id); err != nil {
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
