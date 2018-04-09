package task_config

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	ds "github.com/c2h5oh/datasize"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/proto"
)

const (
	minRamSize  uint64 = 4 * 1024 * 1024
	minCPUCount        = 0.01
)

type DataSize struct {
	size uint64
}

func (d *DataSize) Bytes() uint64 {
	return d.size
}

func (d *DataSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	var bs ds.ByteSize
	if err := bs.UnmarshalText([]byte(strings.ToLower(v))); err != nil {
		return err
	}

	if bs.Bytes() < minRamSize {
		return errors.New("RAM size is too low")
	}

	d.size = bs.Bytes()
	return nil
}

type CpuCount struct {
	count uint64
}

func (cc *CpuCount) Count() uint64 {
	return cc.count
}

func (cc *CpuCount) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	rawCount, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return err
	}

	if rawCount < minCPUCount {
		return errors.New("CPU count is too low")
	}

	cc.count = uint64(math.Floor(rawCount * 100))
	return nil
}

type ethStr struct {
	addr common.Address
}

func (es *ethStr) Address() common.Address {
	return es.addr
}

func (es *ethStr) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	if !common.IsHexAddress(v) {
		return errors.New("invalid ethereum address format")
	}

	es.addr = common.HexToAddress(v)
	return nil
}

type ethPrice struct {
	price *big.Int
}

func (ep *ethPrice) Price() *big.Int {
	return ep.price
}

func (ep *ethPrice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v string
	if err := unmarshal(&v); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("failed to convert \"%s\" to float64", v)
	}

	if f < 0 {
		return errors.New("price cannot be negative")
	}

	bf := big.NewFloat(0).Mul(big.NewFloat(f), big.NewFloat(params.Ether))
	ep.price, _ = bf.Int(nil)
	return nil
}

type AskPlanYAML struct {
	Duration     time.Duration `yaml:"duration" required:"true"`
	PricePerHour *ethPrice     `yaml:"price_per_hour" required:"true"`
	Blacklist    *ethStr       `yaml:"blacklist"`

	Resources struct {
		CPU struct {
			Cores CpuCount `yaml:"cores" required:"true"`
		} `yaml:"cpu"`

		RAM struct {
			Size DataSize `yaml:"size" required:"true"`
		} `yaml:"ram"`

		Storage struct {
			Size DataSize `yaml:"size"`
		} `yaml:"storage"`

		GPU struct {
			Devices []uint64 `yaml:"devices"`
		} `yaml:"gpu"`

		Net struct {
			ThroughputIn  uint64 `yaml:"throughput_in"`
			ThroughputOut uint64 `yaml:"throughput_out"`
			Overlay       bool   `yaml:"overlay"`
			Outbound      bool   `yaml:"outbound"`
			Incoming      bool   `yaml:"incoming"`
		} `yaml:"net"`
	}
}

func (ask *AskPlanYAML) pricePerSecond() *big.Int {
	sec := big.NewInt(3600)
	return big.NewInt(0).Quo(ask.PricePerHour.Price(), sec)
}

func (ask *AskPlanYAML) Unwrap() (*sonm.AskPlan, error) {
	plan := &sonm.AskPlan{
		Duration:       uint64(ask.Duration.Truncate(time.Second).Seconds()),
		PricePerSecond: sonm.NewBigInt(ask.pricePerSecond()),
		BlacklistAddr:  ask.Blacklist.Address().Hex(),
		Resources: &sonm.AskPlanResources{
			Cpu:     &sonm.AskPlanCPU{Cores: ask.Resources.CPU.Cores.Count()},
			Ram:     &sonm.AskPlanRAM{Size: ask.Resources.RAM.Size.Bytes()},
			Storage: &sonm.AskPlanStorage{Size: ask.Resources.Storage.Size.Bytes()},
			Gpu:     &sonm.AskPlanGPU{Devices: ask.Resources.GPU.Devices},
			Network: &sonm.AskPlanNetwork{
				ThroughputIn:  ask.Resources.Net.ThroughputIn,
				ThroughputOut: ask.Resources.Net.ThroughputOut,
				Overlay:       ask.Resources.Net.Overlay,
				Outbound:      ask.Resources.Net.Outbound,
				Incoming:      ask.Resources.Net.Incoming,
			},
		},
	}

	return plan, nil
}

func LoadAskPlan(p string) (*AskPlanYAML, error) {
	ask := &AskPlanYAML{}
	if err := configor.Load(ask, p); err != nil {
		return nil, err
	}

	return ask, nil
}
