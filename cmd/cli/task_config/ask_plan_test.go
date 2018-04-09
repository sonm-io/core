package task_config

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const askPlanTestFile = "ask_test.yaml"

func TestLoadAskPlan(t *testing.T) {
	err := createTestConfigFile(askPlanTestFile, `
duration: 8h
price_per_hour: 23.73

blacklist: 0x8125721c2413d99a33e351e1f6bb4e56b6b633fd

resources:
  cpu:
    cores: 1.5
  ram:
    size: 2gb
  storage: 
    size: 10gb
  gpu:
    devices: [3, 5]
  net:
    throughput_in: 1000
    throughput_out: 3000
    overlay: true
    outbound: true
    incoming: true
`)
	require.NoError(t, err)
	defer deleteTestConfigFile(askPlanTestFile)

	ask, err := LoadAskPlan(askPlanTestFile)
	require.NoError(t, err)
	assert.NotNil(t, ask)

	expectedPrice := big.NewInt(0).Mul(big.NewInt(2373), big.NewInt(1e16))
	assert.Equal(t, expectedPrice, ask.PricePerHour.Price())

	assert.Equal(t, time.Duration(time.Hour*8), ask.Duration)
	assert.Equal(t, common.HexToAddress("0x8125721c2413d99a33e351e1f6bb4e56b6b633fd"), ask.Blacklist.Address())

	assert.Equal(t, uint64(2147483648), ask.Resources.RAM.Size.Bytes())
	assert.Equal(t, uint64(10737418240), ask.Resources.Storage.Size.Bytes())
	assert.Equal(t, uint64(150), ask.Resources.CPU.Cores.Count())

	assert.Len(t, ask.Resources.GPU.Devices, 2)
	assert.Contains(t, ask.Resources.GPU.Devices, uint64(3))
	assert.Contains(t, ask.Resources.GPU.Devices, uint64(5))

	assert.Equal(t, uint64(1000), ask.Resources.Net.ThroughputIn)
	assert.Equal(t, uint64(3000), ask.Resources.Net.ThroughputOut)
	assert.True(t, ask.Resources.Net.Overlay)
	assert.True(t, ask.Resources.Net.Outbound)
	assert.True(t, ask.Resources.Net.Incoming)
}

/* todo: test for value bounds. */
