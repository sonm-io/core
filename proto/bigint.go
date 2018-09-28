package sonm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/util"
)

// NewBigInt constructs a new value using specified big.Int.
func NewBigInt(v *big.Int) *BigInt {
	if v == nil {
		v = big.NewInt(0)
	}

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
	if m == nil {
		return v
	}
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

func (m BigInt) MarshalText() (text []byte, err error) {
	return []byte(m.Unwrap().String()), nil
}

func (m *BigInt) UnmarshalText(text []byte) error {
	unmarshalled, err := NewBigIntFromString(string(text))
	if err != nil {
		return err
	}
	m.Abs = unmarshalled.Abs
	m.Neg = unmarshalled.Neg
	return nil
}

func (m *BigInt) ToPriceString() string {
	v := big.NewFloat(0).SetInt(m.Unwrap())
	div := big.NewFloat(params.Ether)

	r := big.NewFloat(0).Quo(v, div)
	return r.Text('f', -18)
}

func (m *BigInt) PricePerHour() string {
	v := big.NewInt(0).Mul(big.NewInt(3600), m.Unwrap())
	return NewBigInt(v).ToPriceString()
}

func (m *BigInt) PaddedString() string {
	return util.BigIntToPaddedString(m.Unwrap())
}

func (m *BigInt) IsZero() bool {
	if m == nil {
		return true
	}

	return m.Unwrap().BitLen() == 0
}
