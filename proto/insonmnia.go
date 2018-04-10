package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	ds "github.com/c2h5oh/datasize"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var snmPriceRe = regexp.MustCompile(`(\S+)(\s*)(snm)(\S+)`)

func (m *Duration) Unwrap() time.Duration {
	return time.Second * time.Duration(m.GetSeconds())
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

	m.Seconds = int64(d.Truncate(time.Second).Seconds())
	return nil
}

func (m *EthAddress) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	if !common.IsHexAddress(v) {
		return errors.New("invalid ethereum address format")
	}

	m.Address = common.HexToAddress(v).Hex()
	return nil
}

func (m *DataSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	var bs ds.ByteSize
	if err := bs.UnmarshalText([]byte(strings.ToLower(v))); err != nil {
		return err
	}

	m.Size = bs.Bytes()
	return nil
}

func (m *DataSizeRate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	v = trimTimeRate(v)
	var bs ds.ByteSize
	if err := bs.UnmarshalText([]byte(v)); err != nil {
		return err
	}

	m.Size = bs.Bytes()
	return nil
}

func trimTimeRate(v string) string {
	v = strings.ToLower(v)
	v = strings.Trim(v, `\s`)
	v = strings.Trim(v, `/s`)
	return v
}

func (m *Price) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	value, timeDim, err := extractPricePerTimeValues(v)
	if err != nil {
		return err
	}

	price, err := convertToPrice(value, timeDim)
	if err != nil {
		return err
	}

	m.PerSecond = price
	return nil
}

func extractPricePerTimeValues(v string) (string, string, error) {
	v = strings.ToLower(v)
	matches := snmPriceRe.FindStringSubmatch(v)
	if len(matches) != 5 {
		return "", "", errors.New("invalid price format")
	}

	num := matches[1]
	timeDim := strings.TrimFunc(matches[4], func(r rune) bool {
		return r == '/' || r == '\\'
	})

	return num, timeDim, nil
}

func convertToPrice(value, timeDim string) (*BigInt, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert \"%s\" to float64", value)
	}

	if f < 0 {
		return nil, errors.New("price cannot be negative")
	}

	bigPrice, _ := big.NewFloat(0).Mul(big.NewFloat(f), big.NewFloat(params.Ether)).Int(nil)
	if err != nil {
		return nil, err
	}

	var div *big.Int
	switch timeDim {
	case "h":
		div = big.NewInt(3600)
	case "s":
		div = big.NewInt(1)
	default:
		return nil, errors.New("invalid time dimension")
	}

	p := NewBigInt(big.NewInt(0).Quo(bigPrice, div))
	return p, nil
}
