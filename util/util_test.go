package util

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
