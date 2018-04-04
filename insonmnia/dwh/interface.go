package dwh

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

type DealsFilter struct {
	Author common.Address
}

type DWH interface {
	GetDeals(filter DealsFilter) ([]*sonm.MarketDeal, error)
}

type dumbDWH struct {
	ctx context.Context
}

func (dwh *dumbDWH) GetDeals(filter DealsFilter) ([]*sonm.MarketDeal, error) {
	return []*sonm.MarketDeal{}, nil
}

func NewDumbDWH(ctx context.Context) DWH {
	return &dumbDWH{ctx: ctx}
}
