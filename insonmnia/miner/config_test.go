package miner

import (
	"io/ioutil"
	"os"
	"testing"

	"strings"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMinerConfigPath = "test_miner.yaml"
)

func createTestConfigFile(body string) error {
	return ioutil.WriteFile(testMinerConfigPath, []byte(body), 0600)
}

func deleteTestConfigFile() {
	os.Remove(testMinerConfigPath)
}

func TestLoadConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
locator:
  endpoint: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090"`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, []string{"127.0.0.1:10002"}, conf.HubEndpoints())
}

func TestGPUConfigDefault(t *testing.T) {
	err := createTestConfigFile(`
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
logging:
  level: -1
locator:
  endpoint: "127.0.0.1:9090"
`)
	assert.NoError(t, err)
	defer deleteTestConfigFile()

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)
	assert.Equal(t, sonm.GPUVendorType_GPU_UNKNOWN, conf.GPU())
}

var gpuConfigStringTestTable = []struct {
	in  string
	out sonm.GPUVendorType
}{
	{in: "nvidia", out: sonm.GPUVendorType_NVIDIA},
	{in: "Nvidia", out: sonm.GPUVendorType_NVIDIA},
	{in: "NVIDIA", out: sonm.GPUVendorType_NVIDIA},

	{in: "radeon", out: sonm.GPUVendorType_RADEON},
	{in: "Radeon", out: sonm.GPUVendorType_RADEON},
	{in: "RADEON", out: sonm.GPUVendorType_RADEON},

	{in: "", out: sonm.GPUVendorType_GPU_UNKNOWN},
	{in: "intel", out: sonm.GPUVendorType_GPU_UNKNOWN},
	{in: "erhgserh8e5ythwuerghsdklghu", out: sonm.GPUVendorType_GPU_UNKNOWN},
}

func TestGpuConfigTypes(t *testing.T) {
	for _, tt := range gpuConfigStringTestTable {
		conf := &config{GPUConfig: tt.in}
		assert.Equal(t, tt.out, conf.GPU())
	}
}

func TestGPUConfigShouldITtimThis(t *testing.T) {
	tests := append(gpuConfigStringTestTable, []struct {
		in  string
		out sonm.GPUVendorType
	}{
		{in: "  NVIDIA", out: sonm.GPUVendorType_NVIDIA},
		{in: "  NVIDIA  ", out: sonm.GPUVendorType_NVIDIA},
		{in: "NVIDIA  ", out: sonm.GPUVendorType_NVIDIA},
		{in: "  RADEON  ", out: sonm.GPUVendorType_RADEON},
		{in: "  RADEON", out: sonm.GPUVendorType_RADEON},
		{in: "RADEON  ", out: sonm.GPUVendorType_RADEON},
	}...)

	cfgTpl := `
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
logging:
  level: -1
locator:
  endpoint: "127.0.0.1:9090"
GPUConfig: REPLACE_ME
`

	for _, tt := range tests {
		cfgText := strings.Replace(cfgTpl, "REPLACE_ME", tt.in, 1)
		err := createTestConfigFile(cfgText)
		require.NoError(t, err)

		conf, err := NewConfig(testMinerConfigPath)
		require.NoError(t, err)
		assert.Equal(t, tt.out, conf.GPU())

		deleteTestConfigFile()
	}
}
