package task_config

import (
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/config"
)

func LoadConfig(path string) (*sonm.TaskSpec, error) {
	cfg := &sonm.TaskSpec{}
	if err := config.FromFile(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
