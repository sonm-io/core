package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/auth"
	"go.uber.org/zap"
)

type ReservedOrder struct {
	OrderID          OrderID
	MinerID          string
	EthAddr          common.Address
	ReservedFrom     time.Time
	ReservedDuration time.Duration
}

func (o *ReservedOrder) IsExpired() bool {
	return time.Now().Sub(o.ReservedFrom) >= o.ReservedDuration
}

// OrderShelter implements a hub for reserved orders.
type OrderShelter struct {
	hub    *Hub
	mu     sync.Mutex
	orders map[OrderID]ReservedOrder
}

func NewOrderShelter(hub *Hub) *OrderShelter {
	return &OrderShelter{
		hub:    hub,
		orders: make(map[OrderID]ReservedOrder, 0),
	}
}

func (s *OrderShelter) Reserve(orderID OrderID, minerID string, ethAddr common.Address, duration time.Duration) error {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.exists(orderID) {
		return fmt.Errorf("order already reserved")
	}

	s.orders[orderID] = ReservedOrder{
		OrderID:          orderID,
		MinerID:          minerID,
		EthAddr:          ethAddr,
		ReservedFrom:     now,
		ReservedDuration: duration,
	}

	return nil
}

func (s *OrderShelter) Exists(orderID OrderID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exists(orderID)
}

func (s *OrderShelter) exists(orderID OrderID) bool {
	_, ok := s.orders[orderID]
	return ok
}

func (s *OrderShelter) PollCommit(orderID OrderID, ethAddr common.Address) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found")
	}

	if auth.EqualAddresses(order.EthAddr, ethAddr) {
		return nil
	} else {
		return fmt.Errorf("order %s cannot be commited by %s", orderID, ethAddr)
	}
}

// Commit commits the specified reserved order, removing it from the shelter.
// Note, that this method does not releases resources from the miner's tracker,
// because using it means that the resource's lifetime was prolonged by
// accepting a deal.
func (s *OrderShelter) Commit(orderID OrderID) (ReservedOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return ReservedOrder{}, fmt.Errorf("order not found")
	}

	delete(s.orders, orderID)

	return order, nil
}

func (s *OrderShelter) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Marshal(s.orders)
}

func (s *OrderShelter) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	unmarshalled := make(map[OrderID]ReservedOrder, 0)
	err := json.Unmarshal(data, &unmarshalled)

	if err != nil {
		return err
	}
	s.orders = unmarshalled
	return nil
}

func (s *OrderShelter) Dump() map[OrderID]ReservedOrder {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.orders
}

func (s *OrderShelter) RestoreFrom(orders map[OrderID]ReservedOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders = orders
}

func (s *OrderShelter) Run(ctx context.Context) error {
	timer := time.NewTicker(20 * time.Second)
	defer timer.Stop()

	for {
		log.G(ctx).Debug("ticking reserved orders handler")

		select {
		case <-timer.C:
			s.turn(ctx)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *OrderShelter) turn(ctx context.Context) {
	renewedOrders := make(map[OrderID]ReservedOrder, 0)

	s.mu.Lock()
	defer s.mu.Unlock()

	for orderID, orderInfo := range s.orders {
		if orderInfo.IsExpired() {
			miner := s.hub.GetMinerByID(orderInfo.MinerID)
			if miner != nil {
				log.G(ctx).Info("releasing order due to timeout",
					zap.Stringer("orderID", orderID),
					zap.String("minerID", orderInfo.MinerID),
				)
				miner.Release(orderID)
			} else {
				log.G(ctx).Warn("unable to release order from a miner: no such miner",
					zap.Stringer("orderID", orderID),
					zap.String("minerID", orderInfo.MinerID),
				)
			}
		} else {
			renewedOrders[orderID] = orderInfo
		}
	}

	s.orders = renewedOrders
}
