package resource

import (
	"errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	"sync"
)

type Resources struct {
	NumCPUs int
	Memory  int64
	// TODO: It's unclear how to calculate GPU usage.
}

func NewResources(numCPUs int, memory int64) Resources {
	return Resources{
		NumCPUs: numCPUs,
		Memory:  memory,
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
// TODO: May be return some kind of Retainer to be able to auto-retain?
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

	return nil
}

func (p *Pool) PollConsume(usage *Resources) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.pollConsume(usage)
}

func (p *Pool) pollConsume(usage *Resources) error {
	free := NewResources(p.OS.LogicalCPUCount()-p.usage.NumCPUs, int64(p.OS.Memory.Total)-p.usage.Memory)

	if usage.NumCPUs > free.NumCPUs {
		return errors.New("not enough CPU available")
	}

	if usage.Memory > free.Memory {
		return errors.New("not enough memory available")
	}

	return nil
}

func (p *Pool) Retain(usage *Resources) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.usage.NumCPUs -= usage.NumCPUs
	p.usage.Memory -= usage.Memory
}
