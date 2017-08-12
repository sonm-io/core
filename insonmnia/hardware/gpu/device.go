package gpu

type Device struct {
	// Model name.
	Name string
	// Vendor name.
	Vendor string
	// Extension flags.
	Flags []string
	// Maximum configured clock frequency of the device in MHz.
	MaxClockFrequency int
	// The default compute device address space size specified as an unsigned integer value in bits.
	AddressBits int
	// Size of global memory cache line in bytes.
	CacheLineSize int
}

func GetGPUDevices() ([]*Device, error) {
	devices, err := GetGPUDevicesUsingOpenCL()
	if err != nil {
		return nil, err
	}

	return devices, nil
}
