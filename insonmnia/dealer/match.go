package dealer

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

// Matcher interface describes method to select most profitable order from search results.
type Matcher interface {
	Match(orders []*sonm.Order) (*sonm.Order, error)
}

// bidMatcher matches the cheapest order to deal
type bidMatcher struct{}

// NewBidMatcher returns Matcher implementation which
// matches given BID order with most cheapest ASK.
func NewBidMatcher() Matcher {
	return &bidMatcher{}
}

func (m *bidMatcher) Match(orders []*sonm.Order) (*sonm.Order, error) {
	if len(orders) == 0 {
		return nil, errors.New("no orders provided")
	}

	var min = orders[0]
	for _, o := range orders {
		if o.PricePerSecond.Cmp(min.PricePerSecond) < 0 {
			min = o
		}
	}

	return min, nil
}

type askMatcher struct{}

// NewAskMatcher returns Matcher implementation which
// matches given ASK order with most expensive BID.
func NewAskMatcher() Matcher {
	return &askMatcher{}
}

func (m *askMatcher) Match(orders []*sonm.Order) (*sonm.Order, error) {
	if len(orders) == 0 {
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
