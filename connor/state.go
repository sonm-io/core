package connor

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/sonm-io/core/connor/types"
	"go.uber.org/zap"
)

type dealQuality struct {
	Deal   *types.Deal
	ByLogs float64
	ByPool float64
}

type state struct {
	mu           sync.Mutex
	log          *zap.Logger
	activeOrders map[string]*types.Corder
	queuedOrders map[string]*types.Corder
	deals        map[string]*dealQuality
}

func newState(log *zap.Logger) *state {
	return &state{
		log:          log.Named("state"),
		activeOrders: map[string]*types.Corder{},
		queuedOrders: map[string]*types.Corder{},
		deals:        map[string]*dealQuality{},
	}
}

func (s *state) hasOrder(order *types.Corder) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := order.GetId().Unwrap().String()
	hash := order.Hash()

	_, ok := s.activeOrders[id]
	if !ok {
		_, ok = s.queuedOrders[hash]
	}
	return ok
}

func (s *state) hasDeal(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.deals[id]
	return ok
}

func (s *state) addActiveOrder(ord *types.Corder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.activeOrders[ord.GetId().Unwrap().String()] = ord
	s.log.Debug("adding active order",
		zap.String("order_id", ord.GetId().Unwrap().String()),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("creating", len(s.queuedOrders)))
}

func (s *state) addQueuedOrder(ord *types.Corder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := ord.Hash()
	s.queuedOrders[hash] = ord
	s.log.Debug("adding queued order",
		zap.String("order_hash", hash),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) deleteActiveOrder(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeOrders, id)
	s.log.Debug("deleting active order",
		zap.String("order_id", id),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) deleteCreatingOrder(ord *types.Corder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hash := ord.Hash()
	delete(s.queuedOrders, hash)
	s.log.Debug("deleting queued order",
		zap.String("order_hash", hash),
		zap.Int("active", len(s.activeOrders)),
		zap.Int("queued", len(s.queuedOrders)))
}

func (s *state) addDeal(deal *types.Deal) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deals[deal.GetId().Unwrap().String()] = &dealQuality{
		Deal:   deal,
		ByLogs: 1,
		ByPool: 1,
	}
	s.log.Debug("adding deal",
		zap.String("deal_id", deal.GetId().Unwrap().String()),
		zap.Int("total", len(s.deals)))
}

func (s *state) deleteDeal(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Debug("deleting deal", zap.String("deal_id", id), zap.Int("total", len(s.deals)))
	delete(s.deals, id)
}

func (s *state) dump() {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := json.Marshal(struct {
		ActiveOrders map[string]*types.Corder
		QueuedOrders map[string]*types.Corder
		Deals        map[string]*dealQuality
	}{
		ActiveOrders: s.activeOrders,
		QueuedOrders: s.queuedOrders,
		Deals:        s.deals,
	})

	if err != nil {
		s.log.Warn("cannot marshal state", zap.Error(err))
		return
	}

	s.log.Info("dumpling state to the disk")
	if err := ioutil.WriteFile("/tmp/connor_state.json", b, 0600); err != nil {
		s.log.Warn("cannot write state data ", zap.Error(err))
	}
}
