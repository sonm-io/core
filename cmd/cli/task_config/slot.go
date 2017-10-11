package task_config

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
)

type DurationConfig struct {
	Since string `yaml:"since" required:"true"`
	Until string `yaml:"until" required:"true"`
}

type RatingConfig struct {
	Buyer    int64 `yaml:"buyer" required:"true"`
	Supplier int64 `yaml:"supplier" required:"true"`
}

type ResourcesConfig struct {
	Cpu        uint64            `yaml:"cpu_cores" required:"true"`
	Ram        uint64            `yaml:"ram_bytes" required:"true"`
	Gpu        uint64            `yaml:"gpu_count" required:"true"`
	Storage    uint64            `yaml:"storage" required:"true"`
	Network    NetworkConfig     `yaml:"network" required:"true"`
	Properties map[string]string `yaml:"properties" required:"true"`
}

type NetworkConfig struct {
	In   uint64 `yaml:"in" required:"true"`
	Out  uint64 `yaml:"out" required:"true"`
	Type string `yaml:"type" required:"true"`
}

type SlotConfig struct {
	Duration  DurationConfig  `yaml:"duration" required:"true"`
	Rating    RatingConfig    `yaml:"rating" required:"true"`
	Resources ResourcesConfig `yaml:"resources" required:"true"`
}

func ParseNetworkType(ty string) (sonm.NetworkType, error) {
	switch ty {
	case "NO_NETWORK":
		return sonm.NetworkType_NO_NETWORK, nil
	case "INCOMING":
		return sonm.NetworkType_INCOMING, nil
	case "OUTBOUND":
		return sonm.NetworkType_OUTBOUND, nil
	default:
		return sonm.NetworkType_NO_NETWORK, errors.New("unknown network type")
	}
}

func (c *SlotConfig) IntoSlot() (*structs.Slot, error) {
	since, err := time.Parse(time.RFC3339, c.Duration.Since)
	if err != nil {
		return nil, err
	}

	until, err := time.Parse(time.RFC3339, c.Duration.Until)
	if err != nil {
		return nil, err
	}

	networkType, err := ParseNetworkType(c.Resources.Network.Type)
	if err != nil {
		return nil, err
	}

	return structs.NewSlot(&sonm.Slot{
		StartTime: &sonm.Timestamp{
			Seconds: int64(since.Unix()),
		},
		EndTime: &sonm.Timestamp{
			Seconds: int64(until.Unix()),
		},
		BuyerRating:    c.Rating.Buyer,
		SupplierRating: c.Rating.Supplier,
		Resources: &sonm.Resources{
			CpuCores:      c.Resources.Cpu,
			RamBytes:      c.Resources.Ram,
			GpuCount:      c.Resources.Gpu,
			Storage:       c.Resources.Storage,
			NetTrafficIn:  c.Resources.Network.In,
			NetTrafficOut: c.Resources.Network.Out,
			NetworkType:   networkType,
			Props:         c.Resources.Properties,
		},
	})
}
