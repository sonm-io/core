package connor

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func makeTestCordersSet(f CorderFactoriy, from, to uint64) []*Corder {
	set := make([]*Corder, 0)
	for i := from; i <= to; i += 100 {
		set = append(set, f.FromParams(big.NewInt(1), i, newBenchmarksWithGPUMem(0)))
	}
	return set
}

func TestDivideOrders(t *testing.T) {
	f := NewCorderFactory("ETH", 0)

	existing := makeTestCordersSet(f, 200, 400)
	required := makeTestCordersSet(f, 100, 500)

	set := divideOrdersSets(existing, required)
	assert.Len(t, set.toRestore, 3)
	assert.Len(t, set.toCreate, 2)
	assert.Len(t, set.toCancel, 0)
}

func TestDivideOrdersWithExtra(t *testing.T) {
	f := NewCorderFactory("ETH", 0)

	tests := []struct {
		existing  [2]uint64
		required  [2]uint64
		toRestore int
		toCreate  int
		toCancel  int
	}{
		{
			[2]uint64{200, 1000},
			[2]uint64{500, 700},
			3, 0, 6,
		},
		{
			[2]uint64{500, 700},
			[2]uint64{200, 1000},
			3, 6, 0,
		},
		{
			[2]uint64{200, 700},
			[2]uint64{500, 1000},
			3, 3, 3,
		},
		{
			[2]uint64{500, 1000},
			[2]uint64{100, 500},
			1, 4, 5,
		},
	}

	for _, tt := range tests {
		existing := makeTestCordersSet(f, tt.existing[0], tt.existing[1])
		required := makeTestCordersSet(f, tt.required[0], tt.required[1])

		set := divideOrdersSets(existing, required)
		assert.Len(t, set.toRestore, tt.toRestore)
		assert.Len(t, set.toCreate, tt.toCreate)
		assert.Len(t, set.toCancel, tt.toCancel)
	}
}
