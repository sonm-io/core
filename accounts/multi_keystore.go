package accounts

import (
	"crypto/ecdsa"
	"io/ioutil"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"
)

type KeystoreConfig struct {
	KeyDir      string            `yaml:"key_store" required:"true"`
	PassPhrases map[string]string `yaml:"pass_phrases"`
}

func (cfg *KeystoreConfig) getStateFileDir() string {
	return path.Join(cfg.KeyDir, "state")
}

func (cfg *KeystoreConfig) getStateFilePath() string {
	return path.Join(cfg.getStateFileDir(), "data")
}

type multiKeystore struct {
	cfg        *KeystoreConfig
	keyStore   *keystore.KeyStore
	passReader PassPhraser
}

func NewMultiKeystore(cfg *KeystoreConfig, pf PassPhraser) (*multiKeystore, error) {
	ks := keystore.NewKeyStore(
		cfg.KeyDir,
		keystore.LightScryptN,
		keystore.LightScryptP,
	)

	return &multiKeystore{
		cfg:        cfg,
		keyStore:   ks,
		passReader: pf,
	}, nil
}

// List returns list of accounts addresses into keystore
func (m *multiKeystore) List() ([]accounts.Account, error) {
	return m.keyStore.Accounts(), nil
}

// Generate creates new key into keystore
func (m *multiKeystore) Generate() (*ecdsa.PrivateKey, error) {
	pass, err := m.passReader.GetPassPhrase()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read pass phrase")
	}

	acc, err := m.keyStore.NewAccount(pass)
	if err != nil {
		return nil, err
	}

	if len(m.keyStore.Accounts()) == 1 {
		// this is the first account,
		if err := m.setDefaultAccount(acc.Address); err != nil {
			return nil, err
		}
	}

	return m.readAccount(acc)
}

// GetKeyByAddress loads and decrypts key form keystore (if present)
func (m *multiKeystore) GetKeyByAddress(addr common.Address) (*ecdsa.PrivateKey, error) {
	for _, acc := range m.keyStore.Accounts() {
		if acc.Address.Big().Cmp(addr.Big()) == 0 {
			return m.readAccount(acc)
		}
	}

	return nil, errors.New("cannot find given address into keystore")
}

// GetDefault returns default key for the keystore
func (m *multiKeystore) GetDefault() (*ecdsa.PrivateKey, error) {
	if len(m.keyStore.Accounts()) == 0 {
		return nil, errors.New("no accounts present into keystore")
	}

	defaultAddr, err := m.getDefaultAddress()
	if err != nil {
		// no keys marked as default
		return nil, errors.Wrap(err, "cannot load default key")
	}

	return m.GetKeyByAddress(defaultAddr)
}

// SetDefault marks key as default for keystore
func (m *multiKeystore) SetDefault(addr common.Address) error {
	return m.setDefaultAccount(addr)
}

// Import imports exiting key file into keystore
func (m *multiKeystore) Import(path string) (common.Address, error) {
	if !util.FileExists(path) {
		return common.Address{}, errors.New("key file does not exists")
	}

	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "cannot read key file")
	}

	pass, err := m.passReader.GetPassPhrase()
	if err != nil {
		return common.Address{}, errors.Wrap(err, "cannot read pass phrase")
	}

	acc, err := m.keyStore.Import(keyData, pass, pass)
	if err != nil {
		return common.Address{}, errors.Wrap(err, "cannot import key file")
	}

	return acc.Address, nil
}

func (m *multiKeystore) readAccount(acc accounts.Account) (*ecdsa.PrivateKey, error) {
	file, err := ioutil.ReadFile(acc.URL.Path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open account file")
	}

	var pass string
	pass, ok := m.cfg.PassPhrases[acc.Address.Hex()]
	if !ok {
		pass, err = m.passReader.GetPassPhrase()
		if err != nil {
			return nil, errors.Wrap(err, "cannot read pass phrase")
		}
	}

	key, err := keystore.DecryptKey(file, pass)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decrypt key with given pass phrase")
	}

	return key.PrivateKey, nil
}

func (m *multiKeystore) getDefaultAddress() (common.Address, error) {
	if !util.FileExists(m.cfg.getStateFilePath()) {
		return common.Address{}, errors.New("cannot find keystore's state")
	}

	data, err := ioutil.ReadFile(m.cfg.getStateFilePath())
	if err != nil {
		return common.Address{}, errors.New("cannot read state file")
	}

	addr, err := util.HexToAddress(string(data))
	if err != nil {
		return common.Address{}, err
	}

	return addr, nil
}

func (m *multiKeystore) setDefaultAccount(addr common.Address) error {
	if !util.FileExists(m.cfg.getStateFileDir()) {
		if err := os.MkdirAll(m.cfg.getStateFileDir(), 0700); err != nil {
			return errors.WithMessage(err, "cannot create dir for state")
		}
	}

	if err := ioutil.WriteFile(m.cfg.getStateFilePath(), []byte(addr.Hex()), 0600); err != nil {
		return errors.WithMessage(err, "cannot write state to file")
	}

	return nil
}
