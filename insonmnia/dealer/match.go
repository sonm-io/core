package dealer

import (
	"github.com/pkg/errors"
	"github.com/sonm-io/core/proto"
)

type Matcher interface {
	// Match takes orders found on market, returns most profitable order to deal with
	Match(orders []*sonm.Order) (*sonm.Order, error)
}

// bidMatcher matches the cheapest order to deal
type bidMatcher struct{}

func NewBidMatcher() Matcher {
	return &bidMatcher{}
}

func (m *bidMatcher) Match(orders []*sonm.Order) (*sonm.Order, error) {
	if orders == nil || len(orders) == 0 {
		return nil, errors.New("no orders provided")
	}

	// var min = &sonm.Order{PricePerSecond: sonm.NewBigIntFromInt(0)}
	var min = orders[0]
	for _, o := range orders {
		if o.PricePerSecond.Cmp(min.PricePerSecond) < 0 {
			min = o
		}
	}

	return min, nil
}

type askMatcher struct{}

func NewAskMatcher() Matcher {
	return &askMatcher{}
}

func (m *askMatcher) Match(orders []*sonm.Order) (*sonm.Order, error) {
	if orders == nil || len(orders) == 0 {
		return nil, errors.New("no orders provided")
	}

	var max = orders[0]
	for _, o := range orders {
		if o.PricePerSecond.Cmp(max.PricePerSecond) > 0 {
			max = o
		}
	}

	return max, nil
}
