// +build linux

package optimus

import (
	"fmt"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/insonmnia/cgroups"
)

const (
	defaultCPUPeriod = uint64(100000)
)

func RestrictUsage(cfg *RestrictionsConfig) (Deleter, error) {
	control, _, err := cgroups.NewCgroupManager(cfg.Name, convertLinuxResources(cfg))
	if err != nil {
		return nil, err
	}

	if err := control.Add(cgroups.Process{Pid: os.Getpid()}); err != nil {
		return nil, fmt.Errorf("failed to add Optimus into cgroup: %v", err)
	}

	return control, nil
}

func convertLinuxResources(cfg *RestrictionsConfig) *specs.LinuxResources {
	return &specs.LinuxResources{
		CPU:    convertCPUConfig(cfg),
		Memory: convertMemoryConfig(cfg),
	}
}

func convertCPUConfig(cfg *RestrictionsConfig) *specs.LinuxCPU {
	quota := int64(float64(defaultCPUPeriod) * cfg.CPUCount)
	period := defaultCPUPeriod

	return &specs.LinuxCPU{
		Quota:  &quota,
		Period: &period,
	}
}

func convertMemoryConfig(cfg *RestrictionsConfig) *specs.LinuxMemory {
	if cfg.MemoryLimit == 0 {
		return nil
	}

	v := int64(cfg.MemoryLimit * (1 << 20))

	return &specs.LinuxMemory{
		Limit: &v,
	}
}
