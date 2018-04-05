package cpu

import (
	"errors"

	"github.com/cnf/structhash"
	"github.com/shirou/gopsutil/cpu"
)

// Device describes a CPU device.
type Device struct {
	Name    string `json:"model"`
	Cores   uint   `json:"cores"`
	Sockets uint   `json:"sockets"`
}

func (d *Device) Hash() []byte {
	return structhash.Md5(d, 1)
}

func GetCPUDevice() (*Device, error) {
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	if len(info) == 0 {
		// o_O
		return nil, errors.New("no CPU detected")
	}

	dev := &Device{Name: info[0].ModelName}
	for _, c := range info {
		dev.Cores += uint(c.Cores)
		dev.Sockets += 1
	}

	return dev, nil
}
