package miner

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGPUConfig(t *testing.T) {
	err := createTestConfigFile(`
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
logging:
  level: -1
GPUConfig:
  type: nvidiadocker
  args: { nvidiadockerdriver: "localhost:3476" }
locator:
  endpoint: "127.0.0.1:9090"
`)
	assert.NoError(t, err)
	defer deleteTestConfigFile()

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)
	assert.Equal(t, "nvidiadocker", conf.GPU().Type)
	assert.NotEmpty(t, conf.GPU().Args)
	assert.Equal(t, "localhost:3476", conf.GPU().Args["nvidiadockerdriver"])
}
