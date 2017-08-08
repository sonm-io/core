package miner

import (
	"github.com/jinzhu/configor"
)

type LoggingConfig struct {
	Level int `required:"true" default:"1"`
}

type HubConfig struct {
	Endpoint  string     `required:"false" yaml:"endpoint"`
	Resources *Resources `required:"false" yaml:"resources"`
}

type config struct {
	HubConfig     HubConfig     `required:"false" yaml:"hub"`
	LoggingConfig LoggingConfig `yaml:"logging"`
}

func (c *config) HubEndpoint() string {
	return c.HubConfig.Endpoint
}

func (c *config) HubResources() *Resources {
	return c.HubConfig.Resources
}

func (c *config) Logging() LoggingConfig {
	return c.LoggingConfig
}

// NewConfig creates a new Miner config from the specified YAML file.
func NewConfig(path string) (Config, error) {
	cfg := &config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Config represents a Miner configuration interface.
type Config interface {
	// HubEndpoint returns a string representation of a Hub endpoint to communicate with.
	HubEndpoint() string
	// HubResources returns resources allocated for a Hub.
	HubResources() *Resources
	// Logging returns logging settings.
	Logging() LoggingConfig
}
