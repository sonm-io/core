// +build linux

package cgroups

import (
	"fmt"
	"path/filepath"

	"github.com/containerd/cgroups"
	"github.com/mitchellh/mapstructure"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = true
)

type cgroup struct {
	cgroups.Cgroup
	suffix string
}

func (c *cgroup) Add(process Process) error {
	return c.Cgroup.Add(cgroups.Process{
		Subsystem: cgroups.Name(process.Subsystem),
		Pid:       process.Pid,
		Path:      process.Path,
	})
}

func (c *cgroup) Stats() (*Stats, error) {
	cgStat, err := c.Stat(cgroups.IgnoreNotExist)
	if err != nil {
		return nil, err
	}
	if cgStat.Memory == nil {
		return &Stats{MemoryLimit: 0}, nil
	}

	return &Stats{MemoryLimit: cgStat.Memory.HierarchicalMemoryLimit}, nil
}

func (c *cgroup) New(name string, resources *specs.LinuxResources) (CGroup, error) {
	newPath := filepath.Join(c.suffix, name)
	return newCGroup(newPath, resources)
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

func newCGroup(name string, resources *specs.LinuxResources) (CGroup, error) {
	cgroupPath := cgroups.StaticPath(name)
	control, err := cgroups.Load(cgroups.V1, cgroupPath)

	switch err {
	case nil:
	case cgroups.ErrCgroupDeleted:
		// Create an empty cgroup, we update it later.
		control, err = cgroups.New(cgroups.V1, cgroupPath, &specs.LinuxResources{})
		if err != nil {
			return nil, fmt.Errorf("failed to create cgroup %s: %v", name, err)
		}
	default:
		return nil, fmt.Errorf("failed to load cgroup %s: %v", name, err)
	}

	if resources == nil {
		return &cgroup{control, name}, nil
	}

	if err = control.Update(resources); err != nil {
		return nil, fmt.Errorf("failed to set resource limit on cgroup %s: %v", name, err)
	}

	return &cgroup{control, name}, nil
}

func initializeControlGroup(name string, resources *specs.LinuxResources) (CGroup, error) {
	return newCGroup("/"+name, resources)
}
