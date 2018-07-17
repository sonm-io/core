package connor

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertingToUSDBalance(t *testing.T) {
	tests := []struct {
		balancePow int64   // amount of tokens that will be multiplied to 1e18 (Ether).
		sonmPrice  float64 // SONM token price in USD (like 0.12$ per 1e18 SNM)
		expected   float64 // expected amount of money in USD (like 4.8$ per 40 SNM when token price is 0.12$)
	}{
		{
			balancePow: 40,
			sonmPrice:  0.12,
			expected:   4.8,
		},
		{
			balancePow: 1e18,
			sonmPrice:  0.5,
			expected:   5e17,
		},
		{
			balancePow: 100,
			sonmPrice:  20,
			expected:   2000,
		},
		{
			balancePow: 63452,
			sonmPrice:  1.2356734,
			expected:   78405.9485768,
		},
	}

	m := &ProfitableModule{}

	for _, tt := range tests {
		m1 := big.NewInt(1e18)
		m2 := big.NewInt(tt.balancePow)

		in := big.NewInt(0).Mul(m1, m2)
		v := m.ConvertSNMBalanceToUSD(in, tt.sonmPrice)
		assert.Equal(t, v, tt.expected)
	}
}
