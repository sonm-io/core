package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util"
	"gopkg.in/yaml.v2"
)

const (
	OutputModeSimple = "simple"
	OutputModeJSON   = "json"
	configName       = "cli.yaml"
)

// Config describes configuration file for the `sonmcli` tool
type Config struct {
	Eth        accounts.EthConfig `yaml:"ethereum"`
	OutFormat  string             `required:"false" default:"" yaml:"output_format"`
	WorkerAddr string             `yaml:"worker_eth_addr"`
	NodeAddr   string             `yaml:"node_addr"`
	path       string
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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cc *Config) Validate() error {
	if len(cc.WorkerAddr) > 0 {
		if _, err := auth.ParseAddr(cc.WorkerAddr); err != nil {
			return fmt.Errorf("failed to parse worker address: %s", err)
		}
	}

	return nil
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

func getConfigPath(p ...string) (string, error) {
	var cfgPath string
	var err error

	if len(p) > 0 && p[0] != "" {
		cfgPath = p[0]
	} else {
		cfgPath, err = util.GetDefaultConfigDir()
		if err != nil {
			return "", err
		}
	}

	if util.FileExists(cfgPath) == nil {
		return cfgPath, nil
	}

	return path.Join(cfgPath, configName), nil
}
