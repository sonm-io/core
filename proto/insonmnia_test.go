package sonm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToPrice(t *testing.T) {
	tests := []struct {
		in       string
		expected *big.Int
	}{
		{
			in:       "100 SNM/s",
			expected: big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
		},
		{
			in:       "2 SNM/h",
			expected: big.NewInt(0).Quo(big.NewInt(params.Ether), big.NewInt(1800)),
		},
	}

	for _, tt := range tests {
		pr := Price{}
		err := pr.LoadFromString(tt.in)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, pr.PerSecond.Unwrap())
	}
}
