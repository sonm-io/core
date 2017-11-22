package miner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResources(t *testing.T) {
	assert := assert.New(t)
	defer deleteTestConfigFile()
	raw := `
hub:
  endpoint: 0@0.0.0.0:0
  resources:
    cgroup: insonmnia
    resources:
      memory:
        limit: 1000
        swap: 1024
      cpu:
        quota: 1024
        cpus: "ddd"
`
	err := createTestConfigFile(raw)
	assert.NoError(err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.NoError(err)

	res := conf.HubResources()
	assert.NotNil(res)
	assert.NotNil(res.Resources)
	assert.NotNil(res.Resources.Memory)
	assert.Equal(int64(1000), *res.Resources.Memory.Limit)
	assert.Equal(int64(1024), *res.Resources.Memory.Swap)
	assert.Equal(int64(1024), *res.Resources.CPU.Quota)
	assert.Equal("ddd", res.Resources.CPU.Cpus)
}
