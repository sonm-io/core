// +build !linux

package cgroups

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = false
)

func initializeControlGroup(name string, resources *specs.LinuxResources) (CGroup, error) {
	return &nilCgroup{}, nil
}
