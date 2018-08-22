package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/datasize"
)

var bigEther = big.NewFloat(params.Ether).SetPrec(256)

var priceSuffixes = map[string]big.Float{
	"USD/s": *bigEther,
	"USD/h": *big.NewFloat(0).SetPrec(256).Quo(bigEther, big.NewFloat(3600)),
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
	if m == nil {
		return common.Address{}
	}
	return common.BytesToAddress(m.Address)
}

func (m EthAddress) MarshalText() ([]byte, error) {
	return []byte(m.Unwrap().Hex()), nil
}

func (m *EthAddress) UnmarshalText(text []byte) error {
	v := string(text)
	if !common.IsHexAddress(v) {
		return fmt.Errorf("invalid ethereum address format \"%s\"", v)
	}

	m.Address = common.HexToAddress(v).Bytes()
	return nil
}

func (m *EthAddress) IsZero() bool {
	if m == nil {
		return true
	}

	return m.Unwrap().Big().BitLen() == 0
}

func NewEthAddress(addr common.Address) *EthAddress {
	return &EthAddress{Address: addr.Bytes()}
}

func NewEthAddressFromHex(hexAddr string) (*EthAddress, error) {
	addr, err := util.HexToAddress(hexAddr)
	if err != nil {
		return nil, err
	}

	return &EthAddress{Address: addr.Bytes()}, nil
}

func (m *DataSize) Unwrap() datasize.ByteSize {
	if m == nil {
		return datasize.ByteSize{}
	}
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
	if m == nil {
		return datasize.BitRate{}
	}
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
	return r.Text('g', 10) + " USD/h", nil
}

func (m *Price) UnmarshalText(text []byte) error {
	if err := m.LoadFromString(string(text)); err != nil {
		return err
	}

	return nil
}

func (m *Price) LoadFromString(v string) error {
	delimAt := strings.IndexFunc(v, func(c rune) bool {
		return unicode.IsLetter(c)
	})

	if delimAt < 0 {
		return fmt.Errorf("could not load price - %s can not be split to numeric and dimension parts", v)
	}

	parts := []string{
		strings.TrimSpace(v[:delimAt]),
		strings.TrimSpace(v[delimAt:]),
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

func (m *StartTaskRequest) Validate() error {
	if m.GetDealID().IsZero() {
		return errors.New("non-zero deal id is required for start task request")
	}
	return m.GetSpec().Validate()
}

func (m *TaskSpec) Validate() error {
	return m.GetContainer().Validate()
}
