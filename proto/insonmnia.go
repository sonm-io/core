package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var rateSuffixes = map[string]uint64{
	"bit/s":  1,
	"kbit/s": 1e3,
	"kb/s":   1e3,
	"Mb/s":   1e6,
	"Mbit/s": 1e6,
	"Gb/s":   1e9,
	"Gbit/s": 1e9,
	"Tb/s":   1e12,
	"Tbit/s": 1e12,

	"Kibit/s": 1 << 10,
	"Mibit/s": 1 << 20,
	"Gibit/s": 1 << 30,
	"Tibit/s": 1 << 40,
}

var possibleRateSiffixesStr = func() string {
	keys := make([]string, 0, len(rateSuffixes))
	for k := range rateSuffixes {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}()

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

func (m *DataSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	var byteSize datasize.ByteSize
	if err := byteSize.UnmarshalText([]byte(strings.ToLower(v))); err != nil {
		return err
	}

	m.Bytes = byteSize.Bytes()
	return nil
}

func (m *DataSizeRate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	parts := strings.FieldsFunc(v, func(c rune) bool {
		return c == ' '
	})
	if len(parts) != 2 {
		return fmt.Errorf("could not parse DataSizeRate - \"%s\" can not be split to 2 parts", v)
	}

	value, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("could not parse DataSizeRate numeric part -  %s", err)
	}
	multiplier, ok := rateSuffixes[parts[1]]
	if !ok {

		return fmt.Errorf("could not parse DataSizeRate - unknown data rate suffix \"%s\". Possible values are - %s", v, possibleRateSiffixesStr)
	}

	m.BitsPerSecond = value * multiplier
	return nil
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
