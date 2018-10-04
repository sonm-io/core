package optimus

import (
	"context"

	"github.com/sonm-io/core/proto"
	"golang.org/x/sync/errgroup"
)

const (
	pullLimit             = 1000
	ordersPreallocateSize = 1 << 13
)

type MarketOrder = sonm.DWHOrder

type OrderRequestFactory func() *sonm.OrdersRequest

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
	dwh            sonm.DWHClient
	requestFactory OrderRequestFactory
}

func newMarketScanner(cfg marketplaceConfig, dwh sonm.DWHClient) *marketScanner {
	return &marketScanner{
		dwh:            dwh,
		requestFactory: DefaultOrderRequestFactory(cfg),
	}
}

func (m *marketScanner) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	cursor := newCursor(m.dwh, m.requestFactory)

	orders := make([]*MarketOrder, 0, ordersPreallocateSize)
	for {
		nextOrders, err := cursor.Next(ctx)
		if err != nil {
			return nil, err
		}

		if len(nextOrders) == 0 {
			break
		}

		orders = append(orders, nextOrders...)
	}

	return orders, nil
}

func (m *marketScanner) ExecutedOrders(ctx context.Context, orderType sonm.OrderType) ([]*MarketOrder, error) {
	response, err := m.dwh.GetDeals(ctx, &sonm.DealsRequest{})
	if err != nil {
		return nil, err
	}

	orderIds := make([]*sonm.BigInt, 0, len(response.Deals))
	for _, deal := range response.Deals {
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

type cursor struct {
	dwh            sonm.DWHClient
	offset         uint64
	limit          uint64
	requestFactory OrderRequestFactory
}

func newCursor(dwh sonm.DWHClient, requestFactory OrderRequestFactory) *cursor {
	return &cursor{
		dwh:            dwh,
		offset:         0,
		limit:          pullLimit,
		requestFactory: requestFactory,
	}
}

func (m *cursor) Next(ctx context.Context) ([]*MarketOrder, error) {
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
