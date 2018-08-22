package optimus

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/sonm-io/core/proto"
)

type PriceThreshold interface {
	// Exceeds returns "true" if the "price" exceeds "otherPrice" according to
	// the implementation's threshold policy.
	Exceeds(price, otherPrice *big.Int) bool
}

type RelativePriceThreshold struct {
	*big.Int
}

func NewRelativePriceThreshold(threshold float64) (*RelativePriceThreshold, error) {
	if threshold <= 0.0 {
		return nil, errors.New("price threshold must be a positive number")
	}

	m := &RelativePriceThreshold{
		Int: big.NewInt(int64(threshold * 1000)),
	}

	return m, nil
}

func ParseRelativePriceThreshold(threshold string) (*RelativePriceThreshold, error) {
	// Drop whitespaces if any.
	value := strings.Replace(threshold, " ", "", -1)
	idx := strings.IndexByte(value, '%')
	if idx == -1 {
		return nil, errors.New("`%` sign is required")
	}
	if idx != len(value)-1 {
		return nil, errors.New("trailing characters after `%` sign")
	}
	percent, err := strconv.ParseFloat(value[:idx], 64)
	if err != nil {
		return nil, errors.New("threshold must be a number")
	}

	return NewRelativePriceThreshold(percent)
}

func (m *RelativePriceThreshold) Exceeds(price, otherPrice *big.Int) bool {
	v := new(big.Int).Sub(new(big.Int).Div(new(big.Int).Mul(price, big.NewInt(100000)), otherPrice), big.NewInt(100000))
	return v.Cmp(m.Int) >= 0
}

type AbsolutePriceThreshold struct {
	*sonm.Price
}

func NewAbsolutePriceThreshold(threshold *sonm.Price) (*AbsolutePriceThreshold, error) {
	if threshold.GetPerSecond().Unwrap().Sign() <= 0 {
		return nil, errors.New("price threshold must be a positive number")
	}

	m := &AbsolutePriceThreshold{
		Price: threshold,
	}

	return m, nil
}

func ParseAbsolutePriceThreshold(threshold string) (*AbsolutePriceThreshold, error) {
	v := &sonm.Price{}
	if err := v.UnmarshalText([]byte(threshold)); err != nil {
		return nil, fmt.Errorf("failed to parse absolute price threshold: %v", err)
	}

	return NewAbsolutePriceThreshold(v)
}

func (m *AbsolutePriceThreshold) Exceeds(price, otherPrice *big.Int) bool {
	diff := new(big.Int).Sub(price, otherPrice)
	return diff.Cmp(m.PerSecond.Unwrap()) >= 0
}

type priceThreshold struct {
	PriceThreshold
}

func (m *priceThreshold) Exceeds(price, otherPrice *big.Int) bool {
	return m.PriceThreshold.Exceeds(price, otherPrice)
}

func (m *priceThreshold) UnmarshalText(text []byte) error {
	if threshold, err := ParseAbsolutePriceThreshold(string(text)); err == nil {
		m.PriceThreshold = threshold
	}
	if threshold, err := ParseRelativePriceThreshold(string(text)); err == nil {
		m.PriceThreshold = threshold
	}

	return errors.New("invalid price threshold format: must be either `N USD/s`, `N USD/h` or `N%`, where N - number")
}
