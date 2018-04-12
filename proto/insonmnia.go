package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/util/datasize"
)

var bigEther = big.NewFloat(params.Ether).SetPrec(256)

var priceSuffixes = map[string]big.Float{
	"SNM/s": *bigEther,
	"SNM/h": *big.NewFloat(0).SetPrec(256).Quo(bigEther, big.NewFloat(3600)),
}

var possiblePriceSuffixxes = func() string {
	keys := make([]string, 0, len(priceSuffixes))
	for k := range priceSuffixes {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}()

func (m *Duration) Unwrap() time.Duration {
	return time.Nanosecond * time.Duration(m.GetNanoseconds())
}

func (m *Duration) MarshalYAML() (interface{}, error) {
	return m.Unwrap().String(), nil
}

func (m *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}

	m.Nanoseconds = d.Nanoseconds()
	return nil
}

func (m *EthAddress) Unwrap() common.Address {
	return common.BytesToAddress(m.Address)
}

func (m *EthAddress) MarshalYAML() (interface{}, error) {
	return m.Unwrap().Hex(), nil
}

func (m *EthAddress) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	if !common.IsHexAddress(v) {
		return errors.New("invalid ethereum address format")
	}

	m.Address = common.HexToAddress(v).Bytes()
	return nil
}

func (m *DataSize) Unwrap() datasize.ByteSize {
	return datasize.NewByteSize(m.Bytes)
}

func (m *DataSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	var byteSize datasize.ByteSize
	if err := byteSize.UnmarshalText([]byte(v)); err != nil {
		return err
	}

	m.Bytes = byteSize.Bytes()
	return nil
}

func (m *DataSize) MarshalYAML() (interface{}, error) {
	text, err := m.Unwrap().MarshalText()
	return string(text), err
}

func (m *DataSizeRate) Unwrap() datasize.BitRate {
	return datasize.NewBitRate(m.BitsPerSecond)
}

func (m *DataSizeRate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	var bitRate datasize.BitRate
	if err := bitRate.UnmarshalText([]byte(v)); err != nil {
		return err
	}

	m.BitsPerSecond = bitRate.Bits()
	return nil
}

func (m *DataSizeRate) MarshalYAML() (interface{}, error) {
	text, err := m.Unwrap().MarshalText()
	return string(text), err
}

func (m *Price) MarshalYAML() (interface{}, error) {
	v := big.NewFloat(0).SetPrec(256).SetInt(m.PerSecond.Unwrap())
	div := big.NewFloat(params.Ether).SetPrec(256)
	div.Quo(div, big.NewFloat(3600.))

	r := big.NewFloat(0).Quo(v, div)
	return r.Text('g', 10) + " SNM/h", nil
}

func (m *Price) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	if err := m.LoadFromString(v); err != nil {
		return err
	}

	return nil
}

func (m *Price) LoadFromString(v string) error {
	parts := strings.FieldsFunc(v, func(c rune) bool {
		return c == ' '
	})

	if len(parts) != 2 {
		return fmt.Errorf("could not load price - %s can not be split to numeric and dimension parts", v)
	}

	dimensionMultiplier, ok := priceSuffixes[parts[1]]
	if !ok {
		return fmt.Errorf("could not load price - unknown dimension %s, possible values are - %s", parts[1], possiblePriceSuffixxes)
	}

	fractPrice, _, err := big.ParseFloat(parts[0], 10, 256, big.ToNearestEven)
	if err != nil {
		return fmt.Errorf("could not load price - failed to parse numeric part %s to big float: %s", parts[0], err)
	}
	price, _ := fractPrice.Mul(fractPrice, &dimensionMultiplier).Int(nil)
	m.PerSecond = NewBigInt(price)

	return nil
}
