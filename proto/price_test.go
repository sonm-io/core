package sonm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrice(t *testing.T) {
	v := big.NewInt(42000000000)
	price := NewPrice(v)

	assert.Equal(t, v, price.BigInt())
}

func TestNewPriceFromString(t *testing.T) {
	price, err := NewPriceFromString("42000000001")

	require.NotNil(t, price)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(42000000001), price.BigInt())
}

func TestPriceString(t *testing.T) {
	price := NewPrice(big.NewInt(42000000002))

	assert.Equal(t, "42000000002", price.BigInt().String())
}
