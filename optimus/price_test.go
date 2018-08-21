package optimus

import (
	"math/big"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelativePriceThresholdExceeds(t *testing.T) {
	v, err := NewRelativePriceThreshold(1.0)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.True(t, v.Exceeds(big.NewInt(1010), big.NewInt(1000)))
	assert.False(t, v.Exceeds(big.NewInt(1009), big.NewInt(1000)))
}

func TestParseRelativePriceThreshold(t *testing.T) {
	v, err := ParseRelativePriceThreshold("2.0%")
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.True(t, v.Exceeds(big.NewInt(1020), big.NewInt(1000)))
	assert.False(t, v.Exceeds(big.NewInt(1019), big.NewInt(1000)))
}

func TestErrParseRelativePriceThreshold(t *testing.T) {
	tests := []struct {
		name      string
		threshold string
	}{
		{
			name:      "missing percent",
			threshold: "1.5",
		},
		{
			name:      "trailing percent",
			threshold: "1.5%1",
		},
		{
			name:      "not number",
			threshold: "1.5c%",
		},
		{
			name:      "zero number",
			threshold: "0.0%",
		},
		{
			name:      "negative number",
			threshold: "-1.5%",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v, err := ParseRelativePriceThreshold(test.threshold)

			require.Error(t, err)
			require.Nil(t, v)
		})
	}
}

func TestAbsolutePriceThresholdExceeds(t *testing.T) {
	p := &sonm.Price{}
	err := p.LoadFromString("0.02 USD/h")
	require.NoError(t, err)

	v, err := NewAbsolutePriceThreshold(p)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.True(t, v.Exceeds(big.NewInt(105555555555556), big.NewInt(100000000000000)))
	assert.False(t, v.Exceeds(big.NewInt(105555555555554), big.NewInt(100000000000000)))
}
