package price

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestCalculateEthPrice(t *testing.T) {
	prov := &ethPriceProvider{}

	tests := []struct {
		price      *big.Float
		reward     *big.Float
		difficulty *big.Float
		expected   *big.Int
	}{
		{
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(465)),
			reward:     big.NewFloat(2.91),
			difficulty: big.NewFloat(3.28042614076814e+15),
			expected:   big.NewInt(412492),
		},
		{
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(100)),
			reward:     big.NewFloat(2.91),
			difficulty: big.NewFloat(3.28042614076814e+15),
			expected:   big.NewInt(88707),
		},
		{
			// what if the net difficulty will be highly increased and a token price will be dramatically low?
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(3)),
			reward:     big.NewFloat(2.91),
			difficulty: big.NewFloat(3.5e17),
			expected:   big.NewInt(24),
		},
		{
			// what if the net difficulty will be slightly increased,
			// the reward will be also increased, but the token price
			// will be a lot higher than now?
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(50000)),
			reward:     big.NewFloat(3),
			difficulty: big.NewFloat(3.88042614076814e+15),
			expected:   big.NewInt(38655548),
		},
	}

	for _, tt := range tests {
		res := prov.calculate(tt.price, tt.reward, tt.difficulty)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}