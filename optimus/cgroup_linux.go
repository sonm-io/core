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
	control, _, err := cgroups.NewCgroupManager(cfg.Name, convertLinuxResources(cfg.CPUCount))
	if err != nil {
		return nil, err
	}

	if err := control.Add(cgroups.Process{Pid: os.Getpid()}); err != nil {
		return nil, fmt.Errorf("failed to add Optimus into cgroup: %v", err)
	}

	return control, nil
}

func convertLinuxResources(cpuCount float64) *specs.LinuxResources {
	quota := int64(float64(defaultCPUPeriod) * cpuCount)
	period := defaultCPUPeriod
	return &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota:  &quota,
			Period: &period,
		},
	}
}
