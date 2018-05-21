package accounts

import (
	"crypto/ecdsa"
	"io/ioutil"
	"path"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"
)

const (
	defaultStateFileName = "defaults"
)

type KeystoreConfig struct {
	KeyDir      string
	StateFile   string
	PassPhrases map[string]string
}

type multiKeystore struct {
	keysDir     string
	stateDir    string
	keyStore    *keystore.KeyStore
	passReader  PassPhraser
	knownPasses map[string]string
}

func NewMultiKeystore(cfg *KeystoreConfig, pf PassPhraser) (*multiKeystore, error) {
	ks := keystore.NewKeyStore(
		cfg.KeyDir,
		keystore.LightScryptN,
		keystore.LightScryptP,
	)

	if cfg.StateFile == "" {
		home, err := util.GetUserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "cannot get user's home dir")
		}

		cfg.StateFile = path.Join(home, ".sonm")
	}

	return &multiKeystore{
		knownPasses: cfg.PassPhrases,
		keysDir:     cfg.KeyDir,
		stateDir:    cfg.StateFile,
		keyStore:    ks,
		passReader:  pf,
	}, nil
}

func (m *multiKeystore) List() ([]accounts.Account, error) {
	return m.keyStore.Accounts(), nil
}

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
		m.setDefaultAccount(acc.Address)
	}

	return m.readAccount(acc)
}

func (m *multiKeystore) GetKeyByAddress(addr common.Address) (*ecdsa.PrivateKey, error) {
	for _, acc := range m.keyStore.Accounts() {
		if acc.Address.Big().Cmp(addr.Big()) == 0 {
			return m.readAccount(acc)
		}
	}

	return nil, errors.New("cannot find given address into keystore")
}

func (m *multiKeystore) GetDefault() (*ecdsa.PrivateKey, error) {
	if len(m.keyStore.Accounts()) == 0 {
		return nil, errors.New("no accounts present into keystore")
	}

	defaultAddr, err := m.getDefaultAddress()
	if err != nil {
		// no keys marked as default
		return nil, errors.Wrap(err, "cannot load default key")
	}

	if !m.keyStore.HasAddress(defaultAddr) {
		return nil, errors.New("default address was read but not present into keystore")
	}

	return m.GetKeyByAddress(defaultAddr)
}

func (m *multiKeystore) SetDefault(addr common.Address) {
	m.setDefaultAccount(addr)
}

func (m *multiKeystore) Import(path string) (common.Address, error) {
	if !util.DirectoryExists(path) {
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
	pass, ok := m.knownPasses[acc.Address.Hex()]
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
	stateFile := path.Join(m.stateDir, defaultStateFileName)
	if !util.DirectoryExists(stateFile) {
		return common.Address{}, errors.New("cannot find keystore's state")
	}

	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		return common.Address{}, errors.New("cannot read state file")
	}

	if !common.IsHexAddress(string(data)) {
		return common.Address{}, errors.New("keystore's state data is malformed")
	}

	return common.HexToAddress(string(data)), nil
}

func (m *multiKeystore) setDefaultAccount(addr common.Address) {
	stateFile := path.Join(m.stateDir, defaultStateFileName)
	ioutil.WriteFile(stateFile, []byte(addr.Hex()), 0600)
}
