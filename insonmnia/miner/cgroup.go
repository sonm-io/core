package miner

import (
	"errors"
	"sync"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	errNamedCgroupAlreadyExists = errors.New("named cgroup already exists")
	errNamedCgroupNotExists     = errors.New("named cgroup not exists")
)

type cGroup interface {
	// New creates a new cgroup under the calling cgroup.
	New(name string, resources *specs.LinuxResources) (cGroup, error)
	Delete() error
}

type cGroupManager interface {
	// Parent returns a parent cgroup.
	Parent() cGroup
	// Attach attaches a new named cgroup, whose parent cgroup will be the main
	// one.
	Attach(name string, resources *specs.LinuxResources) error
	// Detach detaches and deletes a named cgroup.
	Detach(name string) error
}

type controlGroupManager struct {
	parent cGroup
	mu     sync.Mutex
	named  map[string]cGroup
}

func newCgroupManager(parent cGroup) cGroupManager {
	return &controlGroupManager{
		parent: parent,
		named:  map[string]cGroup{},
	}
}

func (c *controlGroupManager) Parent() cGroup {
	return c.parent
}

func (c *controlGroupManager) Attach(name string, resources *specs.LinuxResources) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.named[name]; exists {
		return errNamedCgroupAlreadyExists
	}
	cgroup, err := c.parent.New(name, resources)
	if err != nil {
		return err
	}

	c.named[name] = cgroup
	return nil
}

func (c *controlGroupManager) Detach(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cgroup, exists := c.named[name]
	if !exists {
		return errNamedCgroupNotExists
	}
	delete(c.named, name)
	if err := cgroup.Delete(); err != nil {
		return err
	}

	return nil
}
