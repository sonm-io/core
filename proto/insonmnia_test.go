package sonm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrimRate(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"111 mb/s", "111 mb"},
		{"222 MB/S", "222 mb"},
		{"333mb/s", "333mb"},
	}

	for _, tt := range tests {
		out := trimTimeRate(tt.in)
		assert.Equal(t, tt.out, out, fmt.Sprintf("ex: %s | out: %s", tt.out, out))
	}
}

func TestConvertToPrice(t *testing.T) {
	tests := []struct {
		in       string
		expected *big.Int
	}{
		{
			in:       "100 snm/s",
			expected: big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
		},
		{
			in:       "2 snm/h",
			expected: big.NewInt(0).Quo(big.NewInt(params.Ether), big.NewInt(1800)),
		},
	}

	for _, tt := range tests {
		pr := Price{}
		p, err := pr.parse(tt.in)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, p.Unwrap())
	}
}
