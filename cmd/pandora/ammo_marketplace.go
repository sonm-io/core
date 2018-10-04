package main

import (
	"context"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
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
	order, err := mExt.market.PlaceOrder(ctx, mExt.privateKey, order())
	if err != nil {
		return err
	}

	mExt.log.Debug("OK", zap.String("ammo", "PlaceOrder"), zap.Any("order", *order))

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
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		CounterpartyID: sonm.NewEthAddress(common.HexToAddress("0x0")),
		Duration:       3600 + uint64(rand.Int63n(3600)),
		Price:          sonm.NewBigIntFromInt(1000 + rand.Int63n(1000)),
		Netflags:       &sonm.NetFlags{},
		IdentityLevel:  sonm.IdentityLevel_ANONYMOUS,
		Blacklist:      "0x0",
		Tag:            []byte("00000"),
		Benchmarks: &sonm.Benchmarks{
			Values: []uint64{
				uint64(rand.Int63n(20000)),      // cpu-sysbench-multi
				uint64(rand.Int63n(20000)),      // cpu-sysbench-one
				1 + uint64(rand.Int63n(16)),     // sys-cores
				1e9 + uint64(rand.Int63n(1e10)), // size-ram
				0,                               //uint64(rand.Int63n(1e12)),       // size-stor
				uint64(rand.Int63n(1e3)),        // download-net
				uint64(rand.Int63n(1e3)),        // upload-net
				0,                               //1 + uint64(rand.Int63n(16)),     // count-gpu
				0,                               //1e9 + uint64(rand.Int63n(1e11)), // mem-gpu
				0,                               //uint64(rand.Int63n(1e9)),
				0,                               //uint64(rand.Int63n(1e9)),
				0,                               //uint64(rand.Int63n(1e9)),
			},
		},
	}
	order.Netflags.SetOverlay(true)
	order.Netflags.SetIncoming(true)
	order.Netflags.SetOutbound(rand.Int()%2 == 0)

	return order
}
