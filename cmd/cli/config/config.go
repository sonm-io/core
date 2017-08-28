package config

import (
	"os"
	"os/user"
	"path"

	"github.com/jinzhu/configor"
)

const (
	OutputModeSimple = "simple"
	OutputModeJSON   = "json"
	homeConfigPath   = ".sonm/cli.yaml"
)

type Config interface {
	OutputFormat() string
	HubAddress() string
}

// cliConfig implements Config interface
type cliConfig struct {
	HubAddr   string `required:"false" default:"" yaml:"hub_address"`
	OutFormat string `required:"false" default:"" yaml:"output_format"`
}

func (cc *cliConfig) OutputFormat() string {
	return cc.OutFormat
}

func (cc *cliConfig) HubAddress() string {
	return cc.HubAddr
}

func (cc *cliConfig) getConfigPath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	cfgPath := path.Join(currentUser.HomeDir, homeConfigPath)
	return cfgPath, nil
}

func (cc *cliConfig) fillWithDefaults() {
	cc.OutFormat = OutputModeSimple
	cc.HubAddr = ""
}

func NewConfig() (Config, error) {
	cfg := &cliConfig{}
	cfgPath, err := cfg.getConfigPath()
	if err != nil {
		return nil, err
	}

	// if config does not exists - use default values
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg.fillWithDefaults()
		return cfg, nil
	}

	err = configor.Load(cfg, cfgPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
