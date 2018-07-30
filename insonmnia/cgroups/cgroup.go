package cgroups

import (
	"errors"
	"sync"

	"github.com/opencontainers/runtime-spec/specs-go"
)

var (
	ErrCgroupAlreadyExists = errors.New("nested cgroup already exists")
	ErrCgroupNotExists     = errors.New("nested cgroup does not exist")
)

type Process struct {
	Subsystem string
	Pid       int
	Path      string
}

type Stats struct {
	MemoryLimit uint64
}

type CGroup interface {
	// New creates a new cgroup under the calling cgroup.
	New(name string, resources *specs.LinuxResources) (CGroup, error)
	// Add adds a process to the cgroup
	Add(Process) error
	// Delete removes the cgroup as a whole.
	Delete() error
	// Stats returns resource description
	Stats() (*Stats, error)

	Suffix() string
}

type CGroupManager interface {
	// Parent returns a parent cgroup.
	Parent() CGroup
	// Attach attaches a new nested cgroup under the parent cgroup, returning
	// that new cgroup handle.
	// Also returns cgroup handle if it is already exists.
	Attach(name string, resources *specs.LinuxResources) (CGroup, error)
	// Detach detaches and deletes a nested cgroup.
	Detach(name string) error
}

type nilCgroup struct{}

func (c *nilCgroup) New(name string, resources *specs.LinuxResources) (CGroup, error) {
	return c, nil
}

func (*nilCgroup) Add(process Process) error {
	return nil
}

func (*nilCgroup) Delete() error {
	return nil
}

func (*nilCgroup) Suffix() string {
	return ""
}

func (*nilCgroup) Stats() (*Stats, error) {
	return nil, errors.New("unimplemented")
}

type nilGroupManager struct{}

func newNilCgroupManager() (CGroup, CGroupManager, error) {
	return &nilCgroup{}, &nilGroupManager{}, nil
}

func (c *nilGroupManager) Parent() CGroup {
	return &nilCgroup{}
}

func (c *nilGroupManager) Attach(name string, resources *specs.LinuxResources) (CGroup, error) {
	return &nilCgroup{}, nil
}

func (c *nilGroupManager) Detach(name string) error {
	return nil
}

type controlGroupManager struct {
	parent     CGroup
	parentName string
	mu         sync.Mutex
	nested     map[string]CGroup
}

// NewCgroupManager initializes a new cgroup with a nested group manager
// associated with it.
func newCgroupManager(name string, resources *specs.LinuxResources) (CGroup, CGroupManager, error) {
	parent, err := initializeControlGroup(name, resources)
	if err != nil {
		return nil, nil, err
	}

	manager := &controlGroupManager{
		parent:     parent,
		parentName: name,
		nested:     map[string]CGroup{},
	}

	return parent, manager, nil
}

func (c *controlGroupManager) Parent() CGroup {
	return c.parent
}

func (c *controlGroupManager) Attach(name string, resources *specs.LinuxResources) (CGroup, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cgroup, exists := c.nested[name]; exists {
		return cgroup, ErrCgroupAlreadyExists
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
		return ErrCgroupNotExists
	}
	delete(c.nested, name)
	if err := cgroup.Delete(); err != nil {
		return err
	}

	return nil
}

func NewCgroupManager(name string, res *specs.LinuxResources) (CGroup, CGroupManager, error) {
	if !platformSupportCGroups || res == nil {
		return newNilCgroupManager()
	}
	return newCgroupManager(name, res)
}
