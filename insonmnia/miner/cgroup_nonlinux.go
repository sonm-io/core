// +build !linux

package miner

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

// Resources is a placeholder for resources
type Resources interface{}

const (
	platformSupportCGroups = false
	parentCgroup           = ""
)

type nilCgroup struct{}

func (c *nilCgroup) New(name string, resources *specs.LinuxResources) (cGroup, error) {
	return c, nil
}

func (*nilCgroup) Delete() error { return nil }

func initializeControlGroup(*Resources) (cGroup, error) {
	return &nilCgroup{}, nil
}
