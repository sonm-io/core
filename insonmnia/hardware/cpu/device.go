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
		// o_O
		return nil, errors.New("no CPU detected")
	}

	dev := &sonm.CPUDevice{ModelName: info[0].ModelName}
	for _, c := range info {
		dev.Cores += uint32(c.Cores)
		dev.Sockets += 1
	}

	return dev, nil
}
