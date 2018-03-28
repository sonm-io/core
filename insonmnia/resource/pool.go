package resource

import (
	"errors"
	"sync"

	"github.com/sonm-io/core/insonmnia/hardware"
)

var (
	ErrNotEnoughCPU    = errors.New("not enough CPU available")
	ErrNotEnoughMemory = errors.New("not enough memory available")
	ErrNotEnoughGPU    = errors.New("not enough GPU available")
)

type Resources struct {
	NumCPUs int
	Memory  int64
	// NumGPUs shows the number of GPUs required for a task.
	// A value of -1 means that a task consumes all of available GPU devices.
	NumGPUs int
}

func NewResources(numCPUs int, memory int64, numGPUs int) Resources {
	return Resources{
		NumCPUs: numCPUs,
		Memory:  memory,
		NumGPUs: numGPUs,
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

func (p *Pool) GetUsage() Resources {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.usage
}

// Consume tries to consume the specified resource usage from the pool.
//
// Does nothing on error.
func (p *Pool) Consume(usage *Resources) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.consume(usage)
}

func (p *Pool) consume(usage *Resources) error {
	if err := p.pollConsume(usage); err != nil {
		return err
	}

	p.usage.NumCPUs += usage.NumCPUs
	p.usage.Memory += usage.Memory
	if usage.NumGPUs == -1 {
		p.usage.NumGPUs = len(p.OS.GPU)
	} else {
		p.usage.NumGPUs += usage.NumGPUs
	}

	return nil
}

func (p *Pool) PollConsume(usage *Resources) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pollConsume(usage)
}

func (p *Pool) pollConsume(usage *Resources) error {
	if usage.NumGPUs == -1 {
		usage.NumGPUs = len(p.OS.GPU)
	}

	free := NewResources(
		p.OS.LogicalCPUCount()-p.usage.NumCPUs,
		int64(p.OS.Memory.Device.Total)-p.usage.Memory,
		len(p.OS.GPU)-p.usage.NumGPUs,
	)

	if usage.NumCPUs > free.NumCPUs {
		return ErrNotEnoughCPU
	}
	if usage.Memory > free.Memory {
		return ErrNotEnoughMemory
	}
	if usage.NumGPUs > free.NumGPUs {
		return ErrNotEnoughGPU
	}

	return nil
}

func (p *Pool) Release(usage *Resources) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.usage.NumCPUs -= usage.NumCPUs
	p.usage.Memory -= usage.Memory
	p.usage.NumGPUs -= usage.NumGPUs
}
