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

func NewBigIntFromInt(v int64) *BigInt {
	return NewBigInt(big.NewInt(v))
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

// Cmp compares this value with the other one, returning:
//  -1 if x <  y
//   0 if x == y
//  +1 if x >  y
func (m *BigInt) Cmp(other *BigInt) int {
	return m.Unwrap().Cmp(other.Unwrap())
}
