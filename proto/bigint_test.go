package sonm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBigInt(t *testing.T) {
	v := big.NewInt(42000000000)
	price := NewBigInt(v)

	assert.Equal(t, v, price.Unwrap())
}

func TestNewBigIntFromString(t *testing.T) {
	price, err := NewBigIntFromString("42000000001")

	require.NotNil(t, price)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(42000000001), price.Unwrap())
}

func TestBigIntString(t *testing.T) {
	price := NewBigInt(big.NewInt(42000000002))
	assert.Equal(t, "42000000002", price.Unwrap().String())
}
