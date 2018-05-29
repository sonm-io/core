package task_config

import (
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/config"
)

func LoadConfig(path string) (*sonm.TaskSpec, error) {
	// Manual renaming from snake_case to lowercase fields here to be able to
	// load them directly in the protobuf.
	cfg := &sonm.TaskSpec{}
	if err := config.LoadWith(cfg, path, config.SnakeToLower); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
