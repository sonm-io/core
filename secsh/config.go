package secsh

import (
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/npp"
)

type Config struct {
	SecExecPath      string             `yaml:"secexec"`
	SeccompPolicyDir string             `yaml:"seccomp_policy_dir"`
	AllowedKeys      []string           `yaml:"allowed_keys"`
	Eth              accounts.EthConfig `yaml:"ethereum"`
	NPP              npp.Config         `yaml:"npp"`
}
