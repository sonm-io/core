package task_config

import (
	"io/ioutil"
	"os"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/config"
	"gopkg.in/yaml.v2"
)

func LoadConfig(path string) (*sonm.TaskSpec, error) {
	// Manual renaming from snake_case to lowercase fields here to be able to
	// load them directly in the protobuf.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	ty := map[interface{}]interface{}{}
	if err := yaml.Unmarshal(data, ty); err != nil {
		return nil, err
	}

	config.SnakeToLower(ty)

	data, err = yaml.Marshal(ty)
	if err != nil {
		return nil, err
	}

	cfg := &sonm.TaskSpec{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
