package main

import (
	"context"

	"github.com/mitchellh/mapstructure"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type DWHOrdersAmmo struct {
	order *sonm.OrdersRequest
}

func (m *DWHOrdersAmmo) Type() AmmoType {
	return DWHOrders
}

func (m *DWHOrdersAmmo) Execute(ctx context.Context, ext interface{}) error {
	mExt := ext.(*dwhExt)
	orders, err := mExt.dwh.GetOrders(ctx, m.order)
	if err != nil {
		return err
	}

	mExt.log.Debug("OK", zap.String("ammo", "DWHOrders"), zap.Any("count", len(orders.Orders)))

	return nil
}

type dwhOrdersAmmoFactory struct {
	PoolAmmoFactory
	Orders []*sonm.OrdersRequest
}

func newDWHOrdersAmmoFactory(cfg interface{}) (AmmoFactory, error) {
	config := struct {
		Orders []*sonm.OrdersRequest
	}{}
	if err := mapstructure.Decode(cfg, &config); err != nil {
		return nil, err
	}

	m := &dwhOrdersAmmoFactory{
		PoolAmmoFactory: newPoolAmmoFactory(func() interface{} {
			return &DWHOrdersAmmo{}
		}),
		Orders: config.Orders,
	}

	return m, nil
}

func (m *dwhOrdersAmmoFactory) New(id int) Ammo {
	ammo := m.pool.Get().(*DWHOrdersAmmo)
	ammo.order = m.Orders[id%len(m.Orders)]

	return ammo
}

func (m *dwhOrdersAmmoFactory) NewDefault() Ammo {
	return &DWHOrdersAmmo{}
}
