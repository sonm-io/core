package hub

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
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
market:
  endpoint: "127.0.0.1:9095"
logging:
  level: info
`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, ":10002", conf.Endpoint)
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
  level: info
market:
  endpoint: "127.0.0.1:9095"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, zapcore.InfoLevel, conf.LogLevel())
}

func TestLoadConfigInvalidLogLevel(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
monitoring:
  endpoint: ":10001"
market:
  endpoint: "127.0.0.1:9095"
logging:
  level: wtf
`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	_, err = NewConfig(testHubConfigPath)
	assert.Error(t, err)
}
