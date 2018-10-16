package optimus

import (
	"context"

	"github.com/sonm-io/core/proto"
	"golang.org/x/sync/errgroup"
)

const (
	pullLimit       = 1000
	preallocateSize = 1 << 13
)

type MarketOrder = sonm.DWHOrder

type DealRequestFactory func() *sonm.DealsRequest
type OrderRequestFactory func() *sonm.OrdersRequest

func DefaultDealRequestFactory(cfg marketplaceConfig) DealRequestFactory {
	return func() *sonm.DealsRequest {
		return &sonm.DealsRequest{
			Price: &sonm.MaxMinBig{
				Min: cfg.MinPrice.GetPerSecond(),
			},
		}
	}
}

func DefaultOrderRequestFactory(cfg marketplaceConfig) OrderRequestFactory {
	return func() *sonm.OrdersRequest {
		return &sonm.OrdersRequest{
			Type:   sonm.OrderType_BID,
			Status: sonm.OrderStatus_ORDER_ACTIVE,
			Price: &sonm.MaxMinBig{
				Min: cfg.MinPrice.GetPerSecond(),
			},
		}
	}
}

type marketScanner struct {
	dwh                  sonm.DWHClient
	dealsRequestFactory  DealRequestFactory
	ordersRequestFactory OrderRequestFactory
}

func newMarketScanner(cfg marketplaceConfig, dwh sonm.DWHClient) *marketScanner {
	return &marketScanner{
		dwh:                  dwh,
		dealsRequestFactory:  DefaultDealRequestFactory(cfg),
		ordersRequestFactory: DefaultOrderRequestFactory(cfg),
	}
}

func (m *marketScanner) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	cursor := newOrderCursor(m.dwh, m.ordersRequestFactory)

	result := make([]*MarketOrder, 0, preallocateSize)
	for {
		next, err := cursor.Next(ctx)
		if err != nil {
			return nil, err
		}

		if len(next) == 0 {
			break
		}

		result = append(result, next...)
	}

	return result, nil
}

func (m *marketScanner) Deals(ctx context.Context) ([]*sonm.DWHDeal, error) {
	cursor := newDealCursor(m.dwh, m.dealsRequestFactory)

	result := make([]*sonm.DWHDeal, 0, preallocateSize)
	for {
		next, err := cursor.Next(ctx)
		if err != nil {
			return nil, err
		}

		if len(next) == 0 {
			break
		}

		result = append(result, next...)
	}

	return result, nil
}

func (m *marketScanner) ExecutedOrders(ctx context.Context, orderType sonm.OrderType) ([]*MarketOrder, error) {
	response, err := m.Deals(ctx)
	if err != nil {
		return nil, err
	}

	orderIds := make([]*sonm.BigInt, 0, len(response))
	for _, deal := range response {
		if orderType == sonm.OrderType_BID || orderType == sonm.OrderType_ANY {
			orderIds = append(orderIds, deal.GetDeal().GetBidID())
		}
		if orderType == sonm.OrderType_ASK || orderType == sonm.OrderType_ANY {
			orderIds = append(orderIds, deal.GetDeal().GetAskID())
		}

		// Ignore deals that costs less than 1e-6 USD/h.
		if deal.GetDeal().GetPrice().Cmp(sonm.NewBigIntFromInt(277777777)) <= 0 {
			continue
		}
	}

	orders, err := m.orders(ctx, orderIds)
	if err != nil {
		return nil, err
	}

	// Leave only orders without counterparty.
	filteredOrders := make([]*MarketOrder, 0, len(orders))
	for _, order := range orders {
		if order.GetOrder().GetCounterpartyID().IsZero() {
			filteredOrders = append(filteredOrders, order)
		}
	}

	return filteredOrders, nil
}

func (m *marketScanner) orders(ctx context.Context, ids []*sonm.BigInt) ([]*MarketOrder, error) {
	const chunkSize = 10000

	orders := make([]*MarketOrder, len(ids))
	wg, ctx := errgroup.WithContext(ctx)
	for id := 0; id < len(ids)/chunkSize+1; id++ {
		id := id

		maxId := (id + 1) * chunkSize
		if maxId > len(ids) {
			maxId = len(ids)
		}

		wg.Go(func() error {
			response, err := m.dwh.GetOrdersByIDs(ctx, &sonm.OrdersByIDsRequest{
				Ids: ids[id*chunkSize : maxId],
			})
			if err != nil {
				return err
			}

			copy(orders[id*chunkSize:maxId], response.GetOrders())
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return orders, nil
}

type cursorOrder struct {
	dwh            sonm.DWHClient
	offset         uint64
	limit          uint64
	requestFactory OrderRequestFactory
}

func newOrderCursor(dwh sonm.DWHClient, requestFactory OrderRequestFactory) *cursorOrder {
	return &cursorOrder{
		dwh:            dwh,
		offset:         0,
		limit:          pullLimit,
		requestFactory: requestFactory,
	}
}

func (m *cursorOrder) Next(ctx context.Context) ([]*MarketOrder, error) {
	request := m.requestFactory()
	request.Offset = m.offset
	request.Limit = m.limit

	response, err := m.dwh.GetOrders(ctx, request)
	if err != nil {
		return nil, err
	}

	m.offset += uint64(len(response.Orders))

	return response.Orders, nil
}

type cursorDeal struct {
	dwh            sonm.DWHClient
	offset         uint64
	limit          uint64
	requestFactory DealRequestFactory
}

func newDealCursor(dwh sonm.DWHClient, requestFactory DealRequestFactory) *cursorDeal {
	return &cursorDeal{
		dwh:            dwh,
		offset:         0,
		limit:          pullLimit,
		requestFactory: requestFactory,
	}
}

func (m *cursorDeal) Next(ctx context.Context) ([]*sonm.DWHDeal, error) {
	request := m.requestFactory()
	request.Offset = m.offset
	request.Limit = m.limit

	response, err := m.dwh.GetDeals(ctx, request)
	if err != nil {
		return nil, err
	}

	m.offset += uint64(len(response.GetDeals()))

	return response.GetDeals(), nil
}
