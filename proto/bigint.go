package sonm

import (
	"fmt"
	"math/big"
)

// NewBigInt constructs a new value using specified big.Int.
func NewBigInt(v *big.Int) *BigInt {
	return &BigInt{
		Neg: v.Sign() < 0,
		Abs: v.Bytes(),
	}
}

// NewBigIntFromString tries to construct a new value from the specified string.
func NewBigIntFromString(s string) (*BigInt, error) {
	v := new(big.Int)
	v, ok := v.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert %s to big.Int", s)
	}
	return NewBigInt(v), nil
}

// Unwrap returns the current price as *big.Int.
func (m *BigInt) Unwrap() *big.Int {
	v := new(big.Int)
	v.SetBytes(m.Abs)
	if m.Neg {
		v.Neg(v)
	}
	return v
}
