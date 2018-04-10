package cpu

import (
	"errors"

	"github.com/shirou/gopsutil/cpu"
	"github.com/sonm-io/core/proto"
)

func GetCPUDevice() (*sonm.CPUDevice, error) {
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	if len(info) == 0 {
		return nil, errors.New("no CPU detected")
	}

	// We've picked up a name of the first CPU because assuming
	// that multi-CPU board will have similar CPUs into the sockets.
	dev := &sonm.CPUDevice{
		ModelName: info[0].ModelName,
		Sockets:   uint32(len(info)),
	}

	for _, c := range info {
		dev.Cores += uint32(c.Cores)
	}

	return dev, nil
}
