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
		res := calculateEthPrice(tt.price, tt.reward, tt.difficulty, 1)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}

func TestPriceCalculationWithMargin(t *testing.T) {
	price := big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(500))
	reward := big.NewFloat(3)
	difficulty := big.NewFloat(3e15)

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

	for _, tt := range tests {
		res := calculateEthPrice(price, reward, difficulty, tt.margin)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}

func TestCalculateXmrPrice(t *testing.T) {
	tests := []struct {
		price      *big.Float
		reward     *big.Float
		difficulty *big.Float
		expected   *big.Int
	}{
		{
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(100)),
			reward:     big.NewFloat(3),
			difficulty: big.NewFloat(5e+10),
			expected:   big.NewInt(6000000000),
		},
		{
			price:      big.NewFloat(0).Mul(big.NewFloat(params.Ether), big.NewFloat(0.001)),
			reward:     big.NewFloat(50),
			difficulty: big.NewFloat(5e+15),
			expected:   big.NewInt(10),
		},
	}

	for _, tt := range tests {
		res := calculateXmrPrice(tt.price, tt.reward, tt.difficulty, 1)
		assert.True(t, tt.expected.Cmp(res) == 0, fmt.Sprintf("%v == %v", tt.expected.String(), res.String()))
	}
}
