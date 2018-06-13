package sonm

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestAskPlanUnmarshallers(t *testing.T) {
	data := []byte(`
duration: 8h
price: 23.73 USD/h

blacklist: 0x8125721c2413d99a33e351e1f6bb4e56b6b633fd

resources:
  cpu:
    cores: 1.50
  ram:
    size: 2GiB
  storage: 
    size: 10GiB
  gpu:
    indexes: [3, 5]
  network:
    throughputin: 25 Mibit/s
    throughputout: 40 Mbit/s
    overlay: true
    outbound: true
    incoming: true
`)
	ask := &AskPlan{}
	err := yaml.Unmarshal(data, ask)
	require.NoError(t, err)
	require.NotNil(t, ask)
	err = ask.Validate()
	require.NoError(t, err)

	expectedPrice := big.NewInt(0).Mul(big.NewInt(2373), big.NewInt(1e16))
	expectedPrice = big.NewInt(0).Quo(expectedPrice, big.NewInt(3600))
	assert.Equal(t, expectedPrice, ask.GetPrice().GetPerSecond().Unwrap())

	assert.Equal(t, time.Duration(time.Hour*8), ask.Duration.Unwrap())
	assert.Equal(t, common.HexToAddress("0x8125721c2413d99a33e351e1f6bb4e56b6b633fd").Bytes(),
		ask.GetBlacklist().GetAddress())

	assert.Equal(t, uint64(2*1024*1024*1024), ask.GetResources().GetRAM().GetSize().GetBytes())
	assert.Equal(t, uint64(10*1024*1024*1024), ask.GetResources().GetStorage().GetSize().GetBytes())
	assert.Equal(t, uint64(150), ask.GetResources().GetCPU().GetCorePercents())

	assert.Len(t, ask.Resources.GPU.Indexes, 2)
	assert.Contains(t, ask.Resources.GPU.Indexes, uint64(3))
	assert.Contains(t, ask.Resources.GPU.Indexes, uint64(5))

	assert.Equal(t, uint64(25*1024*1024), ask.Resources.GetNetwork().GetThroughputIn().GetBitsPerSecond())
	assert.Equal(t, uint64(40e6), ask.Resources.GetNetwork().GetThroughputOut().GetBitsPerSecond())
	assert.True(t, ask.Resources.GetNetwork().GetNetFlags().GetOverlay())
	assert.True(t, ask.Resources.GetNetwork().GetNetFlags().GetOutbound())
	assert.True(t, ask.Resources.GetNetwork().GetNetFlags().GetIncoming())
}

func TestAskPlanIDsAndHashes(t *testing.T) {
	data := []byte(`
resources:
  cpu:
    cores: 1.50
  ram:
    size: 2GiB
  gpu:
    indexes: [3, 5]
    hashes: ["aaa", "bbb"]
`)

	ask := &AskPlan{}
	err := yaml.Unmarshal(data, ask)
	require.NoError(t, err)
	require.NotNil(t, ask)

	err = ask.Validate()
	require.Error(t, err)
}
