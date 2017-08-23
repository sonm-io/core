package gateway

import (
	"errors"

	"gopkg.in/oleiade/lane.v1"
	"math/rand"
	"sync"
)

type PortPool struct {
	queue *lane.Queue
	used  map[string]uint16
	mu    sync.Mutex
}

func NewPortPool(init, size uint16) *PortPool {
	p := &PortPool{
		queue: lane.NewQueue(),
		used:  make(map[string]uint16, size),
	}

	for _, v := range rand.Perm(int(size)) {
		p.queue.Enqueue(init + uint16(v))
	}

	return p
}

func (p *PortPool) Assign(ID string) (uint16, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.used[ID]; exists {
		return 0, errors.New("named port is already in use")
	}

	port := p.queue.Dequeue()
	if port == nil {
		return 0, errors.New("no ports left for allocation")
	}

	p.used[ID] = port.(uint16)

	return port.(uint16), nil
}

func (p *PortPool) Retain(ID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if port, exists := p.used[ID]; exists {
		p.queue.Enqueue(port)
		delete(p.used, ID)
	} else {
		return errors.New("named port was never assigned")
	}

	return nil
}
