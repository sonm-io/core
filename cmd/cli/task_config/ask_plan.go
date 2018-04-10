package task_config

import (
	"errors"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/proto"
)

const (
	minRamSize    = 4 * 1024 * 1024
	minCPUPercent = 1
)

func validate(ask *sonm.AskPlan) error {
	if ask.GetResources().GetCPU().GetCores() < minCPUPercent {
		return errors.New("CPU count is too low")
	}

	if ask.GetResources().GetRAM().GetSize().GetSize() < minRamSize {
		return errors.New("RAM size is too low")
	}

	return nil
}

func LoadAskPlan(p string) (*sonm.AskPlan, error) {
	ask := &sonm.AskPlan{}
	if err := configor.Load(ask, p); err != nil {
		return nil, err
	}

	if err := validate(ask); err != nil {
		return nil, err
	}

	return ask, nil
}
