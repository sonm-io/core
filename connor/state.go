package connor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/sonm-io/core/connor/types"
	"go.uber.org/zap"
)

type dealQuality struct {
	Deal   *types.Deal `json:"deal"`
	ByLogs float64     `json:"by_logs"`
	ByPool float64     `json:"by_pool"`
}

type state struct {
	mu           sync.Mutex
	log          *zap.Logger
	activeOrders map[string]*types.Corder
	queuedOrders map[string]*types.Corder
	deals        map[string]*dealQuality
}

func NewState(log *zap.Logger) *state {
	return &state{
		log:          log.Named("state"),
		activeOrders: map[string]*types.Corder{},
		queuedOrders: map[string]*types.Corder{},
		deals:        map[string]*dealQuality{},
	}
}

func (s *state) HasOrder(order *types.Corder) bool {
	id := order.GetId().Unwrap().String()
	hash := order.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.activeOrders[id]
	if !ok {
		_, ok = s.queuedOrders[hash]
	}
	return ok
}

func (s *state) HasDeal(deal *types.Deal) bool {
	id := deal.GetId().Unwrap().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.deals[id]
	return ok
}

func (s *state) AddActiveOrder(ord *types.Corder) {
	id := ord.GetId().Unwrap().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeOrders[id] = ord
	s.log.Debug("active order added",
		zap.String("order_id", id),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("creating", len(s.queuedOrders)))
}

func (s *state) AddQueuedOrder(ord *types.Corder) {
	hash := ord.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.queuedOrders[hash] = ord
	s.log.Debug("queued order added",
		zap.String("order_hash", hash),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) DeleteActiveOrder(ord *types.Corder) {
	id := ord.GetId().Unwrap().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeOrders, id)
	s.log.Debug("active order deleted",
		zap.String("order_id", id),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) DeleteQueuedOrder(ord *types.Corder) {
	hash := ord.Hash()

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.queuedOrders, hash)
	s.log.Debug("queued order deleted",
		zap.String("order_hash", hash),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) AddDeal(deal *types.Deal) {
	id := deal.GetId().Unwrap().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.deals[id] = &dealQuality{
		Deal:   deal,
		ByLogs: 1,
		ByPool: 1,
	}
	s.log.Debug("deal added", zap.String("deal_id", id), zap.Int("total", len(s.deals)))
}

func (s *state) DeleteDeal(deal *types.Deal) {
	id := deal.GetId().Unwrap().String()

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.deals, id)
	s.log.Debug("deal deleted", zap.String("deal_id", id), zap.Int("total", len(s.deals)))
}

func (s *state) dump() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return json.Marshal(struct {
		ActiveOrders map[string]*types.Corder
		QueuedOrders map[string]*types.Corder
		Deals        map[string]*dealQuality
	}{
		ActiveOrders: s.activeOrders,
		QueuedOrders: s.queuedOrders,
		Deals:        s.deals,
	})
}

func (s *state) DumpToFile() error {
	data, err := s.dump()
	if err != nil {
		return fmt.Errorf("failed to maarshal storage state: %v", err)
	}

	return ioutil.WriteFile("/tmp/connor_state.json", data, 0600)
}
