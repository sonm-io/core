package accounts

import (
	"crypto/ecdsa"
	"io/ioutil"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

type Keystore struct {
	keyStore   *keystore.KeyStore
	passReader PasswordReader
	selector   DefaultSelector
}

type PasswordReader interface {
	ReadPassword(address common.Address) (string, error)
}

type DefaultSelector interface {
	HasDefault(keystore Keystore) bool
	GetDefault(keystore Keystore) (common.Address, error)
	SetDefault(keystore Keystore, address common.Address) error
}

func NewKeystore(path string, passReader PasswordReader, selector DefaultSelector) (*Keystore, error) {
	ks := keystore.NewKeyStore(
		path,
		keystore.LightScryptN,
		keystore.LightScryptP,
	)

	m := &Keystore{
		keyStore:   ks,
		passReader: passReader,
	}

	// We have only one account, mark it as default
	// for current keystore instance.
	accs := ks.Accounts()
	if len(accs) == 1 {
		if err := m.selector.SetDefault(accs[0].Address); err != nil {
			return nil, err
		}

	}

	return m, nil
}

// List returns list of accounts addresses into keystore
func (m *Keystore) List() []accounts.Account {
	return m.keyStore.Accounts()
}

// Generate creates new key into keystore
func (m *Keystore) Generate() (accounts.Account, error) {
	key, err := crypto.GenerateKey()
	if err != nil {
		return accounts.Account{}, err
	}
	pass, err := m.passReader.ReadPassword(crypto.PubkeyToAddress(key.PublicKey))
	if err != nil {
		return accounts.Account{}, errors.Wrap(err, "cannot read pass phrase")
	}
	return m.keyStore.ImportECDSA(key, pass)
}

// GetKeyByAddress loads and decrypts key form keystore (if present)
func (m *Keystore) GetKeyByAddress(addr common.Address) (*ecdsa.PrivateKey, error) {
	acc, err := m.findAccountByAddr(addr)
	if err != nil {
		return nil, err
	}

	return m.readAccount(acc)
}

// GetDefault returns default key for the keystore
func (m *Keystore) GetDefault() (*ecdsa.PrivateKey, error) {
	if len(m.keyStore.Accounts()) == 0 {
		return nil, errors.New("no accounts present into keystore")
	}

	defaultAddr, err := m.selector.GetDefault()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load default key")
	}

	return m.GetKeyByAddress(defaultAddr)
}

func (m *Keystore) findAccountByAddr(addr common.Address) (accounts.Account, error) {
	for _, acc := range m.keyStore.Accounts() {
		if acc.Address.Big().Cmp(addr.Big()) == 0 {
			return acc, nil
		}
	}

	return accounts.Account{}, errors.New("cannot find given address into keystore")
}

func (m *Keystore) readAccount(acc accounts.Account) (*ecdsa.PrivateKey, error) {
	pass, err := m.passReader.ReadPassword(acc.Address)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read pass phrase")
	}
	m.keyStore.Wallets()

	return m.readKeyFile(acc.URL.Path, pass)
}

func (m *Keystore) readKeyFile(path string, pass string) (*ecdsa.PrivateKey, error) {
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
