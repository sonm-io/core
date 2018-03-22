package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	errSlotNotExists     = errors.New("specified slot does not exist")
	errSlotAlreadyExists = errors.New("specified slot already exists")
	errOrderNotExists    = errors.New("specified order does not exist")
	errForbiddenMiner    = errors.New("miner is forbidden")
)

type OrderID string

func (id OrderID) String() string {
	return string(id)
}

// MinerCtx holds all the data related to a connected Miner
type MinerCtx struct {
	ctx    context.Context
	cancel context.CancelFunc

	miner *miner.Miner

	// Miner name received after handshaking.
	uuid string

	// Scheduling.

	mu           sync.Mutex
	capabilities *hardware.Hardware
	usage        *resource.Pool
	usageMapping map[OrderID]resource.Resources
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

func createMinerCtx(ctx context.Context, miner *miner.Miner) (*MinerCtx, error) {
	var err error

	var (
		m = MinerCtx{
			miner:        miner,
			usageMapping: make(map[OrderID]resource.Resources),
		}
	)
	m.ctx, m.cancel = context.WithCancel(ctx)
	if err != nil {
		log.G(ctx).Error("failed to connect to Miner's grpc server", zap.Error(err))
		m.Close()
		return nil, err
	}

	if err := m.handshake(); err != nil {
		m.Close()
		return nil, err
	}

	return &m, nil
}

// ID returns the miner id.
func (m *MinerCtx) ID() string {
	return m.uuid
}

func (m *MinerCtx) handshake() error {
	log.G(m.ctx).Info("sending handshake to a Miner")
	resp, err := m.miner.Handshake(m.ctx, &pb.MinerHandshakeRequest{})
	if err != nil {
		log.G(m.ctx).Error("failed to receive handshake from a Miner", zap.Error(err))
		return err
	}

	log.G(m.ctx).Info("received handshake from a Miner")

	capabilities, err := hardware.HardwareFromProto(resp.Capabilities)
	if err != nil {
		log.G(m.ctx).Error("failed to decode capabilities from a Miner", zap.Error(err))
		return err
	}

	log.G(m.ctx).Debug("received Miner's capabilities",
		zap.String("id", resp.Miner),
		zap.Any("capabilities", capabilities),
	)

	m.uuid = resp.Miner
	m.capabilities = capabilities
	m.usage = resource.NewPool(capabilities)

	return nil
}

// Consume consumes the specified resources from the miner.
func (m *MinerCtx) Consume(Id OrderID, usage *resource.Resources) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.consume(Id, usage)
}

func (m *MinerCtx) consume(id OrderID, usage *resource.Resources) error {
	if m.orderExists(id) {
		return fmt.Errorf("order already exists")
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

	m.usageMapping[OrderID(id)] = *usage

	return nil
}

func (m *MinerCtx) OrderExists(id OrderID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orderExists(id)
}

func (m *MinerCtx) orderExists(id OrderID) bool {
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
func (m *MinerCtx) Release(id OrderID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.releaseDeal(id)
}

func (m *MinerCtx) releaseDeal(id OrderID) {
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

func (m *MinerCtx) OrderUsage(id OrderID) (resource.Resources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.orderUsage(id)
}

func (m *MinerCtx) orderUsage(id OrderID) (resource.Resources, error) {
	usage, exists := m.usageMapping[id]
	if !exists {
		return resource.Resources{}, errOrderNotExists
	}

	return usage, nil
}

// Orders returns a list of allocated orders.
// Useful for looking for a proper miner for starting tasks.
func (m *MinerCtx) Orders() []OrderID {
	m.mu.Lock()
	defer m.mu.Unlock()
	orders := []OrderID{}
	for id := range m.usageMapping {
		orders = append(orders, id)
	}
	return orders
}

// Close frees all connections related to a Miner
func (m *MinerCtx) Close() {
	m.cancel()
}
