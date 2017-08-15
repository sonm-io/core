package resource

import (
	"errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	"sync"
)

type Resources struct {
	numCPUs int
	memory  int64
	// TODO: It's unclear how to calculate GPU usage.
}

func NewResources(numCPUs int, memory int64) Resources {
	return Resources{
		numCPUs: numCPUs,
		memory:  memory,
	}
}

type Pool struct {
	OS    *hardware.Hardware
	mu    sync.Mutex
	usage Resources
}

func NewPool(hardware *hardware.Hardware) *Pool {
	return &Pool{
		OS:    hardware,
		usage: Resources{},
	}
}

// TODO: May be return some kind of Retainer to be able to auto-retain?
func (p *Pool) Consume(usage *Resources) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	free := NewResources(p.OS.LogicalCPUCount()-p.usage.numCPUs, int64(p.OS.Memory.Total)-p.usage.memory)

	if usage.numCPUs > free.numCPUs {
		return errors.New("not enough CPU available")
	}

	if usage.memory > free.memory {
		return errors.New("not enough memory available")
	}

	p.usage.numCPUs += usage.numCPUs
	p.usage.memory += usage.memory

	return nil
}

func (p *Pool) Retain(usage *Resources) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.usage.numCPUs -= usage.numCPUs
	p.usage.memory -= usage.memory
}
