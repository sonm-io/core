package dealer

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
)

type Searcher interface {
	Search(context.Context, *SearchFilter) ([]*sonm.Order, error)
}

type askSearcher struct {
	market sonm.MarketClient
}

func NewAskSearcher(market sonm.MarketClient) Searcher {
	return &askSearcher{
		market: market,
	}
}

func (s *askSearcher) filterByPriceAndAllowance(orders []*sonm.Order, balance, allowance *big.Int) ([]*sonm.Order, error) {
	// disallow unclean input
	if orders == nil {
		return nil, errors.New("orders cannot be nil")
	}

	var err error
	orders, err = s.filterByPrice(orders, balance)
	if err != nil {
		return nil, err
	}

	orders, err = s.filterByAllowance(orders, allowance)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *askSearcher) filterByPrice(orders []*sonm.Order, balance *big.Int) ([]*sonm.Order, error) {
	var res []*sonm.Order
	for _, o := range orders {
		p := structs.CalculateTotalPrice(o)
		// price < balance, we can handle this order
		if p.Cmp(balance) == -1 {
			res = append(res, o)
		}
	}

	if len(res) == 0 {
		return nil, errors.New("no orders fit into available balance")
	}

	return res, nil
}

func (s *askSearcher) filterByAllowance(orders []*sonm.Order, allowance *big.Int) ([]*sonm.Order, error) {
	var res []*sonm.Order
	for _, o := range orders {
		p := structs.CalculateTotalPrice(o)
		// if allowance > price
		if allowance.Cmp(p) > 0 {
			res = append(res, o)
		}
	}

	if len(res) == 0 {
		return nil, errors.New("no orders fit into allowance")
	}

	return res, nil
}

func (s *askSearcher) Search(ctx context.Context, filter *SearchFilter) ([]*sonm.Order, error) {
	req := &sonm.GetOrdersRequest{
		Order: filter.order,
		Count: 100, // wow, such number, very magic
	}

	// query market for orders
	reply, err := s.market.GetOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	// apply extra filter for price and allowance
	// todo(all): can market/dwh perform filtering by price itself?
	orders, err := s.filterByPriceAndAllowance(reply.GetOrders(), filter.balance, filter.allowance)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

// SearchFilter represent params for searching and filtering orders
// on the marketplace. Accepted by `Searcher` interface.
type SearchFilter struct {
	order     *sonm.Order
	balance   *big.Int
	allowance *big.Int
}

// NewSearchFilter validates input data and constructs `SearchFilter`
func NewSearchFilter(order *sonm.Order, balance, allowance *big.Int) (*SearchFilter, error) {
	if order == nil {
		return nil, errors.New("order cannot be nil")
	}

	if balance == nil {
		return nil, errors.New("balance cannot be nil")
	}

	if allowance == nil {
		return nil, errors.New("allowance cannot be nil")
	}

	return &SearchFilter{
		order:     order,
		balance:   balance,
		allowance: allowance,
	}, nil
}
