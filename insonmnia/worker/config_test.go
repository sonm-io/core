package worker

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testWorkerConfigPath = "test_worker.yaml"
)

func createTestConfigFile(body string) error {
	return ioutil.WriteFile(testWorkerConfigPath, []byte(body), 0600)
}

func deleteTestConfigFile() {
	os.Remove(testWorkerConfigPath)
}

func TestLoadConfig(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
master: "0xfa5c7c6a1505d3930e4d30dd1fac5fbe7ba0c469"
endpoint: "127.0.0.5:15010"
metrics_listen_addr: "127.0.0.1:14123"
public_ip_addrs: ["12.34.56.78", "1.2.3.4"]

npp:
  rendezvous:
    endpoints:
      - 0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:14099
  relay:
    endpoints:
      - 2.3.4.5:12345
store:
  path: "/var/lib/sonm/worker.boltdb"
benchmarks:
  url: "http://localhost.dev/blist.json"
whitelist:
  url: "http://localhost.dev/wlist.json"
  enabled: true
`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testWorkerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, "127.0.0.5:15010", conf.Endpoint)
	assert.Equal(t, "127.0.0.1:14123", conf.MetricsListenAddr)
	assert.Contains(t, conf.PublicIPs, "12.34.56.78")
	assert.Contains(t, conf.PublicIPs, "1.2.3.4")

	assert.Len(t, conf.NPP.Rendezvous.Endpoints, 1)
	assert.Contains(t, conf.NPP.Rendezvous.Endpoints[0].String(),
		"0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:14099")

	assert.Len(t, conf.NPP.Relay.Endpoints, 1)
	assert.Contains(t, conf.NPP.Relay.Endpoints[0], "2.3.4.5:12345")

	assert.Equal(t, "/var/lib/sonm/worker.boltdb", conf.Storage.Endpoint)
	assert.Equal(t, "sonm", conf.Storage.Bucket)

	assert.Equal(t, "http://localhost.dev/blist.json", conf.Benchmarks.URL)
	assert.Equal(t, "http://localhost.dev/wlist.json", conf.Whitelist.Url)
	assert.Equal(t, true, *conf.Whitelist.Enabled)
	assert.Equal(t, "0xFa5c7C6a1505d3930e4D30Dd1faC5Fbe7ba0C469", conf.Master.Hex())
}

func TestConfigPlugins(t *testing.T) {
	defer deleteTestConfigFile()
	raw := `
endpoint: "127.0.0.5:15010"
master: 0x0000000000000000000000000000000000000001
plugins:
  socket_dir: /tmp/run/test-plugins
  volume:
    root: /my/random/dir
    drivers:
      cifs: {}
      webdav: {}
      ipfs: {}

  gpus:
    radeon: {}
    nvidia: {}

  overlay:
    drivers:
      tinc:
        enabled: true
      l2tp:
        enabled: true
`
	err := createTestConfigFile(raw)
	assert.Nil(t, err)

	conf, err := NewConfig(testWorkerConfigPath)
	assert.Nil(t, err)

	assert.Equal(t, "/tmp/run/test-plugins", conf.Plugins.SocketDir)
	assert.Equal(t, "/my/random/dir", conf.Plugins.Volumes.Root)

	assert.Len(t, conf.Plugins.Volumes.Drivers, 3)
	assert.Contains(t, conf.Plugins.Volumes.Drivers, "cifs")
	assert.Contains(t, conf.Plugins.Volumes.Drivers, "webdav")
	assert.Contains(t, conf.Plugins.Volumes.Drivers, "ipfs")

	assert.Len(t, conf.Plugins.GPUs, 2)
	assert.Contains(t, conf.Plugins.GPUs, "nvidia")
	assert.Contains(t, conf.Plugins.GPUs, "radeon")

	assert.True(t, conf.Plugins.Overlay.Drivers.L2TP.Enabled)
	assert.True(t, conf.Plugins.Overlay.Drivers.Tinc.Enabled)
}
