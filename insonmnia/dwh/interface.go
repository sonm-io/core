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
	GetOrders(ctx context.Context, filter OrderFilter) ([]*sonm.Order, error)
	GetDeals(ctx context.Context, filter DealsFilter) ([]*sonm.Deal, error)
}

type dumbDWH struct {
	ctx context.Context
}

func (dwh *dumbDWH) GetDeals(ctx context.Context, filter DealsFilter) ([]*sonm.Deal, error) {
	return []*sonm.Deal{}, nil
}

func (dwh *dumbDWH) GetOrders(ctx context.Context, filter OrderFilter) ([]*sonm.Order, error) {
	return []*sonm.Order{}, nil
}

func NewDumbDWH(ctx context.Context) DWH {
	return &dumbDWH{ctx: ctx}
}
