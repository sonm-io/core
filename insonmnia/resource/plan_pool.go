package resource

import (
	"fmt"

	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type pool struct {
	log *zap.SugaredLogger
	all *sonm.AskPlanResources
	// maps ask plan ID to ask plan for each specific pool
	usedSpot     askPlanMap
	usedFw       askPlanMap
	commitedSpot askPlanMap
	commitedFw   askPlanMap
	ejectedPlans askPlanMap
}

func newPool(log *zap.SugaredLogger, resources *sonm.AskPlanResources) *pool {
	return &pool{
		log:          log,
		all:          resources,
		usedSpot:     askPlanMap{},
		usedFw:       askPlanMap{},
		commitedSpot: askPlanMap{},
		commitedFw:   askPlanMap{},
		ejectedPlans: askPlanMap{},
	}
}

func (m *pool) Consume(plan *sonm.AskPlan) error {
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

func (m *pool) Release(ID string) error {
	for _, mapping := range []askPlanMap{m.usedFw, m.usedSpot, m.commitedFw, m.commitedSpot, m.ejectedPlans} {
		if _, ok := mapping[ID]; ok {
			delete(mapping, ID)
			return nil
		}
	}
	return fmt.Errorf("could not release resource with ID %s from pool - no such resource", ID)
}

func (m *pool) ejectAskPlans(required *sonm.AskPlanResources, available *sonm.AskPlanResources, pool askPlanMap) ([]string, error) {
	plans := []*sonm.AskPlan{}
	for {
		if err := available.CheckContains(required); err == nil {
			ids := []string{}
			for _, plan := range plans {
				ids = append(ids, plan.ID)
				m.ejectedPlans[plan.ID] = plan
			}
			return ids, nil
		}
		plan, err := pool.PopLatest()
		if err != nil {
			for _, plan := range plans {
				pool[plan.ID] = plan
			}
			return nil, err
		}
		available.Add(plan.GetResources())
		m.log.Infof("popped out plan %s", plan.ID)
		plans = append(plans, plan)
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

	return m.ejectAskPlans(required, available, m.usedSpot)
}

func (m *pool) shrinkCommitedSpotPool(plan *sonm.AskPlan) ([]string, error) {
	ejectedPlans := []string{}
	available := deepcopy.Copy(m.all).(*sonm.AskPlanResources)
	commitedFwSum, err := m.commitedFw.Sum()
	if err != nil {
		return ejectedPlans, err
	}
	commitedSpotSum, err := m.commitedSpot.Sum()
	if err != nil {
		return ejectedPlans, err
	}
	if err := available.Sub(commitedFwSum); err != nil {
		return ejectedPlans, err
	}
	if err := available.Sub(commitedSpotSum); err != nil {
		return ejectedPlans, err
	}
	return m.ejectAskPlans(plan.GetResources(), available, m.commitedSpot)
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

func (m *pool) MakeRoomAndCommit(plan *sonm.AskPlan) ([]string, error) {

	ejectedPlans, err := m.shrinkSpotPool(plan)
	if err != nil {
		return nil, err
	}
	m.log.Info("shrinked spot pool")

	ejectedCommited, err := m.shrinkCommitedSpotPool(plan)
	if err != nil {
		return nil, err
	}
	m.log.Info("shrinked comited spot pool")
	ejectedPlans = append(ejectedPlans, ejectedCommited...)

	// TODO: do we need to free it? or only spot?
	if err := m.Release(plan.ID); err != nil {
		return nil, err
	}

	m.commit(plan)

	return ejectedPlans, nil
}

func (m *pool) GetCommitedFree() (*sonm.AskPlanResources, error) {
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

func (m *pool) ToProto() *sonm.AskPlanPool {
	return &sonm.AskPlanPool{
		All:          m.all,
		UsedSpot:     deepcopy.Copy(m.usedSpot).(askPlanMap),
		UsedFw:       deepcopy.Copy(m.usedFw).(askPlanMap),
		CommitedSpot: deepcopy.Copy(m.commitedSpot).(askPlanMap),
		CommitedFw:   deepcopy.Copy(m.commitedFw).(askPlanMap),
		EjectedPlans: deepcopy.Copy(m.ejectedPlans).(askPlanMap),
	}
}
