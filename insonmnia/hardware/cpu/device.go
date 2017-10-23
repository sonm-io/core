package cpu

import (
	"github.com/cnf/structhash"
	"github.com/shirou/gopsutil/cpu"
)


// Device describes a CPU device.
type Device cpu.InfoStat

func GetCPUDevices() ([]Device, error) {
	info, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	devices := []Device{}
	for _, device := range info {
		dev := Device(device)
		devices = append(devices, dev)
	}
	return devices, nil
}

func (d *Device) Hash() []byte {
	return structhash.Md5(d, 1)
}
