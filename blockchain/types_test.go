package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalGasPriceEmpty(t *testing.T) {
	price := GasPrice{}

	assert.Error(t, price.UnmarshalText([]byte("")))
}

func TestUnmarshalGasPriceInvalid(t *testing.T) {
	price := GasPrice{}

	assert.Error(t, price.UnmarshalText([]byte("wat")))
}

func TestUnmarshalGasPriceValidWei(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20")))
	assert.Equal(t, "20", price.String())
}

func TestUnmarshalGasPriceValidWeiWithUnit(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20wei")))
	assert.Equal(t, "20", price.String())
}

func TestUnmarshalGasPriceValidWeiWithUnitAndWhitespace(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20 wei")))
	assert.Equal(t, "20", price.String())
}

func TestUnmarshalGasPriceValidWeiWithUnitInUppercase(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20 WEI")))
	assert.Equal(t, "20", price.String())
}

func TestUnmarshalGasPriceValidGwei(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20gwei")))
	assert.Equal(t, "20000000000", price.String())
}

func TestUnmarshalGasPriceValidGweiRaw(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("20000000000")))
	assert.Equal(t, "20000000000", price.String())
}

func TestUnmarshalGasPriceValidEther(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("42 ether")))
	assert.Equal(t, "42000000000000000000", price.String())
}

func TestUnmarshalGasPriceValidGEther(t *testing.T) {
	price := GasPrice{}

	require.NoError(t, price.UnmarshalText([]byte("42 gether")))
	assert.Equal(t, "42000000000000000000000000000", price.String())
}
