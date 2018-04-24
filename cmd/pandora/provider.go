package main

import (
	"context"
	"sync"

	"github.com/yandex/pandora/core"
)

type provider struct {
	channel chan interface{}
	pool    *sync.Pool

	ammoLimit   int
	ammoFactory AmmoFactory
}

func (m *provider) Acquire() (core.Ammo, bool) {
	ammo, ok := <-m.channel
	return ammo, ok
}

func (m *provider) Release(ammo core.Ammo) {
	m.pool.Put(ammo)
}

func (m *provider) Run(ctx context.Context) error {
	defer close(m.channel)

	for id := 0; id < m.ammoLimit; id++ {
		select {
		case m.channel <- m.ammoFactory.New(id):
		case <-ctx.Done():
			break
		}
	}

	return ctx.Err()
}

type OrderInfoProviderConfig struct {
	AmmoLimit int     `config:"limit"`
	OrderIDs  []int64 `config:"order_ids"`
}

func NewOrderInfoProvider() func(cfg OrderInfoProviderConfig) (core.Provider, error) {
	return func(cfg OrderInfoProviderConfig) (core.Provider, error) {
		pool := &sync.Pool{
			New: func() interface{} {
				return &OrderInfoAmmo{}
			},
		}

		m := &provider{
			channel:     make(chan interface{}, 128),
			pool:        pool,
			ammoLimit:   cfg.AmmoLimit,
			ammoFactory: newOrderInfoAmmoFactory(cfg.OrderIDs, pool),
		}

		return m, nil
	}
}

type OrderPlaceProviderConfig struct {
	AmmoLimit int `config:"limit"`
}

func NewOrderPlaceProvider() func(cfg OrderPlaceProviderConfig) (core.Provider, error) {
	return func(cfg OrderPlaceProviderConfig) (core.Provider, error) {
		pool := &sync.Pool{
			New: func() interface{} {
				return &OrderPlaceAmmo{}
			},
		}

		m := &provider{
			channel:     make(chan interface{}, 128),
			pool:        pool,
			ammoLimit:   cfg.AmmoLimit,
			ammoFactory: newOrderPlaceAmmoFactory(pool),
		}

		return m, nil
	}
}
