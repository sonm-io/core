package password

import (
	"os/user"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zalando/go-keyring"
)

type PasspharerConfig struct {
	Type string `yaml:"type" default:"config"`
	Args map[string]string
}

type EthConfig struct {
	Passphraser PasspharerConfig `required:"false" default:"" yaml:"pass_phraser"`
	Keystore    string           `required:"false" default:"" yaml:"key_store"`
	Address     common.Address   `required:"false"`
}

type ConfigPasswordReader struct {
	cfg *EthConfig
}

func NewConfigPasswordReader(cfg *EthConfig) *ConfigPasswordReader {
	return &ConfigPasswordReader{}
}

func (m *ConfigPasswordReader) ReadPassword(address common.Address) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	secret, err := keyring.Get(sonmServicePrefix+address.Hex(), user.Username)
	if err != nil {
		if err == keyring.ErrNotFound {
			pw, err := m.cli.ReadPassword(address)
			if err != nil {
				return "", err
			}
			keyring.Set(sonmServicePrefix+address.Hex(), user.Username, string(pw))
			return string(pw), nil
		}
		return "", err
	}
	return secret, nil
}

func (m *ConfigPasswordReader) ForgetPassword(address common.Address) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	return keyring.Delete(sonmServicePrefix+address.Hex(), user.Username)
}
