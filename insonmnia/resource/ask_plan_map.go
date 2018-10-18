package resource

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/proto"
)

type askPlanMap map[string]*sonm.AskPlan

func (m askPlanMap) Sum() (*sonm.AskPlanResources, error) {
	sum := &sonm.AskPlanResources{}
	for _, askPlan := range m {
		if err := sum.Add(askPlan.GetResources()); err != nil {
			return nil, err
		}
	}
	return sum, nil
}

func (m askPlanMap) PopLatest() (*sonm.AskPlan, error) {
	if len(m) == 0 {
		return nil, errors.New("failed to pop latest ask plan: ask plan map is empty")
	}
	lastTs := time.Unix(0, 0)
	var lastPlan *sonm.AskPlan
	for _, plan := range m {
		planTime := plan.GetCreateTime().Unix()
		if !planTime.Before(lastTs) {
			lastTs = planTime
			lastPlan = plan
		}
	}
	delete(m, lastPlan.GetID())
	return lastPlan, nil
}
