package main

import (
	"context"
	"sync"

	"github.com/yandex/pandora/core"
)

type AmmoType uint8

type Ammo interface {
	core.Ammo

	// Type describes an ammo type used for pooling.
	// The returned value must be unique across the entire application and also
	// must be constant.
	Type() AmmoType
	Execute(ctx context.Context, ext interface{}) error
}

type AmmoFactory interface {
	New(id int) Ammo
	NewDefault() Ammo
	Pool() *sync.Pool
}

type PoolAmmoFactory struct {
	pool *sync.Pool
}

func newPoolAmmoFactory(fn func() interface{}) PoolAmmoFactory {
	return PoolAmmoFactory{
		pool: &sync.Pool{New: fn},
	}
}

func (m *PoolAmmoFactory) Pool() *sync.Pool {
	return m.pool
}
