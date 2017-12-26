package structs

import (
	"math/big"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_GetTotalPrice(t *testing.T) {
	proto := &sonm.Order{
		PricePerSecond: sonm.NewBigInt(big.NewInt(42)),
		Slot: &sonm.Slot{
			Duration:  600,
			Resources: &sonm.Resources{},
		},
	}
	order, err := NewOrder(proto)

	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, big.NewInt(42*600), order.GetTotalPrice())
}
