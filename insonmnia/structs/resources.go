package structs

import (
	"reflect"

	"github.com/docker/docker/api/types/container"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
)

var (
	errResourcesRequired = errors.New("resources field is required")
)

// Resources wraps the underlying protobuf object with full validation, such
// as checking all required fields exists etc.
type Resources struct {
	inner *pb.Resources
}

// NewResources constructs a new resources wrapper using provided protobuf
// object.
func NewResources(resources *pb.Resources) (*Resources, error) {
	if err := ValidateResources(resources); err != nil {
		return nil, err
	}

	return &Resources{inner: resources}, nil
}

// Unwrap unwraps this resources yielding the underlying protobuf object.
func (r *Resources) Unwrap() *pb.Resources {
	return r.inner
}

// GetCpuCores returns the total number of logical CPU cores.
func (r *Resources) GetCpuCores() uint64 {
	return r.inner.GetCpuCores()
}

// GetMemoryInBytes returns the total number of memory bytes requested.
func (r *Resources) GetMemoryInBytes() uint64 {
	return r.inner.GetRamBytes()
}

// GetGPUCount returns the number of GPU devices required.
func (r *Resources) GetGPUCount() int {
	switch r.inner.GetGpuCount() {
	case pb.GPUCount_NO_GPU:
		return 0
	case pb.GPUCount_SINGLE_GPU:
		return 1
	case pb.GPUCount_MULTIPLE_GPU:
		return -1
	}

	return 0
}

// ValidateResources validates the specified protobuf object to be wrapped.
func ValidateResources(resources *pb.Resources) error {
	if resources == nil {
		return errResourcesIsNil
	}
	return nil
}

// Eq method tests for self and other values to be equal.
func (r *Resources) Eq(o *Resources) bool {
	if r.inner.GetCpuCores() != o.inner.GetCpuCores() {
		return false
	}
	if r.inner.GetRamBytes() != o.inner.GetRamBytes() {
		return false
	}
	if r.inner.GetGpuCount() != o.inner.GetGpuCount() {
		return false
	}
	if r.inner.GetStorage() != o.inner.GetStorage() {
		return false
	}
	if r.inner.GetNetTrafficIn() != o.inner.GetNetTrafficIn() {
		return false
	}
	if r.inner.GetNetTrafficOut() != o.inner.GetNetTrafficOut() {
		return false
	}
	if r.inner.GetNetworkType() != o.inner.GetNetworkType() {
		return false
	}
	if !reflect.DeepEqual(r.inner.GetProperties(), o.inner.GetProperties()) {
		return false
	}
	return true
}

type TaskResources struct {
	inner *pb.TaskResourceRequirements
}

func NewTaskResources(r *pb.TaskResourceRequirements) (*TaskResources, error) {
	if r == nil {
		return nil, errResourcesRequired
	}

	return &TaskResources{inner: r}, nil
}

func (r *TaskResources) RequiresGPU() bool {
	return r.inner.GetGPUSupport() != pb.GPUCount_NO_GPU
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

func (r *TaskResources) ToContainerResources() container.Resources {
	return container.Resources{
		Memory: r.inner.GetMaxMemory(),
	}
}

func (r *TaskResources) ToCgroup() *specs.LinuxResources {
	maxMemory := r.inner.GetMaxMemory()

	return &specs.LinuxResources{
		Memory: &specs.LinuxMemory{
			Limit: &maxMemory,
		},
	}
}
