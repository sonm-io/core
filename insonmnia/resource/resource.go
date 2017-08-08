package resource

import "github.com/cloudfoundry/gosigar"

// OS represents available resources on the OS, where a Miner is running.
type OS struct {
	CPU sigar.CpuList
	Mem sigar.Mem
}

// Collector represents a resource collector.
type Collector interface {
	// OS returns total resources available on the OS.
	OS() (*OS, error)
}

type collector struct {
	inner sigar.ConcreteSigar
}

func (c *collector) OS() (*OS, error) {
	cpu := sigar.CpuList{}
	if err := cpu.Get(); err != nil {
		return nil, err
	}

	mem := sigar.Mem{}
	if err := mem.Get(); err != nil {
		return nil, err
	}

	os := &OS{
		CPU: cpu,
		Mem: mem,
	}

	return os, nil
}

// New returns a new OS specific resource collector.
func New() Collector {
	return &collector{
		inner: sigar.ConcreteSigar{},
	}
}
