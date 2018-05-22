package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"gopkg.in/yaml.v2"
)

const (
	OutputModeSimple = "simple"
	OutputModeJSON   = "json"
	homeConfigDir    = ".sonm"
	configName       = "cli.yaml"
)

// cliConfig implements Config interface
type Config struct {
	Eth       accounts.EthConfig `yaml:"ethereum"`
	OutFormat string             `required:"false" default:"" yaml:"output_format"`
	path      string
}

func NewConfig(p ...string) (*Config, error) {
	cfgPath, err := getConfigPath(p...)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		path: cfgPath,
	}

	// If config does not exist - use default values.
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfg.fillWithDefaults()
		cfg.Save()
		return cfg, nil
	}

	err = configor.Load(cfg, cfgPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cc *Config) Save() error {
	data, err := yaml.Marshal(cc)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(cc.path, data, os.FileMode(0600))
}

func (cc *Config) OutputFormat() string {
	return cc.OutFormat
}

func (cc *Config) PassPhrase() string {
	return cc.Eth.Passphrase
}

func (cc *Config) KeyStore() string {
	return cc.Eth.Keystore
}

func (cc *Config) fillWithDefaults() {
	cc.OutFormat = OutputModeSimple
}

func getDefaultConfigDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := path.Join(currentUser.HomeDir, homeConfigDir)
	return dir, nil
}

func getConfigPath(p ...string) (string, error) {
	var cfgPath string
	var err error

	if len(p) > 0 && p[0] != "" {
		cfgPath = p[0]
	} else {
		cfgPath, err = getDefaultConfigDir()
		if err != nil {
			return "", err
		}
	}

	cfgPath = path.Join(cfgPath, configName)
	return cfgPath, nil
}
