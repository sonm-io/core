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
	cfg marketplaceConfig
	dwh sonm.DWHClient
}

func newMarketScanner(cfg marketplaceConfig, dwh sonm.DWHClient) *marketScanner {
	return &marketScanner{
		cfg: cfg,
		dwh: dwh,
	}
}

func (m *marketScanner) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	cursor := newCursor(m.cfg, m.dwh)

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
	cfg    marketplaceConfig
	dwh    sonm.DWHClient
	offset uint64
	limit  uint64
}

func newCursor(cfg marketplaceConfig, dwh sonm.DWHClient) *cursor {
	return &cursor{
		cfg:    cfg,
		dwh:    dwh,
		offset: 0,
		limit:  ordersPullLimit,
	}
}

func (m *cursor) Next(ctx context.Context) ([]*MarketOrder, error) {
	request := &sonm.OrdersRequest{
		Type:   sonm.OrderType_ANY,
		Status: sonm.OrderStatus_ORDER_ACTIVE,
		Price: &sonm.MaxMinBig{
			Min: m.cfg.MinPrice.GetPerSecond(),
		},
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
