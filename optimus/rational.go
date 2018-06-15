package optimus

import (
	"errors"
	"math/big"
)

type Rational struct {
	v *big.Rat
}

func NewRational(a, b uint64) Rational {
	return Rational{
		v: big.NewRat(int64(a), int64(b)),
	}
}

func (m Rational) Mul(v uint64) Rational {
	return Rational{
		v: big.NewRat(1, 1).Mul(m.v, big.NewRat(int64(v), 1)),
	}
}

func (m Rational) Div(v uint64) Rational {
	return Rational{
		v: big.NewRat(1, 1).Quo(m.v, big.NewRat(int64(v), 1)),
	}
}

func (m Rational) Float64() float64 {
	r, _ := m.v.Float64()
	return r
}

func (m Rational) Cmp(other Rational) int {
	return m.v.Cmp(other.v)
}

func Max(v []Rational) (Rational, error) {
	if len(v) == 0 {
		return Rational{}, errors.New("empty input")
	}

	max := v[0]

	for i := 1; i < len(v); i++ {
		if v[i].Cmp(max) > 0 {
			max = v[i]
		}
	}

	return max, nil
}
