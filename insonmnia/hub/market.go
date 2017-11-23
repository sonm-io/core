package hub

import (
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type Market interface {
	// OrderExists checks whether an order with the specified ID exists in the
	// marketplace.
	OrderExists(ID string) (bool, error)
	// CancelOrder removes order from marketplace
	CancelOrder(ID string) error
}

type market struct {
	ctx    context.Context
	client pb.MarketClient
}

func (m *market) OrderExists(ID string) (bool, error) {
	_, err := m.client.GetOrderByID(m.ctx, &pb.ID{Id: ID})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *market) CancelOrder(ID string) error {
	_, err := m.client.CancelOrder(m.ctx, &pb.Order{Id: ID})
	return err
}

// NewMarket constructs a new SONM marketplace client.
func NewMarket(ctx context.Context, addr string) (Market, error) {
	cc, err := util.MakeGrpcClient(ctx, addr, nil)
	if err != nil {
		return nil, err
	}

	return &market{
		ctx:    ctx,
		client: pb.NewMarketClient(cc),
	}, nil
}
