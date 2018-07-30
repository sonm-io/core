package structs

import (
	"errors"
	"strings"

	"github.com/pborman/uuid"
	"github.com/sonm-io/core/proto"
)

type NetworkSpec struct {
	*sonm.NetworkSpec
	NetID string
}

func validateNetworkSpec(id string, spec *sonm.NetworkSpec) error {
	if len(spec.GetType()) == 0 {
		return errors.New("network type is required in network spec")
	}
	return nil
}

func NewNetworkSpec(spec *sonm.NetworkSpec) (*NetworkSpec, error) {
	id := strings.Replace(uuid.New(), "-", "", -1)
	err := validateNetworkSpec(id, spec)
	if err != nil {
		return nil, err
	}
	return &NetworkSpec{spec, id}, nil
}

func NewNetworkSpecs(specs []*sonm.NetworkSpec) ([]*NetworkSpec, error) {
	result := make([]*NetworkSpec, 0, len(specs))
	for _, s := range specs {
		spec, err := NewNetworkSpec(s)
		if err != nil {
			return nil, err
		}
		result = append(result, spec)
	}
	return result, nil
}
