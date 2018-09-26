package resource

import (
	"fmt"

	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/proto"
)

type pool struct {
	all *sonm.AskPlanResources
	// maps ask plan ID to ask plan for each specific pool
	usedSpot     askPlanMap
	usedFw       askPlanMap
	commitedSpot askPlanMap
	commitedFw   askPlanMap
}

func newPool(resources *sonm.AskPlanResources) *pool {
	return &pool{
		all:          resources,
		usedSpot:     askPlanMap{},
		usedFw:       askPlanMap{},
		commitedSpot: askPlanMap{},
		commitedFw:   askPlanMap{},
	}
}

func (m *pool) consume(plan *sonm.AskPlan) error {
	available := deepcopy.Copy(m.all).(*sonm.AskPlanResources)
	commitedSum, err := m.commitedFw.Sum()
	if err != nil {
		return err
	}
	if err := available.Sub(commitedSum); err != nil {
		return err
	}
	var used askPlanMap
	if plan.IsSpot() {
		used = m.usedSpot
	} else {
		used = m.usedFw
	}
	usedSum, err := used.Sum()
	if err != nil {
		return err
	}
	if err := available.Sub(usedSum); err != nil {
		return err
	}
	if err := available.CheckContains(plan.GetResources()); err != nil {
		return err
	}
	used[plan.ID] = plan

	return nil
}

func (m *pool) release(ID string) error {
	err := fmt.Errorf("could not release resource with ID %s from pool - no such resource", ID)
	for _, mapping := range []askPlanMap{m.usedFw, m.usedSpot, m.commitedFw, m.commitedSpot} {
		if _, ok := mapping[ID]; ok {
			delete(mapping, ID)
			err = nil
		}
	}
	return err
}

func ejectAskPlans(required *sonm.AskPlanResources, available *sonm.AskPlanResources, pool askPlanMap) ([]string, error) {
	ids := []string{}
	for {
		if err := required.CheckContains(available); err == nil {
			return ids, nil
		}
		plan, err := pool.PopLatest()
		if err != nil {
			return nil, err
		}
		ids = append(ids, plan.GetID())
	}
}

func (m *pool) shrinkSpotPool(plan *sonm.AskPlan) ([]string, error) {
	if plan.IsSpot() {
		return []string{}, nil
	}
	available := deepcopy.Copy(m.all).(*sonm.AskPlanResources)
	spotSum, err := m.usedSpot.Sum()
	if err != nil {
		return nil, err
	}
	if err := available.Sub(spotSum); err != nil {
		return nil, err
	}

	required, err := m.commitedFw.Sum()
	if err != nil {
		return nil, err
	}
	required.Add(plan.GetResources())

	return ejectAskPlans(required, available, m.usedSpot)
}

func (m *pool) shrinkCommitedSpotPool(plan *sonm.AskPlan) ([]string, error) {
	ejectedPlans := []string{}
	available := deepcopy.Copy(m.all).(*sonm.AskPlanResources)
	commitedSum, err := m.commitedFw.Sum()
	if err != nil {
		return ejectedPlans, err
	}
	if err := available.Sub(commitedSum); err != nil {
		return ejectedPlans, err
	}
	return ejectAskPlans(plan.GetResources(), available, m.commitedSpot)
}

func (m *pool) commit(plan *sonm.AskPlan) {
	var target askPlanMap
	if plan.IsSpot() {
		target = m.commitedSpot
	} else {
		target = m.commitedFw
	}
	target[plan.ID] = plan
}

func (m *pool) makeRoomAndCommit(plan *sonm.AskPlan) ([]string, error) {

	ejectedPlans, err := m.shrinkSpotPool(plan)
	if err != nil {
		return nil, err
	}

	ejectedCommited, err := m.shrinkCommitedSpotPool(plan)
	if err != nil {
		return nil, err
	}
	ejectedPlans = append(ejectedPlans, ejectedCommited...)

	// TODO: do we need to free it? or only spot?
	if err := m.release(plan.ID); err != nil {
		return nil, err
	}
	m.commit(plan)
	for _, id := range ejectedPlans {
		if err := m.release(id); err != nil {
			//this should never happen
			//TODO: log
			return nil, err
		}
	}

	return ejectedPlans, nil
}

func (m *pool) getCommitedFree() (*sonm.AskPlanResources, error) {
	resources := deepcopy.Copy(m.all).(*sonm.AskPlanResources)
	for _, plan := range m.commitedSpot {
		if err := resources.Sub(plan.GetResources()); err != nil {
			return nil, err
		}
	}
	for _, plan := range m.commitedFw {
		if err := resources.Sub(plan.GetResources()); err != nil {
			return nil, err
		}
	}
	return resources, nil
}
