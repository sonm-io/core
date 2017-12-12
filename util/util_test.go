package util

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		mustErr  bool
	}{
		{
			input:    "192.168.0.1:10001",
			expected: "10001",
			mustErr:  false,
		},
		{
			input:    ":10002",
			expected: "10002",
			mustErr:  false,
		},
		{
			input:    "192.168.0.1",
			expected: "",
			mustErr:  true,
		},
		{
			input:    "192.168.0.1:qwer",
			expected: "",
			mustErr:  true,
		},
		{
			input:    "192.168.0.1:99999",
			expected: "",
			mustErr:  true,
		},
	}

	for _, tt := range tests {
		port, err := ParseEndpointPort(tt.input)
		assert.Equal(t, tt.expected, port)
		if tt.mustErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestEqualAddresses(t *testing.T) {
	cases := []struct {
		a    string
		b    string
		isEq bool
	}{
		{
			a:    "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			b:    "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab",
			isEq: true,
		},
		{
			a:    "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			b:    "0x1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab",
			isEq: true,
		},
		{
			a:    "1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			b:    "2aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			isEq: false,
		},
		{
			a:    "0x1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			b:    "0x1aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaB",
			isEq: true,
		},
		{
			a:    "0x0",
			b:    "0x1",
			isEq: false,
		},
		{
			a:    "0",
			b:    "1",
			isEq: false,
		},
		{
			a:    "0x",
			b:    "0x",
			isEq: true,
		},
	}

	for _, cc := range cases {
		a := common.HexToAddress(cc.a)
		b := common.HexToAddress(cc.b)
		assert.Equal(t, cc.isEq, EqualAddresses(a, b), fmt.Sprintf("compare %s and %s failed", cc.a, cc.b))
	}
}
