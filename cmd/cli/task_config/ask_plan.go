package task_config

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/proto"
)

func LoadAskPlan(path string) (*sonm.AskPlan, error) {
	ask := &sonm.AskPlan{}
	if err := configor.Load(ask, path); err != nil {
		return nil, err
	}

	if err := ask.Validate(); err != nil {
		return nil, err
	}

	return ask, nil
}
