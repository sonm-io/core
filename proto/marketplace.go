package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const (
	MinNumBenchmarks = 12
	MinDealDuration  = time.Minute * 10
)

func (m *IdentityLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	level, ok := IdentityLevel_value[strings.ToUpper(v)]
	if !ok {
		return errors.New("unknown identity level")
	}

	*m = IdentityLevel(level)
	return nil
}

func (m *BidOrder) Validate() error {
	if len(m.GetTag()) > 32 {
		return errors.New("tag value is too long")
	}

	return nil
}

func (m *Benchmarks) Validate() error {
	if len(m.Values) < MinNumBenchmarks {
		return fmt.Errorf("expected at least %d benchmarks, got %d", MinNumBenchmarks, len(m.Values))
	}
	return nil
}

func (m *Benchmarks) Get(idx int) uint64 {
	if len(m.Values) <= idx {
		return 0
	}
	return m.Values[idx]
}

func (m *Benchmarks) CPUSysbenchMulti() uint64 {
	return m.Get(0)
}

func (m *Benchmarks) CPUSysbenchOne() uint64 {
	return m.Get(1)
}

func (m *Benchmarks) CPUCores() uint64 {
	return m.Get(2)
}

func (m *Benchmarks) RAMSize() uint64 {
	return m.Get(3)
}

func (m *Benchmarks) StorageSize() uint64 {
	return m.Get(4)
}

func (m *Benchmarks) NetTrafficIn() uint64 {
	return m.Get(5)
}

func (m *Benchmarks) NetTrafficOut() uint64 {
	return m.Get(6)
}

func (m *Benchmarks) GPUCount() uint64 {
	return m.Get(7)
}

func (m *Benchmarks) GPUMem() uint64 {
	return m.Get(8)
}

func (m *Benchmarks) GPUEthHashrate() uint64 {
	return m.Get(9)
}

func (m *Benchmarks) GPUCashHashrate() uint64 {
	return m.Get(10)
}

func (m *Benchmarks) GPURedshift() uint64 {
	return m.Get(11)
}

func (m *Deal) GetTypeName() string {
	if m.IsSpot() {
		return "Spot"
	} else {
		return "Forward"
	}
}

func (m *Deal) IsSpot() bool {
	return m.GetDuration() == 0
}

func (m *Order) TotalPrice() string {
	return formatPriceString(m.GetPrice(), m.GetDuration())
}

func (m *Deal) TotalPrice() string {
	return formatPriceString(m.GetPrice(), m.GetDuration())
}

func formatPriceString(price *BigInt, duration uint64) string {
	d := big.NewInt(int64(duration))
	p := big.NewInt(0).Mul(price.Unwrap(), d)
	return NewBigInt(p).ToPriceString()
}
