package structs

import (
	"reflect"

	pb "github.com/sonm-io/core/proto"
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
	if !reflect.DeepEqual(r.inner.GetProps(), o.inner.GetProps()) {
		return false
	}
	return true
}
