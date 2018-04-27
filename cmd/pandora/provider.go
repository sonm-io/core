package main

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/yandex/pandora/core"
)

const PoolSize = math.MaxUint8 + 1

type provider struct {
	channel chan Ammo
	pools   [PoolSize]*sync.Pool

	ammoLimit     int
	ammoFactories []AmmoFactory
}

func newProvider(pools [PoolSize]*sync.Pool, ammoLimit int, ammoFactories []AmmoFactory) core.Provider {
	return &provider{
		channel:       make(chan Ammo, 128),
		pools:         pools,
		ammoLimit:     ammoLimit,
		ammoFactories: ammoFactories,
	}
}

func (m *provider) Acquire() (core.Ammo, bool) {
	ammo, ok := <-m.channel
	return ammo, ok
}

func (m *provider) Release(ammo core.Ammo) {
	a := ammo.(Ammo)
	m.pools[a.Type()].Put(a)
}

func (m *provider) Run(ctx context.Context) error {
	defer close(m.channel)

	for id := 0; id < m.ammoLimit; id++ {
		select {
		case m.channel <- m.ammoFactories[id%len(m.ammoFactories)].New(id):
		case <-ctx.Done():
			break
		}
	}

	return ctx.Err()
}

type Config struct {
	AmmoLimit int                      `config:"limit"`
	Select    string                   `config:"select"`
	Detail    []map[string]interface{} `config:"detail"`
}

func NewProvider(cfg Config) (core.Provider, error) {
	var factories []AmmoFactory
	for id, ammoCfg := range cfg.Detail {
		ty, ok := ammoCfg["type"]
		if !ok {
			return nil, fmt.Errorf("invalid `%d` ammo desciption: field `type` is required", id)
		}

		ammoType, ok := ty.(string)
		if !ok {
			return nil, fmt.Errorf("`%d` ammo type should be a string: %T instead", id, ty)
		}

		delete(ammoCfg, "type")
		factory, err := AmmoRegistry.Get(ammoType, ammoCfg)
		if err != nil {
			return nil, err
		}

		factories = append(factories, factory)
	}

	if len(factories) == 0 {
		return nil, fmt.Errorf("no ammo was specified")
	}

	ids := map[AmmoType]struct{}{}
	for id, factory := range factories {
		ammo := factory.NewDefault()
		ty := ammo.Type()
		_, ok := ids[ty]
		if ok {
			return nil, fmt.Errorf("`%d` ammo type already registered for %T", id, ammo)
		}

		ids[ty] = struct{}{}
	}

	pools := [PoolSize]*sync.Pool{}
	for id := range factories {
		pools[factories[id].NewDefault().Type()] = factories[id].Pool()
	}

	return newProvider(pools, cfg.AmmoLimit, factories), nil
}
