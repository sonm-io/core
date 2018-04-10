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
price: 23.73 SNM/h

blacklist: 0x8125721c2413d99a33e351e1f6bb4e56b6b633fd

resources:
  cpu:
    cores: 150
  ram:
    size: 2gb
  storage: 
    size: 10gb
  gpu:
    devices: [3, 5]
  network:
    throughputin: 25 mb/s
    throughputout: 40 mb/s
    overlay: true
    outbound: true
    incoming: true
`)
	require.NoError(t, err)
	defer deleteTestConfigFile(askPlanTestFile)

	ask, err := LoadAskPlan(askPlanTestFile)
	require.NoError(t, err)
	require.NotNil(t, ask)

	expectedPrice := big.NewInt(0).Mul(big.NewInt(2373), big.NewInt(1e16))
	expectedPrice = big.NewInt(0).Quo(expectedPrice, big.NewInt(3600))
	assert.Equal(t, expectedPrice, ask.GetPrice().GetPerSecond().Unwrap())

	assert.Equal(t, time.Duration(time.Hour*8), ask.Duration.Unwrap())
	assert.Equal(t, common.HexToAddress("0x8125721c2413d99a33e351e1f6bb4e56b6b633fd").Hex(), ask.GetBlacklist().GetAddress())

	assert.Equal(t, uint64(2147483648), ask.GetResources().GetRAM().GetSize().GetSize())
	assert.Equal(t, uint64(10737418240), ask.GetResources().GetStorage().GetSize().GetSize())
	assert.Equal(t, uint64(150), ask.GetResources().GetCPU().GetCores())

	assert.Len(t, ask.Resources.GPU.Devices, 2)
	assert.Contains(t, ask.Resources.GPU.Devices, uint64(3))
	assert.Contains(t, ask.Resources.GPU.Devices, uint64(5))

	assert.Equal(t, uint64(26214400), ask.Resources.GetNetwork().GetThroughputIn().GetSize())
	assert.Equal(t, uint64(41943040), ask.Resources.GetNetwork().GetThroughputOut().GetSize())
	assert.True(t, ask.Resources.GetNetwork().Overlay)
	assert.True(t, ask.Resources.GetNetwork().Outbound)
	assert.True(t, ask.Resources.GetNetwork().Incoming)
}
