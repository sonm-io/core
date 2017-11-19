// +build linux

package miner

import (
	"fmt"
	"path/filepath"

	"github.com/containerd/cgroups"
	"github.com/mitchellh/mapstructure"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = true
)

type cgroup struct {
	cgroups.Cgroup
	suffix string
}

func (c *cgroup) New(name string, resources *specs.LinuxResources) (cGroup, error) {
	control, err := c.Cgroup.New(name, resources)
	if err != nil {
		return nil, err
	}

	return &cgroup{control, filepath.Join(c.suffix, name)}, nil
}

func (c *cgroup) Suffix() string {
	return c.suffix
}

// CGroups is a type alias for OCI CGroups spec
type Resources specs.LinuxResources

// SetYAML implements goyaml.Setter
func (r *Resources) SetYAML(tag string, value interface{}) bool {
	// specs-go provides structures with 'json' tagged fields
	// but Yaml requires 'yaml' tag
	// value is expected to be map[interface{}]interface{}
	if tag != "!!map" {
		return false
	}

	cfg := mapstructure.DecoderConfig{
		Result:           r,
		TagName:          "json",
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(&cfg)
	if err != nil {
		return false
	}

	if decoder.Decode(value) != nil {
		return false
	}

	return true
}

func initializeControlGroup(name string, resources *specs.LinuxResources) (cGroup, error) {
	// Cook or update parent cgroup for all containers we spawn.
	cgroupPath := cgroups.StaticPath(name)
	control, err := cgroups.Load(cgroups.V1, cgroupPath)
	switch err {
	case nil:
	case cgroups.ErrCgroupDeleted:
		// Create an empty cgroup, we update it later.
		control, err = cgroups.New(cgroups.V1, cgroupPath, &specs.LinuxResources{})
		if err != nil {
			return nil, fmt.Errorf("failed to create parent cgroup %s: %v", name, err)
		}
	default:
		return nil, fmt.Errorf("failed to load parent cgroup %s: %v", name, err)
	}

	if resources == nil {
		return &cgroup{control, name}, nil
	}

	if err = control.Update(resources); err != nil {
		return nil, fmt.Errorf("failed to set resource limit on parent cgroup %s: %v", name, err)
	}

	return &cgroup{control, name}, nil
}
