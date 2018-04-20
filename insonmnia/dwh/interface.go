package dwh

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

type DealsFilter struct {
	Author common.Address
	Status sonm.DealStatus
}

type OrderFilter struct {
	Count        uint64
	Type         sonm.OrderType
	Price        *big.Int
	Counterparty common.Address
}

type MockDWH interface {
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

func NewDumbDWH(ctx context.Context) MockDWH {
	return &dumbDWH{ctx: ctx}
}
