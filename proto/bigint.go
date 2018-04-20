package sonm

import (
	"encoding/json"
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

func (m *BigInt) IsZero() bool {
	return m.Unwrap().BitLen() == 0
}

func (m BigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Unwrap().String())
}

func (m *BigInt) UnmarshalJSON(data []byte) error {
	var unmarshaled string
	err := json.Unmarshal(data, &unmarshaled)
	if err != nil {
		return err
	}

	v, err := NewBigIntFromString(unmarshaled)
	if err != nil {
		return err
	}

	m.Abs = v.Abs
	m.Neg = v.Neg

	return nil
}

func (m *BigInt) ToPriceString() string {
	v := big.NewFloat(0).SetInt(m.Unwrap())
	div := big.NewFloat(params.Ether)

	r := big.NewFloat(0).Quo(v, div)
	return r.Text('f', -18)
}

func (m *BigInt) PaddedString() string {
	return util.BigIntToPaddedString(m.Unwrap())
}
