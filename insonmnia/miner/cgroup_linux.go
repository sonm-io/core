// +build linux

package miner

import (
	"fmt"

	"github.com/containerd/cgroups"
	"github.com/mitchellh/mapstructure"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	platformSupportCGroups = true
	parentCgroup           = "insonmnia"
)

// Resources is a type alias for OCI Resources spec
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

func initializeControlGroup(res *Resources) (cGroupDeleter, error) {
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

	if res == nil {
		return control, nil
	}

	if err = control.Update((*specs.LinuxResources)(res)); err != nil {
		return nil, fmt.Errorf("failed to set resource limit on parent cgroup %s: %v", parentCgroup, err)
	}

	return control, nil
}
