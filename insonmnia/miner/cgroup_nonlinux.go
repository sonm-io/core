// +build !linux

package miner

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = false
)

func initializeControlGroup(name string, resources *specs.LinuxResources) (cGroup, error) {
	return &nilCgroup{}, nil
}
