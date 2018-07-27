package connor

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCorderFromParams(t *testing.T) {
	c1, err := NewCorderFromParams("ETH", big.NewInt(100), 1000,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 3000e6}))
	require.NoError(t, err)

	assert.Equal(t, c1.GetHashrate(), uint64(1000))
	assert.Equal(t, c1.Order.Benchmarks.GPUMem(), uint64(3000e6))
	assert.Equal(t, c1.Order.GetBenchmarks().GPUEthHashrate(), uint64(1000))

	c2, err := NewCorderFromParams("ZEC", big.NewInt(100), 500,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 900e6}))
	require.NoError(t, err)

	assert.Equal(t, c2.GetHashrate(), uint64(500))
	assert.Equal(t, c2.Order.Benchmarks.GPUMem(), uint64(900e6))
	assert.Equal(t, c2.Order.GetBenchmarks().GPUCashHashrate(), uint64(500))

	c3, err := NewCorderFromParams("NULL", big.NewInt(100), 5000,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 1e6}))
	require.NoError(t, err)

	assert.Equal(t, c3.GetHashrate(), uint64(5000))
	assert.Equal(t, c3.Order.Benchmarks.GPUMem(), uint64(1e6))
	assert.Equal(t, c3.Order.GetBenchmarks().GPURedshift(), uint64(5000))
}

func TestCorder_AsBID(t *testing.T) {
	eth, err := NewCorderFromParams("ETH", big.NewInt(100), 1000,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 3000e6}))
	require.NoError(t, err)

	zec, err := NewCorderFromParams("ZEC", big.NewInt(100), 130,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 900e6}))
	require.NoError(t, err)

	null, err := NewCorderFromParams("NULL", big.NewInt(100), 550,
		newBenchmarkFromMap(map[string]uint64{"gpu-mem": 1e6}))
	require.NoError(t, err)

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
