package hub

type Market interface {
	// OrderExists checks whether an order with the specified ID exists in the
	// marketplace.
	OrderExists(ID string) (bool, error)
}

// TODO (3Hren): Currently stub. Need integration with market.
type market struct {
}

func (m *market) OrderExists(ID string) (bool, error) {
	return true, nil
}

// NewMarket constructs a new SONM marketplace client.
func NewMarket() (Market, error) {
	return &market{}, nil
}
