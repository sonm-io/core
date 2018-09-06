package connor

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func makeTestCordersSet(f CorderFactory, from, to uint64) []*Corder {
	set := make([]*Corder, 0)
	for i := from; i <= to; i += 100 {
		set = append(set, f.FromParams(big.NewInt(1), i, newBenchmarksWithGPUMem(0)))
	}
	return set
}

func TestDivideOrders(t *testing.T) {
	f := NewCorderFactory("ETH", 0, common.Address{})

	existing := makeTestCordersSet(f, 200, 400)
	required := makeTestCordersSet(f, 100, 500)

	set := divideOrdersSets(existing, required)
	assert.Len(t, set.toRestore, 3)
	assert.Len(t, set.toCreate, 2)
	assert.Len(t, set.toCancel, 0)
}

func TestDivideOrderHashed(t *testing.T) {
	f := NewCorderFactory("ETH", 0, common.Address{})

	existing := makeTestCordersSet(f, 200, 400)
	existing = append(existing, f.FromParams(big.NewInt(1), uint64(300), newBenchmarksWithGPUMem(100)))

	required := make([]*Corder, 0)
	for i := 100; i <= 500; i += 100 {
		required = append(required, f.FromParams(big.NewInt(1), uint64(i), newBenchmarksWithGPUMem(100)))
	}

	set := divideOrdersSets(existing, required)
	assert.Len(t, set.toRestore, 1)
	assert.Len(t, set.toCreate, 4)
	assert.Len(t, set.toCancel, 2)
}

func TestDivideOrderHashedCounterparty(t *testing.T) {
	f := NewCorderFactory("ETH", 0, common.Address{})

	// without counterparty
	existing := makeTestCordersSet(f, 100, 500)
	required := makeTestCordersSet(f, 100, 500)
	for _, r := range required {
		r.CounterpartyID = sonm.NewEthAddress(common.HexToAddress("0xB8f5c92aDDB5e3D8e137e13868caA427EaFf1140"))
	}

	set := divideOrdersSets(existing, required)
	// should fully replace whole set on Market because counterparty was added.
	assert.Len(t, set.toRestore, 0)
	assert.Len(t, set.toCreate, 5)
	assert.Len(t, set.toCancel, 5)
}

func TestDivideOrderHashedNetFlags(t *testing.T) {
	f := NewCorderFactory("ETH", 0, common.Address{})
	existing := makeTestCordersSet(f, 100, 500)
	required := makeTestCordersSet(f, 100, 500)
	for _, r := range required {
		r.Netflags = &sonm.NetFlags{Flags: sonm.NetworkOutbound | sonm.NetworkIncoming}
	}

	set := divideOrdersSets(existing, required)
	// should fully replace whole set on Market because netflags was changed.
	assert.Len(t, set.toRestore, 0)
	assert.Len(t, set.toCreate, 5)
	assert.Len(t, set.toCancel, 5)
}

func TestDivideOrdersWithExtra(t *testing.T) {
	f := NewCorderFactory("ETH", 0, common.Address{})

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
