package miner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResources(t *testing.T) {
	assertions := assert.New(t)
	defer deleteTestConfigFile()
	raw := `
hub:
  eth_addr: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD"
  endpoints: ["127.0.0.1:10002", "127.0.0.1:10002"]
  resources:
    cgroup: insonmnia
    resources:
      memory:
        limit: 1000
        swap: 1024
      cpu:
        quota: 1024
        cpus: "ddd"
locator:
  endpoint: "8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090"
logging:
  level: debug
`
	err := createTestConfigFile(raw)
	assertions.NoError(err)

	conf, err := NewConfig(testMinerConfigPath)
	assertions.NoError(err)

	res := conf.HubResources()
	assertions.NotNil(res)
	assertions.NotNil(res.Resources)
	assertions.NotNil(res.Resources.Memory)
	assertions.Equal(int64(1000), *res.Resources.Memory.Limit)
	assertions.Equal(int64(1024), *res.Resources.Memory.Swap)
	assertions.Equal(int64(1024), *res.Resources.CPU.Quota)
	assertions.Equal("ddd", res.Resources.CPU.Cpus)
}
