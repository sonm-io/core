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
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
monitoring:
  endpoint: ":10001"
locator:
  address: "127.0.0.1:9090"`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, ":10002", conf.Endpoint)
	assert.Equal(t, ":10001", conf.Monitoring.Endpoint)
}

func TestLoadConfigWithBootnodes(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
bootnodes:
  - "enode://node1"
  - "enode://node2"
monitoring:
  endpoint: ":10001"
locator:
  address: "127.0.0.1:9090"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Len(t, conf.Bootnodes, 2)
}

func TestLoadInvalidConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ""
monitoring:
  endpoint: ":10002"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, conf)
	assert.Contains(t, err.Error(), "Endpoint is required")
}

func TestLoadConfigLogger(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
monitoring:
  endpoint: ":10001"
logging:
  level: -1
locator:
  address: "127.0.0.1:9090"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, -1, conf.Logging.Level)
}

func TestLoadConfigLoggerDefault(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
monitoring:
  endpoint: ":10001"
locator:
  address: "127.0.0.1:9090"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, 1, conf.Logging.Level)
}

func TestLoadConfigWithoutLocator(t *testing.T) {
	err := createTestConfigFile(`
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
bootnodes:
  - "enode://node1"
  - "enode://node2"
monitoring:
  endpoint: ":10001"
locator:
  address: ""`)
	assert.Nil(t, err)

	defer deleteTestConfigFile()

	conf, err := NewConfig(testHubConfigPath)
	assert.Error(t, err)
	assert.Nil(t, conf)
}
