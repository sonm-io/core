package structs

import (
	"errors"

	"github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
)

var (
	// The CPU CFS scheduler period in nanoseconds. Used alongside CPU quota.
	defaultCPUPeriod     = int64(100000)
	errResourcesRequired = errors.New("resources field is required")
)

type TaskResources struct {
	inner *pb.TaskResourceRequirements
}

func NewTaskResources(r *pb.TaskResourceRequirements) (*TaskResources, error) {
	if r == nil {
		return nil, errResourcesRequired
	}

	return &TaskResources{inner: r}, nil
}

func (r *TaskResources) ToUsage() resource.Resources {
	numGPUs := -1
	switch r.inner.GetGPUSupport() {
	case pb.GPUCount_NO_GPU:
		numGPUs = 0
	case pb.GPUCount_SINGLE_GPU:
		numGPUs = 1
	default:
	}

	return resource.Resources{
		NumCPUs: int(r.inner.GetCPUCores()),
		Memory:  r.inner.GetMaxMemory(),
		NumGPUs: numGPUs,
	}
}

func (r *TaskResources) ToContainerResources(cgroupParent string) container.Resources {
	return container.Resources{
		CgroupParent: cgroupParent,
		CPUQuota:     r.cpuQuota(),
		CPUPeriod:    defaultCPUPeriod,
		Memory:       r.inner.GetMaxMemory(),
	}
}

func (r *TaskResources) ToCgroupResources() *specs.LinuxResources {
	maxMemory := r.inner.GetMaxMemory()

	return &specs.LinuxResources{
		Memory: &specs.LinuxMemory{
			Limit: &maxMemory,
		},
	}
}

func (r *TaskResources) cpuQuota() int64 {
	return defaultCPUPeriod * int64(r.inner.GetCPUCores())
}
