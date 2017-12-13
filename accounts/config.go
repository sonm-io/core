package accounts

import (
	"crypto/ecdsa"
)

type EthConfig struct {
	Passphrase string `required:"false" default:"" yaml:"pass_phrase"`
	Keystore   string `required:"false" default:"" yaml:"key_store"`
}

func (c *EthConfig) LoadKey() (*ecdsa.PrivateKey, error) {
	key, err := LoadKeys(c.Keystore, c.Passphrase)
	if err != nil {
		return nil, err
	}

	return key, nil
}
