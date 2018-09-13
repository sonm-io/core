package types

import (
	"github.com/sonm-io/core/proto"
)

type TaskStatus struct {
	*sonm.TaskStatusReply
	ID string
}

type OrdersSets struct {
	Create  []*Corder
	Restore []*Corder
	Cancel  []*Corder
}

func DivideOrdersSets(existingCorders, targetCorders []*Corder) *OrdersSets {
	existingByBenchmark := map[uint64]*Corder{}
	for _, ord := range existingCorders {
		existingByBenchmark[ord.GetHashrate()] = ord
	}

	targetByBenchmark := map[uint64]*Corder{}
	for _, ord := range targetCorders {
		targetByBenchmark[ord.GetHashrate()] = ord
	}

	set := &OrdersSets{
		Create:  make([]*Corder, 0),
		Restore: make([]*Corder, 0),
		Cancel:  make([]*Corder, 0),
	}

	for _, ord := range targetCorders {
		if ex, ok := existingByBenchmark[ord.GetHashrate()]; ok {
			if ex.hash() == ord.hash() {
				set.Restore = append(set.Restore, ex)
			} else {
				set.Cancel = append(set.Cancel, ex)
				set.Create = append(set.Create, ord)
			}
		} else {
			set.Create = append(set.Create, ord)
		}
	}

	for _, ord := range existingCorders {
		// order is exists on market but shouldn't be presented
		// in the target orders set.
		if _, ok := targetByBenchmark[ord.GetHashrate()]; !ok {
			set.Cancel = append(set.Cancel, ord)
		}
	}

	return set
}
