package accounts

import (
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/mitchellh/go-homedir"
	"github.com/sonm-io/core/util"
)

func OpenSingleKeystore(path, pass string, pf PassPhraser) (*ecdsa.PrivateKey, error) {
	var acc accounts.Account
	var err error

	path, err = getKeystoreDir(path)
	if err != nil {
		return nil, err
	}

	ks := keystore.NewKeyStore(path, keystore.LightScryptN, keystore.LightScryptP)

	pass, err = getPassPhrase(pass, pf)
	if err != nil {
		return nil, err
	}

	if len(ks.Accounts()) == 0 {
		acc, err = ks.NewAccount(pass)
		if err != nil {
			return nil, err
		}
	} else {
		acc = ks.Accounts()[0]
	}

	key, err := decryptKeyFile(acc.URL.Path, pass)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func getPassPhrase(pass string, pf PassPhraser) (string, error) {
	var err error
	if len(pass) == 0 {
		// fallback to passPhrase reader if no password provided
		pass, err = pf.GetPassPhrase()
		if err != nil {
			return "", fmt.Errorf("cannot read pass phrase: %v", err)
		}
	}

	return pass, nil
}

func getKeystoreDir(path string) (string, error) {
	var err error
	path, err = homedir.Expand(path)
	if err != nil {
		return "", fmt.Errorf("cannot expand path `%s`: %v", path, err)
	}

	// Use default key store dir if not specified in config.
	if path == "" {
		path, err = util.GetDefaultKeyStoreDir()
		if err != nil {
			return "", fmt.Errorf("cannot obtain default keystore dir: %v", err)
		}
	}

	return path, nil
}

func decryptKeyFile(path string, pass string) (*ecdsa.PrivateKey, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open account file: %v", err)
	}

	key, err := keystore.DecryptKey(file, pass)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt key with given pass phrase: %v", err)
	}

	return key.PrivateKey, nil
}
