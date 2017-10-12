package hub

import "github.com/sonm-io/core/insonmnia/structs"

type ETH interface {
	// CreatePendingDeal creates a new pending deal.
	CreatePendingDeal(order *structs.Order) error
	// RevokePendingDeal revokes a pending deal.
	RevokePendingDeal(order *structs.Order) error
}

// TODO (3Hren): Currently stub. Need integration with ETH.
type eth struct {
}

func (e *eth) CreatePendingDeal(order *structs.Order) error {
	return nil
}

func (e *eth) RevokePendingDeal(order *structs.Order) error {
	return nil
}

// NewETH constructs a new Ethereum client.
func NewETH() (ETH, error) {
	return &eth{}, nil
}
