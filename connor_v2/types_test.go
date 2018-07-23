package connor

import (
	"math/big"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCorderFromParams(t *testing.T) {
	c1, err := NewCorderFromParams("ETH", big.NewInt(100), 1000)
	require.NoError(t, err)

	assert.Equal(t, c1.GetHashrate(), uint64(1000))
	assert.Equal(t, c1.Order.GetBenchmarks().GPUEthHashrate(), uint64(1000))

	c2, err := NewCorderFromParams("ZEC", big.NewInt(100), 500)
	require.NoError(t, err)

	assert.Equal(t, c2.GetHashrate(), uint64(500))
	assert.Equal(t, c2.Order.GetBenchmarks().GPUCashHashrate(), uint64(500))
}
