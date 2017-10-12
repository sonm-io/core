package hub

import (
	"errors"
	"sync"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
)

type AllocationStatus int

const (
	FREE AllocationStatus = iota
	RESERVED
	ALLOCATED
)

var (
	errSlotNotFree = errors.New("specified slot is not free")
)

type slotItem struct {
	slot   *structs.Slot
	status AllocationStatus
}

type Scheduler struct {
	mu    sync.Mutex
	slots []*slotItem
}

func (s *Scheduler) Exists(slot *structs.Slot) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.exists(slot)
}

func (s *Scheduler) exists(slot *structs.Slot) bool {
	return s.get(slot) != nil
}

func (s *Scheduler) All() []*structs.Slot {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := []*structs.Slot{}
	for _, item := range s.slots {
		result = append(result, item.slot)
	}

	return result
}

func (s *Scheduler) Get(slot *structs.Slot) *structs.Slot {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.get(slot).slot
}

func (s *Scheduler) get(slot *structs.Slot) *slotItem {
	for _, item := range s.slots {
		if slot.Compare(item.slot, pb.OrderType_BID) {
			return item
		}
	}
	return nil
}

func (s *Scheduler) Add(slot *structs.Slot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.add(slot)
}

func (s *Scheduler) add(slot *structs.Slot) error {
	if s.exists(slot) {
		return errSlotAlreadyExists
	}
	s.slots = append(s.slots, &slotItem{slot: slot, status: FREE})
	return nil
}

func (s *Scheduler) Reserve(slot *structs.Slot) error {
	item := s.get(slot)
	if item.status != FREE {
		return errSlotNotFree
	}
	item.status = RESERVED
	return nil
}
