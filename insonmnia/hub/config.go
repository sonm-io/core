package hub

import (
	"github.com/jinzhu/configor"
)

type LoggingConfig struct {
	Level int `required:"true" default:"1"`
}

type MonitoringConfig struct {
	Endpoint string `required:"true" yaml:"endpoint"`
}

type HubConfig struct {
	Endpoint   string           `required:"true" yaml:"endpoint"`
	Bootnodes  []string         `required:"false" yaml:"bootnodes"`
	Monitoring MonitoringConfig `required:"true" yaml:"monitoring"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// NewConfig loads a hub config from the specified YAML file.
func NewConfig(path string) (*HubConfig, error) {
	conf := &HubConfig{}
	err := configor.Load(conf, path)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// TODO: Currently stubbed for simplifying testing.
type Config interface {
	Endpoint() string
	MonitoringEndpoint() string
	Logging() LoggingConfig
}
