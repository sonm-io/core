package sonm

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/sonm-io/core/util/multierror"
)

const (
	MinNumBenchmarks = 12
	MinDealDuration  = time.Minute * 10
	MaxTagLength     = 32
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
	if len(m.GetTag()) > MaxTagLength {
		return errors.New("tag value is too long")
	}

	return nil
}

func (m *Benchmarks) GetNValues(targetSize uint64) []uint64 {
	out := make([]uint64, targetSize)
	if m != nil {
		copy(out, m.Values)
	}

	return out
}

func (m *Benchmarks) Validate() error {
	if len(m.Values) < MinNumBenchmarks {
		return fmt.Errorf("expected at least %d benchmarks, got %d", MinNumBenchmarks, len(m.Values))
	}
	return nil
}

func (m *Benchmarks) Get(idx int) uint64 {
	if m == nil {
		return 0
	}
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

func (m *Benchmarks) CPUCryptonight() uint64 {
	return m.Get(12)
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

func (m *Order) PricePerHour() string {
	secondsInHour := uint64(3600)
	return formatPriceString(m.GetPrice(), secondsInHour)
}

func (m *Deal) TotalPrice() string {
	return formatPriceString(m.GetPrice(), m.GetDuration())
}

func (m *Deal) PricePerHour() string {
	secondsInHour := uint64(3600)
	return formatPriceString(m.GetPrice(), secondsInHour)
}

func formatPriceString(price *BigInt, duration uint64) string {
	d := big.NewInt(int64(duration))
	p := big.NewInt(0).Mul(price.Unwrap(), d)
	return NewBigInt(p).ToPriceString()
}

func CombinedError(status *ErrorByID) error {
	merr := multierror.NewMultiError()
	for _, err := range status.GetResponse() {
		if len(err.Error) != 0 {
			merr = multierror.Append(merr, fmt.Errorf("failed to process id %s: %v", err.GetId().Unwrap().String(), err.GetError()))
		}
	}
	return merr.ErrorOrNil()
}

func NewTSErrorByID() *TSErrorByID {
	return &TSErrorByID{
		inner: &ErrorByID{
			Response: []*ErrorByID_Item{},
		},
	}
}

type TSErrorByID struct {
	mu    sync.Mutex
	inner *ErrorByID
}

func (m *TSErrorByID) Append(id *BigInt, err error) {
	strErr := ""
	if err != nil {
		strErr = err.Error()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inner.Response = append(m.inner.Response, &ErrorByID_Item{
		Id:    id,
		Error: strErr,
	})
}

func (m *TSErrorByID) Unwrap() *ErrorByID {
	return m.inner
}

func NewTSErrorByStringID() *TSErrorByStringID {
	return &TSErrorByStringID{
		inner: &ErrorByStringID{
			Response: []*ErrorByStringID_Item{},
		},
	}
}

type TSErrorByStringID struct {
	mu    sync.Mutex
	inner *ErrorByStringID
}

func (m *TSErrorByStringID) Append(id string, err error) {
	strErr := ""
	if err != nil {
		strErr = err.Error()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inner.Response = append(m.inner.Response, &ErrorByStringID_Item{
		ID:    id,
		Error: strErr,
	})
}

func (m *TSErrorByStringID) Unwrap() *ErrorByStringID {
	return m.inner
}
