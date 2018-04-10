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
		{`333 MB\s`, "333 mb"},
		{"444mb/s", "444mb"},
		{"555 Mb\\S", "555 mb"},
	}

	for _, tt := range tests {
		out := trimTimeRate(tt.in)
		assert.Equal(t, tt.out, out, fmt.Sprintf("ex: %s | out: %s", tt.out, out))
	}
}

func TestExtractPriceRe(t *testing.T) {
	tests := []struct {
		in     string
		val    string
		timesz string
	}{
		{
			in:     "100 SNM/h",
			val:    "100",
			timesz: "h",
		},
		{
			in:     "100snm/H",
			val:    "100",
			timesz: "h",
		},
		{
			in:     "2 snm/s",
			val:    "2",
			timesz: "s",
		},
		{
			in:     "500 snm\\H",
			val:    "500",
			timesz: "h",
		},
		{
			in:     `1 SNM/S`,
			val:    "1",
			timesz: "s",
		},
	}

	for _, tt := range tests {
		val, tsz, err := extractPricePerTimeValues(tt.in)
		require.NoError(t, err)
		assert.Equal(t, tt.val, val)
		assert.Equal(t, tt.timesz, tsz)
	}

	_, _, err := extractPricePerTimeValues("snm/s")
	require.Error(t, err)

	_, _, err = extractPricePerTimeValues("")
	require.Error(t, err)

	_, _, err = extractPricePerTimeValues("123 eth/s")
	require.Error(t, err)
}

func TestConvertToPrice(t *testing.T) {
	tests := []struct {
		val      string
		dim      string
		expected *big.Int
	}{
		{
			val:      "100",
			dim:      "s",
			expected: big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
		},
		{
			val:      "2",
			dim:      "h",
			expected: big.NewInt(0).Quo(big.NewInt(params.Ether), big.NewInt(1800)),
		},
	}

	for _, tt := range tests {
		p, err := convertToPrice(tt.val, tt.dim)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, p.Unwrap())
	}
}
