package main

import (
	"fmt"
	"sync"
)

var AmmoRegistry = ammoRegistry{
	registry: map[string]func(cfg interface{}) (AmmoFactory, error){},
}

type ammoRegistry struct {
	mu       sync.RWMutex
	registry map[string]func(cfg interface{}) (AmmoFactory, error)
}

func (m *ammoRegistry) Register(ty string, factory func(cfg interface{}) (AmmoFactory, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.registry[ty] = factory
}

func (m *ammoRegistry) Get(ty string, cfg interface{}) (AmmoFactory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factory, ok := m.registry[ty]
	if !ok {
		return nil, fmt.Errorf("`%s` ammo factory not registered", ty)
	}

	return factory(cfg)
}
