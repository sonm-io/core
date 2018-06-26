package accounts

import (
	"crypto/ecdsa"
)

type EthConfig struct {
	Passphrase string `required:"false" default:"" yaml:"pass_phrase"`
	Keystore   string `required:"false" default:"" yaml:"key_store"`
}

func (c *EthConfig) LoadKey() (*ecdsa.PrivateKey, error) {
	return OpenSingleKeystore(c.Keystore, c.Passphrase, NewInteractivePassPhraser())
}
