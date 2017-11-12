package miner

import (
	"errors"
	"sync"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	errCgroupAlreadyExists = errors.New("nested cgroup already exists")
	errCgroupNotExists     = errors.New("nested cgroup not exists")
)

type cGroup interface {
	// New creates a new cgroup under the calling cgroup.
	New(name string, resources *specs.LinuxResources) (cGroup, error)
	// Delete removes the cgroup as a whole.
	Delete() error
	Suffix() string
}

type cGroupManager interface {
	// Parent returns a parent cgroup.
	Parent() cGroup
	// Attach attaches a new nested cgroup under the parent cgroup, returning
	// that new cgroup handle.
	Attach(name string, resources *specs.LinuxResources) (cGroup, error)
	// Detach detaches and deletes a nested cgroup.
	Detach(name string) error
}

type nilCgroup struct{}

func (c *nilCgroup) New(name string, resources *specs.LinuxResources) (cGroup, error) {
	return c, nil
}

func (*nilCgroup) Delete() error {
	return nil
}

func (*nilCgroup) Suffix() string {
	return ""
}

type nilGroupManager struct{}

func newNilCgroupManager() (cGroup, cGroupManager, error) {
	return &nilCgroup{}, &nilGroupManager{}, nil
}

func (c *nilGroupManager) Parent() cGroup {
	return &nilCgroup{}
}

func (c *nilGroupManager) Attach(name string, resources *specs.LinuxResources) (cGroup, error) {
	return &nilCgroup{}, nil
}

func (c *nilGroupManager) Detach(name string) error {
	return nil
}

type controlGroupManager struct {
	parent     cGroup
	parentName string
	mu         sync.Mutex
	nested     map[string]cGroup
}

// NewCgroupManager initializes a new cgroup with a nested group manager
// associated with it.
func newCgroupManager(name string, resources *specs.LinuxResources) (cGroup, cGroupManager, error) {
	parent, err := initializeControlGroup(name, resources)
	if err != nil {
		return nil, nil, err
	}

	manager := &controlGroupManager{
		parent:     parent,
		parentName: name,
		nested:     map[string]cGroup{},
	}

	return parent, manager, nil
}

func (c *controlGroupManager) Parent() cGroup {
	return c.parent
}

func (c *controlGroupManager) Attach(name string, resources *specs.LinuxResources) (cGroup, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cgroup, exists := c.nested[name]; exists {
		return cgroup, nil
	}

	cgroup, err := c.parent.New(name, resources)
	if err != nil {
		return nil, err
	}

	c.nested[name] = cgroup
	return cgroup, nil
}

func (c *controlGroupManager) Detach(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cgroup, exists := c.nested[name]
	if !exists {
		return errCgroupNotExists
	}
	delete(c.nested, name)
	if err := cgroup.Delete(); err != nil {
		return err
	}

	return nil
}
