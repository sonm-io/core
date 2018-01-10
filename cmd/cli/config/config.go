package config

import (
	"os"
	"os/user"
	"path"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
)

const (
	OutputModeSimple = "simple"
	OutputModeJSON   = "json"
	homeConfigDir    = ".sonm"
	configName       = "cli.yaml"
)

type Config interface {
	OutputFormat() string
	// KeyStorager included into config because of
	// cli instance must know how to open the keystore
	accounts.KeyStorager
}

// cliConfig implements Config interface
type cliConfig struct {
	Eth       accounts.EthConfig `yaml:"ethereum"`
	OutFormat string             `required:"false" default:"" yaml:"output_format"`
}

func (cc *cliConfig) OutputFormat() string {
	return cc.OutFormat
}

func (cc *cliConfig) PassPhrase() string {
	return cc.Eth.Passphrase
}

func (cc *cliConfig) KeyStore() string {
	return cc.Eth.Keystore
}

func (cc *cliConfig) getDefaultConfigDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	dir := path.Join(currentUser.HomeDir, homeConfigDir)
	return dir, nil
}

func (cc *cliConfig) getConfigPath(p ...string) (string, error) {
	var cfgPath string
	var err error

	if len(p) > 0 && p[0] != "" {
		cfgPath = p[0]
	} else {
		cfgPath, err = cc.getDefaultConfigDir()
		if err != nil {
			return "", err
		}
	}

	cfgPath = path.Join(cfgPath, configName)
	return cfgPath, nil
}

func (cc *cliConfig) fillWithDefaults() {
	cc.OutFormat = OutputModeSimple
}

func NewConfig(p ...string) (Config, error) {
	cfg := &cliConfig{}

	cfgPath, err := cfg.getConfigPath(p...)
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
