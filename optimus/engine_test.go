package optimus

import (
	"sort"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func newTestPlan(price int64) *sonm.AskPlan {
	return &sonm.AskPlan{
		Price: &sonm.Price{
			PerSecond: sonm.NewBigIntFromInt(price),
		},
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := map[string]struct {
		create []*sonm.AskPlan
		remove []*sonm.AskPlan

		expectedCreate []*sonm.AskPlan
		expectedRemove []*sonm.AskPlan
	}{
		"OneEqRemoveEmpty": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(3),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(2),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(3),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"OneEqRemoveNotEmpty": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(3),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(3),
				newTestPlan(4),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
			},
		},
		"TwoEq": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(2),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(4),
				newTestPlan(4),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(4),
			},
		},
		"ThreeCompletelyEq": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{},
			expectedRemove: []*sonm.AskPlan{},
		},
		"ThreeEqExtraRemove": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(4),
			},
		},
		"TwoEqExtraCreate": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"CombinedNoRemoveLeft": {
			create: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(1),
				newTestPlan(2),
				newTestPlan(2),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(1),
				newTestPlan(1),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(2),
			},
			expectedRemove: []*sonm.AskPlan{},
		},
		"Combined": {
			create: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(3),
				newTestPlan(5),
				newTestPlan(4),
				newTestPlan(1),
				newTestPlan(4),
			},
			remove: []*sonm.AskPlan{
				newTestPlan(4),
				newTestPlan(2),
				newTestPlan(2),
				newTestPlan(3),
				newTestPlan(3),
			},
			expectedCreate: []*sonm.AskPlan{
				newTestPlan(1),
				newTestPlan(4),
				newTestPlan(5),
			},
			expectedRemove: []*sonm.AskPlan{
				newTestPlan(2),
				newTestPlan(3),
			},
		},
	}

	sortPlans := func(plans []*sonm.AskPlan) {
		sort.Slice(plans, func(i, j int) bool {
			return plans[i].Price.PerSecond.Cmp(plans[j].Price.PerSecond) < 0
		})
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			create, remove := removeDuplicates(test.create, test.remove)

			sortPlans(create)
			sortPlans(remove)
			sortPlans(test.expectedCreate)
			sortPlans(test.expectedRemove)

			assert.Equal(t, test.expectedCreate, create)
			assert.Equal(t, test.expectedCreate, create)
		})
	}
}
