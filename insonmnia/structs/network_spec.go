package structs

import (
	"errors"
	"strings"

	"github.com/pborman/uuid"
	"github.com/sonm-io/core/proto"
)

type Network interface {
	// ID returns a unique identifier that will be used as a new network name.
	ID() string
	// NetworkType returns a network driver name used to establish networking.
	NetworkType() string
	// NetworkOptions return configuration map, passed directly to network driver, this map should not be mutated.
	NetworkOptions() map[string]string
	// Returns network subnet in CIDR notation if applicable.
	NetworkCIDR() string
	// Returns specified addr to join the network.
	NetworkAddr() string
}

type NetworkSpec struct {
	*sonm.NetworkSpec
	NetID string
}

func (n *NetworkSpec) ID() string {
	return n.NetID
}

func (n *NetworkSpec) NetworkType() string {
	return n.GetType()
}

func (n *NetworkSpec) NetworkOptions() map[string]string {
	return n.GetOptions()
}

func (n *NetworkSpec) NetworkCIDR() string {
	return n.GetSubnet()
}

func (n *NetworkSpec) NetworkAddr() string {
	return n.GetAddr()
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

func NewNetworkSpecs(specs []*sonm.NetworkSpec) ([]Network, error) {
	result := make([]Network, 0, len(specs))
	for _, s := range specs {
		spec, err := NewNetworkSpec(s)
		if err != nil {
			return nil, err
		}
		result = append(result, spec)
	}
	return result, nil
}
