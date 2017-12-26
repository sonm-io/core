package sonm

import (
	"fmt"
	"math/big"
)

// NewPrice constructs a new price using specified big.Int.
func NewPrice(v *big.Int) *Price {
	return &Price{
		Neg: v.Sign() < 0,
		Abs: v.Bytes(),
	}
}

// NewPriceFromString tries to construct a new price from the specified string.
func NewPriceFromString(s string) (*Price, error) {
	v := new(big.Int)
	v, ok := v.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert %s to big.Int", s)
	}
	return NewPrice(v), nil
}

// BigInt returns the current price as *big.Int.
func (m *Price) BigInt() *big.Int {
	v := new(big.Int)
	v.SetBytes(m.Abs)
	if m.Neg {
		v.Neg(v)
	}
	return v
}
