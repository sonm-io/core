package main

import (
	"fmt"
	"sync"

	"github.com/yandex/pandora/core"
)

type AmmoFactory interface {
	New(id int) core.Ammo
}

type orderInfoAmmoFactory struct {
	orderIDs []int64
	pool     *sync.Pool
}

func newOrderInfoAmmoFactory(orderIDs []int64, pool *sync.Pool) AmmoFactory {
	return &orderInfoAmmoFactory{
		orderIDs: orderIDs,
		pool:     pool,
	}
}

func (m *orderInfoAmmoFactory) New(id int) core.Ammo {
	orderID := m.orderIDs[id%len(m.orderIDs)]

	ammo := m.pool.Get().(*OrderInfoAmmo)
	ammo.Message = fmt.Sprintf(`Job #%d %d"`, id, orderID)
	ammo.OrderID = orderID

	return ammo
}

type OrderInfoAmmo struct {
	Message string
	OrderID int64
}

type OrderPlaceAmmo struct {
	Message string
}

type orderPlaceAmmoFactory struct {
	pool *sync.Pool
}

func newOrderPlaceAmmoFactory(pool *sync.Pool) AmmoFactory {
	return &orderPlaceAmmoFactory{
		pool: pool,
	}
}

func (m *orderPlaceAmmoFactory) New(id int) core.Ammo {
	ammo := m.pool.Get().(*OrderPlaceAmmo)
	ammo.Message = fmt.Sprintf(`Job #%d"`, id)

	return ammo
}
