package main

import (
	"context"
	"math/big"
	"math/rand"

	"github.com/mitchellh/mapstructure"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	MarketplaceOrderInfo AmmoType = iota
	MarketplaceOrderPlace
	DWHOrders
)

type OrderInfoAmmo struct {
	OrderID int64
}

func (m *OrderInfoAmmo) Type() AmmoType {
	return MarketplaceOrderInfo
}

func (m *OrderInfoAmmo) Execute(ctx context.Context, ext interface{}) error {
	mExt := ext.(*marketplaceExt)
	order, err := mExt.market.GetOrderInfo(ctx, big.NewInt(m.OrderID))
	if err != nil {
		return err
	}

	mExt.log.Debug("OK", zap.String("ammo", "GetOrderInfo"), zap.Any("order", *order))

	return nil
}

type orderInfoAmmoFactory struct {
	PoolAmmoFactory
	orderIDs []int64
}

func newOrderInfoAmmoFactory(cfg interface{}) (AmmoFactory, error) {
	config := struct {
		OrderIDs []int64 `mapstructure:"order_ids"`
	}{}

	if err := mapstructure.Decode(cfg, &config); err != nil {
		return nil, err
	}

	m := &orderInfoAmmoFactory{
		PoolAmmoFactory: newPoolAmmoFactory(func() interface{} {
			return &OrderInfoAmmo{}
		}),
		orderIDs: config.OrderIDs,
	}

	return m, nil
}

func (m *orderInfoAmmoFactory) New(id int) Ammo {
	ammo := m.pool.Get().(*OrderInfoAmmo)
	ammo.OrderID = m.orderIDs[id%len(m.orderIDs)]

	return ammo
}

func (m *orderInfoAmmoFactory) NewDefault() Ammo {
	return &OrderInfoAmmo{}
}

type OrderPlaceAmmo struct{}

func (m *OrderPlaceAmmo) Type() AmmoType {
	return MarketplaceOrderPlace
}

func (m *OrderPlaceAmmo) Execute(ctx context.Context, ext interface{}) error {
	mExt := ext.(*marketplaceExt)
	orderOrError := <-mExt.market.PlaceOrder(ctx, mExt.privateKey, order())
	if orderOrError.Err != nil {
		return orderOrError.Err
	}

	mExt.log.Debug("OK", zap.String("ammo", "PlaceOrder"), zap.Any("order", *orderOrError.Order))

	return nil
}

type orderPlaceAmmoFactory struct {
	PoolAmmoFactory
}

func newOrderPlaceAmmoFactory(cfg interface{}) (AmmoFactory, error) {
	m := &orderPlaceAmmoFactory{
		PoolAmmoFactory: newPoolAmmoFactory(func() interface{} {
			return &OrderPlaceAmmo{}
		}),
	}

	return m, nil
}

func (m *orderPlaceAmmoFactory) New(id int) Ammo {
	return m.pool.Get().(*OrderPlaceAmmo)
}

func (m *orderPlaceAmmoFactory) NewDefault() Ammo {
	return &OrderPlaceAmmo{}
}

func order() *sonm.Order {
	order := &sonm.Order{
		OrderType:      sonm.OrderType_BID,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		CounterpartyID: "0x0",
		Duration:       3600 + uint64(rand.Int63n(3600)),
		Price:          sonm.NewBigIntFromInt(1000 + rand.Int63n(1000)),
		Netflags:       sonm.NetflagsToUint([3]bool{true, true, (rand.Int() % 2) == 0}),
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Blacklist:      "0x0",
		Tag:            []byte("00000"),
		Benchmarks: &sonm.Benchmarks{
			Values: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}

	return order
}
