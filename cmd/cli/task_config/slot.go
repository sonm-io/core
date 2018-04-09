package task_config

/*
	Note: whole `Slot` structure and it's wrapper seems deprecated;
*/

import (
	"strings"
	"time"

	ds "github.com/c2h5oh/datasize"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
)

type DurationConfig struct {
	Since string `yaml:"since" required:"true"`
	Until string `yaml:"until" required:"true"`
}

type ResourcesConfig struct {
	Cpu        uint64             `yaml:"cpu_cores" required:"true"`
	Ram        string             `yaml:"ram_bytes" required:"true"`
	Gpu        string             `yaml:"gpu_count" required:"true"`
	Storage    string             `yaml:"storage" required:"true"`
	Network    NetworkConfig      `yaml:"network" required:"true"`
	Properties map[string]float64 `yaml:"properties" required:"true"`
}

type NetworkConfig struct {
	In   string `yaml:"in" required:"true"`
	Out  string `yaml:"out" required:"true"`
	Type string `yaml:"type" required:"true"`
}

type SlotConfig struct {
	Duration  string          `yaml:"duration" required:"true"`
	Resources ResourcesConfig `yaml:"resources" required:"true"`
}

func (c *SlotConfig) IntoSlot() (*structs.Slot, error) {
	networkType, err := structs.ParseNetworkType(c.Resources.Network.Type)
	if err != nil {
		return nil, err
	}

	gpuCount, err := structs.ParseGPUCount(c.Resources.Gpu)
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(c.Duration)
	if err != nil {
		return nil, err
	}

	var ram, storage, netIn, netOut ds.ByteSize
	err = ram.UnmarshalText([]byte(strings.ToLower(c.Resources.Ram)))
	if err != nil {
		return nil, err
	}

	err = storage.UnmarshalText([]byte(strings.ToLower(c.Resources.Storage)))
	if err != nil {
		return nil, err
	}

	err = netIn.UnmarshalText([]byte(strings.ToLower(c.Resources.Network.In)))
	if err != nil {
		return nil, err
	}

	err = netOut.UnmarshalText([]byte(strings.ToLower(c.Resources.Network.Out)))
	if err != nil {
		return nil, err
	}

	return structs.NewSlot(&sonm.Slot{
		Duration: uint64(duration.Round(time.Second).Seconds()),
		Resources: &sonm.Resources{
			CpuCores:      c.Resources.Cpu,
			RamBytes:      ram.Bytes(),
			GpuCount:      gpuCount,
			Storage:       storage.Bytes(),
			NetTrafficIn:  netIn.Bytes(),
			NetTrafficOut: netOut.Bytes(),
			NetworkType:   networkType,
			Properties:    c.Resources.Properties,
		},
	})
}
