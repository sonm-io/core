package util

import (
	"fmt"
	"math/big"
	"net"
	"testing"

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

func TestIsPrivateIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		isPriv bool
	}{
		{ip: "127.0.0.1", isPriv: true},
		{ip: "127.1.2.3", isPriv: true},
		{ip: "10.20.30.40", isPriv: true},
		{ip: "192.168.0.1", isPriv: true},
		{ip: "172.16.0.1", isPriv: true},
		{ip: "169.254.0.1", isPriv: true},
		{ip: "169.254.1.0", isPriv: true},
		{ip: "169.254.123.222", isPriv: true},
		{ip: "169.254.255.255", isPriv: true},
		{ip: "224.0.0.1", isPriv: true},

		{ip: "1.2.3.4", isPriv: false},
		{ip: "0.0.0.0", isPriv: false},
	}

	for _, cc := range tests {
		b := isPrivateIPv4(net.ParseIP(cc.ip))
		assert.Equal(t, b, cc.isPriv, cc.ip)
	}
}

func TestIsPrivateIPv6(t *testing.T) {
	tests := []struct {
		ip     string
		isPriv bool
	}{
		{ip: "::", isPriv: false},
		{ip: "::1", isPriv: true},
		{ip: "fe80::", isPriv: true},
		{ip: "feaf::", isPriv: true},
		{ip: "fc00::", isPriv: true},
	}

	for _, cc := range tests {
		b := isPrivateIPv6(net.ParseIP(cc.ip))
		assert.Equal(t, b, cc.isPriv, cc.ip)
	}
}

func TestStringToEtherPrice(t *testing.T) {
	tests := []struct {
		in       string
		out      float64
		mustFail bool
	}{
		{
			// value is too low
			in:       "0.0000000000000000001",
			mustFail: true,
		},
		{
			in:  "10000000000000000000000",
			out: 1e40,
		},
		{
			in:  "1000000000000",
			out: 1e30,
		},
		{
			in:  "1",
			out: 1e18,
		},
		{
			in:  "0.1",
			out: 1e17,
		},
		{
			in:  "0.00000001",
			out: 1e10,
		},
		{
			in:       "-1",
			out:      0,
			mustFail: true,
		},
		{
			in:       "-10000000000000000",
			out:      -1e34,
			mustFail: true,
		},
		{
			in:       "",
			mustFail: true,
		},
		{
			in:       "-",
			mustFail: true,
		},
		{
			in:       "099",
			out:      99e18,
			mustFail: false,
		},
		{
			in:       "-099",
			out:      -99e18,
			mustFail: true,
		},
		{
			in:  "0xff",
			out: 255e18,
		},
		{
			in:       "    1",
			out:      0,
			mustFail: true,
		},
		{
			in:       "1    ",
			out:      0,
			mustFail: true,
		},
	}

	for _, tt := range tests {
		out, err := StringToEtherPrice(tt.in)
		if !tt.mustFail {
			f, _ := big.NewFloat(tt.out).Int(nil)
			assert.True(t, out.Cmp(f) == 0, fmt.Sprintf("expect %s == %s", tt.in, out.String()))
		} else {
			assert.Error(t, err, fmt.Sprintf("test must fail for value %s", tt.in))
		}
	}
}
