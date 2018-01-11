package auth

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, cc.isEq, equalAddresses(a, b), fmt.Sprintf("compare %s and %s failed", cc.a, cc.b))
	}
}
