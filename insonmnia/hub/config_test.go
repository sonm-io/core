package hub

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testHubConfigPath = "test_hub.yaml"
)

func createTestConfigFile(body string) error {
	return ioutil.WriteFile(testHubConfigPath, []byte(body), 0600)
}

func deleteTestConfigFile() {
	os.Remove(testHubConfigPath)
}

func TestLoadConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `hub:
  grpc_endpoint: ":10001"
  miner_endpoint: ":10002"`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, ":10001", conf.Hub.GRPCEndpoint)
	assert.Equal(t, ":10002", conf.Hub.MinerEndpoint)
}

func TestLoadConfigWithBootnodes(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `hub:
  grpc_endpoint: ":10001"
  miner_endpoint: ":10002"
  bootnodes:
    - "enode://node1"
    - "enode://node2"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Len(t, conf.Hub.Bootnodes, 2)
}

func TestLoadInvalidConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `hub:
  grpc_endpoint: ""
  miner_endpoint: ":10002"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "GRPCEndpoint is required")
}

func TestLoadConfigLogger(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `logger:
  level: -1
hub:
  grpc_endpoint: ":10001"
  miner_endpoint: ":10002"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, -1, conf.Logger.Level)
}

func TestLoadConfigLoggerDefault(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `hub:
  grpc_endpoint: ":10001"
  miner_endpoint: ":10002"`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, 1, conf.Logger.Level)
}
