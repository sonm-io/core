// +build linux

package miner

import (
	"fmt"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = true
	parentCgroup           = "insonmnia"
)

func initializeControlGroup() (cGroupDeleter, error) {
	// Cook or update parent cgroup for all containers
	cgroupPath := cgroups.StaticPath("/" + parentCgroup)
	control, err := cgroups.Load(cgroups.V1, cgroupPath)
	switch err {
	case nil:
		// pass
	case cgroups.ErrCgroupDeleted:
		// create an empty cgroup
		// we update it later
		control, err = cgroups.New(cgroups.V1, cgroupPath, &specs.LinuxResources{})
		if err != nil {
			return nil, fmt.Errorf("failed to create parent cgroup %s: %v", parentCgroup, err)
		}
	default:
		return nil, fmt.Errorf("failed to load parent cgroup %s: %v", parentCgroup, err)
	}

	// TODO: read resources from configuration
	var memLimit int64 = 100000000 // bytes
	err = control.Update(&specs.LinuxResources{
		Memory: &specs.LinuxMemory{
			Limit: &memLimit, // 1 GB
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set resource limit on parent cgroup %s: %v", parentCgroup, err)
	}

	return control, nil
}
