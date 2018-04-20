package resource

import (
	"fmt"
	"sync"

	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/proto"
)

type Scheduler struct {
	OS            *hardware.Hardware
	mu            sync.Mutex
	pool          *pool
	taskToAskPlan map[string]string
	askPlanPools  map[string]*pool
	parentCGroups map[string]cgroups.CGroup
}

func NewScheduler(hardware *hardware.Hardware) *Scheduler {
	return &Scheduler{
		OS:            hardware,
		pool:          newPool(hardware.AskPlanResources()),
		taskToAskPlan: map[string]string{},
		askPlanPools:  map[string]*pool{},
		parentCGroups: map[string]cgroups.CGroup{},
	}
}

func (m *Scheduler) GetUsage() (*sonm.AskPlanResources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool.getUsage()
}

func (m *Scheduler) GetFree() (*sonm.AskPlanResources, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool.getFree()
}

func (m *Scheduler) CGroup(askPlanID string) (cgroups.CGroup, error) {
	panic("implement me")
}

func (m *Scheduler) PollConsume(askPlan *sonm.AskPlan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pool.pollConsume(askPlan.Resources)
}

// Consume tries to consume the specified resource usage from the pool.
//
// Does nothing on error.
func (m *Scheduler) Consume(askPlan *sonm.AskPlan) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.pool.consume(askPlan.ID, askPlan.Resources)
}

func (m *Scheduler) ConsumeTask(askPlanID string, taskID string, resources *sonm.AskPlanResources) error {
	copy := &sonm.AskPlanResources{
		GPU:     deepcopy.Copy(resources.GPU).(*sonm.AskPlanGPU),
		Storage: deepcopy.Copy(resources.Storage).(*sonm.AskPlanStorage),
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.taskToAskPlan[taskID] = askPlanID
	pool, ok := m.askPlanPools[askPlanID]
	if !ok {
		return fmt.Errorf("could not consume resources for task - ask Plan with id %s not found", askPlanID)
	}

	return pool.consume(taskID, copy)
}

func (m *Scheduler) Release(askPlanID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pool.release(askPlanID)
}

func (m *Scheduler) ReleaseTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	askPlanID, ok := m.taskToAskPlan[taskID]
	if !ok {
		return fmt.Errorf("could not find corresponding ask plan id for task %s", taskID)
	}
	pool, ok := m.askPlanPools[askPlanID]
	if !ok {
		return fmt.Errorf("could not consume resources for task - ask Plan with id %s not found", askPlanID)
	}

	err := pool.release(taskID)
	if err == nil {
		delete(m.taskToAskPlan, taskID)
	}
	return err
}

type pool struct {
	all  *sonm.AskPlanResources
	used map[string]*sonm.AskPlanResources
}

func newPool(resources *sonm.AskPlanResources) *pool {
	return &pool{all: resources}
}

func (p *pool) getFree() (*sonm.AskPlanResources, error) {
	res := deepcopy.Copy(p.all).(*sonm.AskPlanResources)
	usage, err := p.getUsage()
	if err != nil {
		return nil, err
	}
	err = res.Sub(usage)
	if err != nil {
		return &sonm.AskPlanResources{}, fmt.Errorf("resource pool inconsistency found - used resources are greater than available for scheduling(%s)", err)
	}
	return res, nil
}

func (p *pool) getUsage() (*sonm.AskPlanResources, error) {
	sum := &sonm.AskPlanResources{}
	for _, askPlan := range p.used {
		if err := sum.Add(askPlan); err != nil {
			return nil, err
		}
	}
	return sum, nil
}

func (p *pool) pollConsume(resources *sonm.AskPlanResources) error {
	available, err := p.getFree()
	if err != nil {
		return err
	}
	err = available.Sub(resources)
	if err != nil {
		return fmt.Errorf("not enough resources - %s", err)
	}
	return nil
}

func (p *pool) consume(ID string, resources *sonm.AskPlanResources) error {
	if err := p.pollConsume(resources); err != nil {
		return err
	}
	if _, ok := p.used[ID]; ok {
		return fmt.Errorf("resources with ID %s has been already consumed", ID)
	}

	p.used[ID] = resources
	return nil
}

func (p *pool) release(ID string) error {
	if _, ok := p.used[ID]; !ok {
		return fmt.Errorf("could not release resource with ID %s from pool - no such resource", ID)
	}
	delete(p.used, ID)
	return nil
}
