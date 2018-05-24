package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResources(t *testing.T) {
	assertions := assert.New(t)
	defer deleteTestConfigFile()
	raw := `
master: 0x0000000000000000000000000000000000000001
endpoint: ":0"
resources:
  cgroup: insonmnia
  resources:
    memory:
      limit: 1000
      swap: 1024
    cpu:
      quota: 1024
      cpus: "ddd"
logging:
  level: debug
`
	err := createTestConfigFile(raw)
	assertions.NoError(err)

	conf, err := NewConfig(testWorkerConfigPath)
	assertions.NoError(err)

	res := conf.Resources
	assertions.NotNil(res)
	assertions.NotNil(res.Resources)
	assertions.NotNil(res.Resources.Memory)
	assertions.Equal(int64(1000), *res.Resources.Memory.Limit)
	assertions.Equal(int64(1024), *res.Resources.Memory.Swap)
	assertions.Equal(int64(1024), *res.Resources.CPU.Quota)
	assertions.Equal("ddd", res.Resources.CPU.Cpus)
}
