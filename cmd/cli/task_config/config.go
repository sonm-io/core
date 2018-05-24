package task_config

import (
	"github.com/jinzhu/configor"
	"github.com/sevlyar/retag"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/config"
)

func LoadConfig(path string) (*sonm.StartTaskRequest, error) {
	cfg := &sonm.StartTaskRequest{}
	cfgView := retag.Convert(cfg, config.SnakeCaseTagger("yaml"))

	err := configor.Load(cfgView, path)
	if err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
