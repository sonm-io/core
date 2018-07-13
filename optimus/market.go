package optimus

import (
	"context"

	"github.com/sonm-io/core/proto"
)

const (
	ordersPullLimit       = 1000
	ordersPreallocateSize = 1 << 13
)

type MarketOrder = sonm.DWHOrder

type marketScanner struct {
	dwh sonm.DWHClient
}

func newMarketScanner(dwh sonm.DWHClient) *marketScanner {
	return &marketScanner{
		dwh: dwh,
	}
}

func (m *marketScanner) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	cursor := newCursor(m.dwh)

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

type cursor struct {
	dwh    sonm.DWHClient
	offset uint64
	limit  uint64
}

func newCursor(dwh sonm.DWHClient) *cursor {
	return &cursor{
		dwh:    dwh,
		offset: 0,
		limit:  ordersPullLimit,
	}
}

func (m *cursor) Next(ctx context.Context) ([]*MarketOrder, error) {
	request := &sonm.OrdersRequest{
		Type:   sonm.OrderType_ANY,
		Status: sonm.OrderStatus_ORDER_ACTIVE,
		Offset: m.offset,
		Limit:  m.limit,
	}

	response, err := m.dwh.GetOrders(ctx, request)
	if err != nil {
		return nil, err
	}

	m.offset += uint64(len(response.Orders))

	return response.Orders, nil
}
