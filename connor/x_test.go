package connor

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDivideOrders(t *testing.T) {
	f := NewCorderFactory("ETH", 0)

	ex1 := f.FromParams(big.NewInt(1), 100, newBenchmarksWithGPUMem(0))
	ex2 := f.FromParams(big.NewInt(2), 200, newBenchmarksWithGPUMem(0))
	ex3 := f.FromParams(big.NewInt(3), 300, newBenchmarksWithGPUMem(0))

	req0 := f.FromParams(big.NewInt(1), 50, newBenchmarksWithGPUMem(0))
	req1 := f.FromParams(big.NewInt(1), 100, newBenchmarksWithGPUMem(0))
	req2 := f.FromParams(big.NewInt(2), 200, newBenchmarksWithGPUMem(0))
	req3 := f.FromParams(big.NewInt(3), 300, newBenchmarksWithGPUMem(0))
	req4 := f.FromParams(big.NewInt(4), 400, newBenchmarksWithGPUMem(0))

	existing := []*Corder{ex1, ex2, ex3}
	required := []*Corder{req0, req1, req2, req3, req4}

	set := divideOrdersSets(existing, required)
	assert.Len(t, set.toRestore, 3)
	assert.Len(t, set.toCreate, 2)
}
