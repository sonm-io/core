package price

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestCalculateEthPrice(t *testing.T) {
	tests := []struct {
		price      *big.Int
		reward     float64
		difficulty float64
		expected   *big.Int
	}{
		{
			price:      big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(465)),
			reward:     2.91,
			difficulty: 3.28042614076814e+15,
			expected:   big.NewInt(412492),
		},
		{
			price:      big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
			reward:     2.91,
			difficulty: 3.28042614076814e+15,
			expected:   big.NewInt(88707),
		},
		{
			// what if the net difficulty will be highly increased and a token price will be dramatically low?
			price:      big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(3)),
			reward:     2.91,
			difficulty: 3.5e17,
			expected:   big.NewInt(24),
		},
		{
			// what if the net difficulty will be slightly increased,
			// the reward will be also increased, but the token price
			// will be a lot higher than now?
			price:      big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(50000)),
			reward:     3,
			difficulty: 3.88042614076814e+15,
			expected:   big.NewInt(38655548),
		},
	}

	for _, tt := range tests {
		coin := &coinParams{Difficulty: tt.difficulty, BlockReward: tt.reward}
		res := calculateEthPrice(tt.price, coin, 1)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}

func TestPriceCalculationWithMargin(t *testing.T) {
	tests := []struct {
		margin   float64
		expected *big.Int
	}{
		{
			margin:   0.9,
			expected: big.NewInt(450000),
		},
		{
			margin:   1,
			expected: big.NewInt(500000),
		},
		{
			margin:   1.1,
			expected: big.NewInt(550000),
		},
		{
			margin:   10,
			expected: big.NewInt(5000000),
		},
		{
			margin:   0.1,
			expected: big.NewInt(50000),
		},
		{
			margin:   1.123,
			expected: big.NewInt(561500),
		},
	}

	price := big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(500))
	for _, tt := range tests {
		coin := &coinParams{Difficulty: 3e15, BlockReward: 3}
		res := calculateEthPrice(price, coin, tt.margin)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}

func TestCalculateXmrPrice(t *testing.T) {
	tests := []struct {
		price    *big.Int
		reward   float64
		netHash  float64
		expected *big.Int
	}{
		{
			price:    big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
			reward:   4,
			netHash:  563816182,
			expected: big.NewInt(5912092344),
		},
		{
			price:    big.NewInt(params.Finney),
			reward:   4,
			netHash:  563816182,
			expected: big.NewInt(59120),
		},
	}

	for _, tt := range tests {
		coin := &coinParams{BlockReward: tt.reward, Nethash: int(tt.netHash)}
		res := calculateXmrPrice(tt.price, coin, 1)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}
