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
cluster:
  endpoint: ":10001"
locator:
  endpoint: "127.0.0.1:9090"
market:
  endpoint: "127.0.0.1:9095"`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testHubConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, ":10002", conf.Endpoint)
	assert.Equal(t, ":10001", conf.Cluster.Endpoint)
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
  endpoint: "127.0.0.1:9090"
market:
  endpoint: "127.0.0.1:9095"`

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
  endpoint: "127.0.0.1:9090"
market:
  endpoint: "127.0.0.1:9095"`

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
monitoring:
  endpoint: ":10001"
locator:
  endpoint: ""`)
	assert.Nil(t, err)

	defer deleteTestConfigFile()

	conf, err := NewConfig(testHubConfigPath)
	assert.Error(t, err)
	assert.Nil(t, conf)
}

func TestLoadConfigLocatorPeriod(t *testing.T) {
	err := createTestConfigFile(`
ethereum:
  private_key: "1000000000000000000000000000000000000000000000000000000000000000"
endpoint: ":10002"
monitoring:
  endpoint: ":10001"
market:
  endpoint: "127.0.0.1:9095"
locator:
  endpoint: "127.0.0.1:9090"
  period: 500
  `)
	assert.Nil(t, err)

	defer deleteTestConfigFile()

	conf, err := NewConfig(testHubConfigPath)
	assert.NoError(t, err)
	assert.Equal(t, conf.Locator.Period, 500)
}

func TestParseEndpoints(t *testing.T) {
	assert := assert.New(t)
	var fixtures = []struct {
		Endpoint    string
		ExpectError bool
	}{
		{":10000", false},
		{"0.0.0.0:10000", false},
		{"aaaa:50000", true},
	}

	for _, fixture := range fixtures {
		result, err := parseEndpoints(&ClusterConfig{Endpoint: fixture.Endpoint})
		if fixture.ExpectError {
			assert.NotNil(err)
		} else {
			assert.NoError(err)
			assert.NotEmpty(result)
		}
	}
}
