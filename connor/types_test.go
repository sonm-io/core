package connor

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ethBenchmarkIndex  = 9
	zecBenchmarkIndex  = 10
	nullBenchmarkIndex = 11
)

func newBenchmarksWithGPUMem(mem uint64) Benchmarks {
	b := Benchmarks{Values: make([]uint64, sonm.MinNumBenchmarks)}
	b.Values[8] = mem
	return b
}

func TestNewCorderFactory(t *testing.T) {
	c1 := NewCorderFactory("ETH", ethBenchmarkIndex, common.Address{}).FromParams(big.NewInt(100), 1000, newZeroBenchmarks())

	assert.Equal(t, c1.GetHashrate(), uint64(1000))
	assert.Equal(t, c1.Order.GetBenchmarks().GPUEthHashrate(), uint64(1000))

	c2 := NewCorderFactory("NULL", nullBenchmarkIndex, common.Address{}).FromParams(big.NewInt(200), 2000, newZeroBenchmarks())
	assert.Equal(t, c2.GetHashrate(), uint64(2000))
	assert.Equal(t, c2.Order.GetBenchmarks().GPURedshift(), uint64(2000))

	counterParty := common.HexToAddress("0xeE0b6a7D7EC0a03e59319E6eAeBE1D25C32fADF7")
	c3 := NewCorderFactory("NULL", nullBenchmarkIndex, counterParty).FromParams(big.NewInt(200), 2000, newZeroBenchmarks())
	assert.Equal(t, c3.CounterpartyID.Unwrap(), counterParty)
	assert.Equal(t, c3.AsBID().Counterparty.Unwrap(), counterParty)
}

func TestCorder_AsBID(t *testing.T) {
	eth := NewCorderFactory("ETH", ethBenchmarkIndex, common.Address{}).FromParams(big.NewInt(100), 1000, newBenchmarksWithGPUMem(3000e6))
	zec := NewCorderFactory("ZEC", zecBenchmarkIndex, common.Address{}).FromParams(big.NewInt(100), 130, newBenchmarksWithGPUMem(900e6))
	null := NewCorderFactory("NULL", nullBenchmarkIndex, common.Address{}).FromParams(big.NewInt(100), 550, newBenchmarksWithGPUMem(1e6))

	hashrate, ok := eth.AsBID().GetResources().GetBenchmarks()["gpu-eth-hashrate"]
	gpuMem, ok := eth.AsBID().GetResources().GetBenchmarks()["gpu-mem"]
	require.True(t, ok)
	assert.Equal(t, hashrate, uint64(1000))
	assert.Equal(t, gpuMem, uint64(3000e6))

	hashrate, ok = zec.AsBID().GetResources().GetBenchmarks()["gpu-cash-hashrate"]
	gpuMem, ok = zec.AsBID().GetResources().GetBenchmarks()["gpu-mem"]
	require.True(t, ok)
	assert.Equal(t, hashrate, uint64(130))
	assert.Equal(t, gpuMem, uint64(900e6))

	hashrate, ok = null.AsBID().GetResources().GetBenchmarks()["gpu-redshift"]
	gpuMem, ok = null.AsBID().GetResources().GetBenchmarks()["gpu-mem"]
	require.True(t, ok)
	assert.Equal(t, hashrate, uint64(550))
	assert.Equal(t, gpuMem, uint64(1e6))
}

func TestCorder_isReplaceable(t *testing.T) {
	tests := []struct {
		currentPrice  *big.Float
		newPrice      *big.Float
		delta         float64
		shouldReplace bool
	}{
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(110),
			delta:         0.10,
			shouldReplace: true,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(90),
			delta:         0.10,
			shouldReplace: true,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(109),
			delta:         0.10,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(91),
			delta:         0.10,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(101),
			delta:         0.01,
			shouldReplace: true,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(99),
			delta:         0.01,
			shouldReplace: true,
		},
	}

	for _, tt := range tests {
		result := isOrderReplaceable(tt.currentPrice, tt.newPrice, tt.delta)
		assert.Equal(t, tt.shouldReplace, result, fmt.Sprintf("%v | %v | %v", tt.currentPrice, tt.newPrice, tt.delta))
	}
}

func TestDeal_isReplaceable(t *testing.T) {
	tests := []struct {
		currentPrice  *big.Float
		newPrice      *big.Float
		delta         float64
		shouldReplace bool
	}{
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(91),
			delta:         0.1,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(90),
			delta:         0.1,
			shouldReplace: true,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(105),
			delta:         0.1,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(110),
			delta:         0.1,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(111),
			delta:         0.1,
			shouldReplace: false,
		},
		{
			currentPrice:  big.NewFloat(100),
			newPrice:      big.NewFloat(99),
			delta:         0.01,
			shouldReplace: true,
		},
	}

	for _, tt := range tests {
		result := isDealReplaceable(tt.currentPrice, tt.newPrice, tt.delta)
		assert.Equal(t, tt.shouldReplace, result, fmt.Sprintf("%v | %v | %v", tt.currentPrice, tt.newPrice, tt.delta))
	}
}
