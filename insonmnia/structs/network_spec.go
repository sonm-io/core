package structs

import (
	"fmt"

	"github.com/pkg/errors"
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
	Id string
}

func (n *NetworkSpec) ID() string {
	return n.Id
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
	if spec.Type == "tinc" {
		if len(spec.Addr) == 0 || len(spec.Subnet) == 0 {
			return errors.New("address and subnet are required for tinc driver")
		}
	}
	return nil
}

func NewNetworkSpec(id string, spec *sonm.NetworkSpec) (*NetworkSpec, error) {
	err := validateNetworkSpec(id, spec)
	if err != nil {
		return nil, err
	}
	return &NetworkSpec{spec, id}, nil
}

func NewNetworkSpecs(idPrefix string, specs []*sonm.NetworkSpec) ([]Network, error) {
	result := make([]Network, 0, len(specs))
	for i, s := range specs {
		spec, err := NewNetworkSpec(idPrefix+"__"+fmt.Sprint(i), s)
		if err != nil {
			return nil, err
		}
		result = append(result, spec)
	}
	return result, nil
}
