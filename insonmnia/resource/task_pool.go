package resource

import (
	"fmt"

	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/proto"
	yaml "gopkg.in/yaml.v1"
)

type taskPool struct {
	all  *sonm.AskPlanResources
	used map[string]*sonm.AskPlanResources
}

func newTaskPool(resources *sonm.AskPlanResources) *taskPool {
	return nil
}

func (p *taskPool) consume(ID string, resources *sonm.AskPlanResources) error {
	if err := p.pollConsume(resources); err != nil {
		return err
	}
	if _, ok := p.used[ID]; ok {
		return fmt.Errorf("resources with ID %s has been already consumed", ID)
	}

	p.used[ID] = resources
	return nil
}

func (p *taskPool) release(ID string) error {
	if _, ok := p.used[ID]; ok {
		delete(p.used, ID)
		return nil
	}
	return fmt.Errorf("could not release task with ID %s from pool - no such resource", ID)
}

func (p *taskPool) pollConsume(resources *sonm.AskPlanResources) error {
	available, err := p.getFree()
	if err != nil {
		return err
	}
	err = available.Sub(resources)
	if err != nil {
		return fmt.Errorf("not enough resources: %s", err)
	}
	return nil
}

func (p *taskPool) getFree() (*sonm.AskPlanResources, error) {
	res := deepcopy.Copy(p.all).(*sonm.AskPlanResources)
	usage, err := p.getUsage()
	if err != nil {
		return nil, err
	}
	err = res.Sub(usage)
	if err != nil {
		pool, _ := yaml.Marshal(res)
		use, _ := yaml.Marshal(usage)
		return &sonm.AskPlanResources{}, fmt.Errorf("resource pool inconsistency found - used resources are greater than available for scheduling(%s). pool - %s, used - %s",
			err, pool, use)
	}
	return res, nil
}

func (p *taskPool) getUsage() (*sonm.AskPlanResources, error) {
	sum := &sonm.AskPlanResources{}
	for _, askPlan := range p.used {
		if err := sum.Add(askPlan); err != nil {
			return nil, err
		}
	}
	return sum, nil
}
