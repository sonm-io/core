package dwh

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

type DealsFilter struct {
	Author common.Address
	Status sonm.MarketDealStatus
}

type OrderFilter struct {
	Count        uint64
	Type         sonm.MarketOrderType
	Price        *big.Int
	Counterparty common.Address
}

type DWH interface {
	GetOrders(ctx context.Context, filter OrderFilter) ([]*sonm.MarketOrder, error)
	GetDeals(ctx context.Context, filter DealsFilter) ([]*sonm.MarketDeal, error)
}

type dumbDWH struct {
	ctx context.Context
}

func (dwh *dumbDWH) GetDeals(ctx context.Context, filter DealsFilter) ([]*sonm.MarketDeal, error) {
	return []*sonm.MarketDeal{}, nil
}

func (dwh *dumbDWH) GetOrders(ctx context.Context, filter OrderFilter) ([]*sonm.MarketOrder, error) {
	return []*sonm.MarketOrder{}, nil
}

func NewDumbDWH(ctx context.Context) DWH {
	return &dumbDWH{ctx: ctx}
}
