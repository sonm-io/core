package miner

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
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
logging:
  level: warn
benchmarks:
  url: "http://localhost.dev/list.json"
`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, zapcore.WarnLevel, conf.LogLevel())
	assert.Equal(t, "/var/lib/sonm/worker.boltdb", conf.Storage().Endpoint)
	assert.Equal(t, "sonm", conf.Storage().Bucket)
}

func TestConfigPluginsDefaults(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
benchmarks:
  url: "http://localhost.dev/list.json"
`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, "/run/docker/plugins", conf.Plugins().SocketDir)
	assert.Equal(t, "/var/lib/docker-volumes", conf.Plugins().Volumes.Root)
}

func TestConfigPlugins(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002"]
benchmarks:
  url: "http://localhost.dev/list.json"

plugins:
  socket_dir: /tmp/run/test-plugins
  volume:
    root: /my/random/dir
    volumes:
      cifs: {}
      webdav: {}
      ipfs: {}

  gpus:
    radeon: {}
    nvidia: {}
`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, "/tmp/run/test-plugins", conf.Plugins().SocketDir)
	assert.Equal(t, "/my/random/dir", conf.Plugins().Volumes.Root)

	assert.Len(t, conf.Plugins().Volumes.Volumes, 3)
	assert.Contains(t, conf.Plugins().Volumes.Volumes, "cifs")
	assert.Contains(t, conf.Plugins().Volumes.Volumes, "webdav")
	assert.Contains(t, conf.Plugins().Volumes.Volumes, "ipfs")

	assert.Len(t, conf.Plugins().GPUs, 2)
	assert.Contains(t, conf.Plugins().GPUs, "nvidia")
	assert.Contains(t, conf.Plugins().GPUs, "radeon")
}
