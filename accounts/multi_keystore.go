package accounts

import (
	"crypto/ecdsa"
	"fmt"
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

type MultiKeystore struct {
	cfg        *KeystoreConfig
	keyStore   *keystore.KeyStore
	passReader PassPhraser
}

func NewMultiKeystore(cfg *KeystoreConfig, pf PassPhraser) (*MultiKeystore, error) {
	ks := keystore.NewKeyStore(
		cfg.KeyDir,
		keystore.LightScryptN,
		keystore.LightScryptP,
	)

	m := &MultiKeystore{
		cfg:        cfg,
		keyStore:   ks,
		passReader: pf,
	}

	// We have only one account, mark it as default
	// for current keystore instance.
	accs := ks.Accounts()
	if len(accs) == 1 {
		m.SetDefault(accs[0].Address)
	}

	return m, nil
}

// List returns list of accounts addresses into keystore
func (m *MultiKeystore) List() []accounts.Account {
	return m.keyStore.Accounts()
}

// Generate creates new key into keystore
func (m *MultiKeystore) Generate() (*ecdsa.PrivateKey, error) {
	pass, err := m.passReader.GetPassPhrase()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read pass phrase")
	}

	return m.GenerateWithPassword(pass)
}

// GenerateWithPassword generates new key with given pass-phrase
func (m *MultiKeystore) GenerateWithPassword(pass string) (*ecdsa.PrivateKey, error) {
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
func (m *MultiKeystore) GetKeyByAddress(addr common.Address) (*ecdsa.PrivateKey, error) {
	acc, err := m.findAccountByAddr(addr)
	if err != nil {
		return nil, err
	}

	return m.readAccount(acc)
}

func (m *MultiKeystore) GetKeyWithPass(addr common.Address, pass string) (*ecdsa.PrivateKey, error) {
	acc, err := m.findAccountByAddr(addr)
	if err != nil {
		return nil, err
	}

	return m.readKeyFile(acc.URL.Path, pass)
}

// GetDefault returns default key for the keystore
func (m *MultiKeystore) GetDefault() (*ecdsa.PrivateKey, error) {
	if len(m.keyStore.Accounts()) == 0 {
		return nil, errors.New("no accounts present into keystore")
	}

	defaultAddr, err := m.GetDefaultAddress()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load default key")
	}

	return m.GetKeyByAddress(defaultAddr)
}

// SetDefault marks key as default for keystore
func (m *MultiKeystore) SetDefault(addr common.Address) error {
	if !m.keyStore.HasAddress(addr) {
		return errors.New("given address does not present into keystore")
	}

	return m.setDefaultAccount(addr)
}

func (m *MultiKeystore) GetDefaultAddress() (common.Address, error) {
	err := util.FileExists(m.cfg.getStateFilePath())
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to access keystore's state: %s", err)
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

func (m *MultiKeystore) findAccountByAddr(addr common.Address) (accounts.Account, error) {
	for _, acc := range m.keyStore.Accounts() {
		if acc.Address.Big().Cmp(addr.Big()) == 0 {
			return acc, nil
		}
	}

	return accounts.Account{}, errors.New("cannot find given address into keystore")
}

func (m *MultiKeystore) readAccount(acc accounts.Account) (*ecdsa.PrivateKey, error) {
	var pass string
	var err error
	pass, ok := m.cfg.PassPhrases[acc.Address.Hex()]
	if !ok {
		pass, err = m.passReader.GetPassPhrase()
		if err != nil {
			return nil, errors.Wrap(err, "cannot read pass phrase")
		}
	}

	return m.readKeyFile(acc.URL.Path, pass)
}

func (m *MultiKeystore) readKeyFile(path string, pass string) (*ecdsa.PrivateKey, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open account file")
	}

	key, err := keystore.DecryptKey(file, pass)
	if err != nil {
		return nil, errors.Wrap(err, "cannot decrypt key with given pass phrase")
	}

	return key.PrivateKey, nil
}

func (m *MultiKeystore) setDefaultAccount(addr common.Address) error {
	if err := util.DirectoryExists(m.cfg.getStateFileDir()); err != nil {
		if err := os.MkdirAll(m.cfg.getStateFileDir(), 0700); err != nil {
			return errors.WithMessage(err, "cannot create dir for state")
		}
	}

	if err := ioutil.WriteFile(m.cfg.getStateFilePath(), []byte(addr.Hex()), 0600); err != nil {
		return errors.WithMessage(err, "cannot write state to file")
	}

	return nil
}
