package dealer

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

// Selector interface describes method to select most profitable order from search results.
type Selector interface {
	Select(orders []*sonm.Order) (*sonm.Order, error)
}

// selector provides generic select implementation (thnx, @antmat!)
type selector struct {
	better func(lhs, rhs *sonm.Order) bool
}

func (m *selector) Select(orders []*sonm.Order) (*sonm.Order, error) {
	if len(orders) == 0 {
		return nil, errors.New("no orders provided")
	}

	var best = orders[0]
	for _, o := range orders {
		if m.better(o, best) {
			best = o
		}
	}

	return best, nil
}

// NewAskSelector returns Selector implementation which
// returns most cheapest ASK order.
func NewAskSelector() Selector {
	return &selector{
		better: func(lhs, rhs *sonm.Order) bool {
			return lhs.PricePerSecond.Cmp(rhs.PricePerSecond) < 0
		},
	}
}

// NewBidSelector returns Selector implementation which
// returns most expensive BID.
func NewBidSelector() Selector {
	return &selector{
		better: func(lhs, rhs *sonm.Order) bool {
			return lhs.PricePerSecond.Cmp(rhs.PricePerSecond) > 0
		},
	}
}
