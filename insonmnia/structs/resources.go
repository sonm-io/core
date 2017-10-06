package structs

import (
	pb "github.com/sonm-io/core/proto"
)

type Resources struct {
	inner *pb.Resources
}

func NewResources(resources *pb.Resources) (*Resources, error) {
	return nil, nil
}

func ValidateResources(resources *pb.Resources) error {
	if resources == nil {
		return errResourcesIsNil
	}
	return nil
}

func (r *Resources) Eq(o *Resources) bool {
	return false
}
