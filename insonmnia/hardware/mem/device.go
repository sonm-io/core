package mem

type Device interface {
	// PhysicalId returns device's physical id on the motherboard.
	PhysicalId() int
	// DeviceName returns device name for example "DIMM DDR3 1333 MHz (0.8 ns)".
	DeviceName() string
	// Vendor returns vendor name.
	VendorName() string
	// MemorySize returns memory size in bytes.
	MemorySize() uint64
	// Width returns device's bus width in bits.
	Width() uint64
	// ClockFrequency returns a memory device clock frequency in Hz.
	ClockFrequency() uint64
}

// On other platforms get just MemorySize
