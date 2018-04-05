package hub

import (
	"encoding/json"
	"errors"
	"sync"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx   context.Context
	miner *miner.Miner

	// Scheduling.
	mu           sync.Mutex
	capabilities *hardware.Hardware
	usage        *resource.Pool
	usageMapping map[structs.OrderID]resource.Resources
}

func (m *MinerCtx) MarshalJSON() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Marshal(m.usageMapping)
}

func (m *MinerCtx) UnmarshalJSON(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return json.Unmarshal(data, &m.usageMapping)
}

// todo: drop method, init worker in an other way
// fixme: (._. )
func createMinerCtx(ctx context.Context, miner *miner.Miner) (*MinerCtx, error) {
	hw := miner.Hardware()

	m := MinerCtx{
		ctx:          ctx,
		miner:        miner,
		usageMapping: make(map[structs.OrderID]resource.Resources),
		capabilities: hw,
		usage:        resource.NewPool(hw),
	}

	// run benchmarks before "handshake" because
	// worker must fill their hardware capabilities with actual data
	err := miner.RunBenchmarks()
	if err != nil {
		return nil, err
	}

	return &m, nil
}

// Consume consumes the specified resources from the miner.
func (m *MinerCtx) Consume(id structs.OrderID, usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.consume(id, usage)
}

func (m *MinerCtx) consume(id structs.OrderID, usage *resource.Resources) error {
	if m.orderExists(id) {
		return errors.New("order already exists")
	}
	if err := m.usage.Consume(usage); err != nil {
		return err
	}

	log.G(m.ctx).Debug("consumed resources for a task",
		zap.Stringer("id", id),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	m.usageMapping[id] = *usage

	return nil
}

func (m *MinerCtx) OrderExists(id structs.OrderID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orderExists(id)
}

func (m *MinerCtx) orderExists(id structs.OrderID) bool {
	_, exists := m.usageMapping[id]
	return exists
}

func (m *MinerCtx) PollConsume(usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.G(m.ctx).Debug("checking for resources",
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	return m.usage.PollConsume(usage)
}

// Release returns back resources for the miner.
//
// Should be called when a deal has finished no matter for what reason.
func (m *MinerCtx) Release(id structs.OrderID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.releaseDeal(id)
}

func (m *MinerCtx) releaseDeal(id structs.OrderID) {
	usage, exists := m.usageMapping[id]
	if !exists {
		return
	}

	log.G(m.ctx).Debug("retained resources for a task",
		zap.Stringer("id", id),
		zap.Any("usage", usage),
		zap.Any("usageTotal", m.usage.GetUsage()),
		zap.Any("capabilities", m.capabilities),
	)

	delete(m.usageMapping, id)
	m.usage.Release(&usage)
}

func (m *MinerCtx) OrderUsage(id structs.OrderID) (resource.Resources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orderUsage(id)
}

func (m *MinerCtx) orderUsage(id structs.OrderID) (resource.Resources, error) {
	usage, exists := m.usageMapping[id]
	if !exists {
		return resource.Resources{}, errors.New("order not exists")
	}

	return usage, nil
}

// Orders returns a list of allocated orders.
// Useful for looking for a proper miner for starting tasks.
func (m *MinerCtx) Orders() []structs.OrderID {
	m.mu.Lock()
	defer m.mu.Unlock()
	orders := []structs.OrderID{}
	for id := range m.usageMapping {
		orders = append(orders, id)
	}
	return orders
}
