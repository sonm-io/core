// +build linux

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
  resources: {
        Memory: {Limit: 1000, Swap: 1024 },
        CPU: {Quota: 1024, Cpus: "ddd"}
      }`
	err := createTestConfigFile(raw)
	assert.NoError(err)

	conf, err := NewConfig(testMinerConfigPath)
	assert.NoError(err)

	res := conf.HubResources()
	assert.NotNil(res)
	assert.NotNil(res.Memory)
	assert.Equal(int64(1000), *res.Memory.Limit)
	assert.Equal(int64(1024), *res.Memory.Swap)
	assert.Equal(int64(1024), *res.CPU.Quota)
	assert.Equal("ddd", res.CPU.Cpus)
}
