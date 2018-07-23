package sonm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestName(t *testing.T) {
	into := struct {
		Level IdentityLevel
	}{}

	input := []byte(`level: registered`)
	err := yaml.Unmarshal(input, &into)

	require.NoError(t, err)
	assert.Equal(t, IdentityLevel_REGISTERED, into.Level)
}

func TestBidOrderValidate(t *testing.T) {
	bid := &BidOrder{Tag: "this-string-is-too-long-for-tag-value"}
	err := bid.Validate()
	require.Error(t, err)

	bid.Tag = "short-and-valid"
	err = bid.Validate()
	require.NoError(t, err)
}

func TestNewBenchmarksFromMap(t *testing.T) {
	m1 := map[string]uint64{
		"cpu-sysbench-multi":  1,
		"cpu-sysbench-single": 2,
		"cpu-cores":           3,
		"ram-size":            4,
		"storage-size":        5,
		"net-download":        6,
		"net-upload":          7,
		"gpu-count":           8,
		"gpu-mem":             9,
		"gpu-eth-hashrate":    10,
		"gpu-cash-hashrate":   11,
		"gpu-redshift":        12,
	}

	b1, err := NewBenchmarksFromMap(m1)
	require.NoError(t, err)
	assert.Equal(t, b1.CPUSysbenchMulti(), uint64(1))
	assert.Equal(t, b1.CPUSysbenchOne(), uint64(2))
	assert.Equal(t, b1.CPUCores(), uint64(3))
	assert.Equal(t, b1.RAMSize(), uint64(4))
	assert.Equal(t, b1.StorageSize(), uint64(5))
	assert.Equal(t, b1.NetTrafficIn(), uint64(6))
	assert.Equal(t, b1.NetTrafficOut(), uint64(7))
	assert.Equal(t, b1.GPUCount(), uint64(8))
	assert.Equal(t, b1.GPUMem(), uint64(9))
	assert.Equal(t, b1.GPUEthHashrate(), uint64(10))
	assert.Equal(t, b1.GPUCashHashrate(), uint64(11))
	assert.Equal(t, b1.GPURedshift(), uint64(12))
}

func TestNewBenchmarksFromMapMissing(t *testing.T) {
	_, err := NewBenchmarksFromMap(map[string]uint64{})
	require.Error(t, err)

	_, err = NewBenchmarksFromMap(nil)
	require.Error(t, err)
}
