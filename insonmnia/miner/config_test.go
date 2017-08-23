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
  endpoint: 127.0.0.1`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, "127.0.0.1", conf.HubEndpoint())
}

func TestLoadEmptyConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
hub:
  endpoint: ""`

	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)
	assert.Equal(t, conf.HubEndpoint(), "")
}

func TestLoadConfigWithoutEndpoint(t *testing.T) {
	defer deleteTestConfigFile()
	err := createTestConfigFile("")
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.NoError(t, err)
	assert.Nil(t, conf.HubResources())
	assert.Empty(t, conf.HubEndpoint())
}
