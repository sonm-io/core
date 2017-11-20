package hub

import (
	"context"
)

type ETH interface {
	// CheckContract checks whether a given contract exists.
	// TODO: Don't know what exactly to pass for now.
	CheckContract(interface{}) (bool, error)
}

// TODO: Currently stub. Need integration with ETH.
type eth struct {
}

func (e *eth) CheckContract(interface{}) (bool, error) {
	return true, nil
}

// NewETH constructs a new Ethereum client.
func NewETH(ctx context.Context) (ETH, error) {
	return &eth{}, nil
}
