package accounts

import (
	"crypto/ecdsa"
	"io/ioutil"
	"net/url"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
)

var (
	ErrWalletNoAccount = errors.New("Wallet does not have any account")
	ErrWalletNotOpen   = errors.New("Wallet is not open")
	ErrWalletIsEmpty   = errors.New("Keystore does not have any wallets")
)

func init() {
	var _ Identity = &identityPassphrase{}
}

// Identity interface uses for auth and detect all objects in network
// source implementation going to go-ethereum accounting
// its need to be storing wallets in one dir and opened it by passphrase
type Identity interface {
	// GetPrivateKey return *ecdsa.PrivateKey, it include PublicKey and ethereum Address shortly
	GetPrivateKey() (*ecdsa.PrivateKey, error)

	// New creates new account in keystore.
	// Do not open created account.
	New(passphrase string) error

	// Open opens loaded account
	Open(passphrase string) error

	// Import imports existing account from given json and pass-phrase
	Import(json []byte, passphrase string) error

	// ImportECDSA imports existing account from given private key and pass-phrase
	ImportECDSA(key *ecdsa.PrivateKey, passphrase string) error
}

// identityPassphrase implements Identity interface
// and allows to use pass-phrase to decrypt keys
type identityPassphrase struct {
	keystore *keystore.KeyStore

	defaultWallet  accounts.Wallet
	defaultAccount accounts.Account

	passphrase string

	privateKey *ecdsa.PrivateKey
	key        *keystore.Key
}

// NewIdentity creates new identity instance
// which operates given key storage dir
func NewIdentity(keydir string) Identity {
	idt := &identityPassphrase{}
	idt.initKeystore(keydir)
	return idt
}

func (idt *identityPassphrase) Open(passphrase string) error {
	wallets := idt.keystore.Wallets()

	if len(wallets) == 0 {
		return ErrWalletIsEmpty
	}
	idt.defaultWallet = wallets[0]

	accs := idt.defaultWallet.Accounts()
	if len(accs) == 0 {
		return ErrWalletNoAccount
	}
	idt.defaultAccount = accs[0]

	return idt.readPrivateKey(passphrase)
}

func (idt *identityPassphrase) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	if idt.privateKey == nil {
		return nil, ErrWalletNotOpen
	}
	return idt.privateKey, nil
}

func (idt *identityPassphrase) New(passphrase string) error {
	acc, err := idt.keystore.NewAccount(passphrase)
	if err != nil {
		return err
	}
	idt.defaultAccount = acc
	return err
}

func (idt *identityPassphrase) Import(keyJson []byte, passphrase string) error {
	_, err := idt.keystore.Import(keyJson, passphrase, passphrase)
	if err != nil {
		return err
	}
	return nil
}

func (idt *identityPassphrase) ImportECDSA(key *ecdsa.PrivateKey, passphrase string) error {
	_, err := idt.keystore.ImportECDSA(key, passphrase)
	if err != nil {
		return err
	}
	return nil
}

// initKeystore inits keystore in given directory
func (idt *identityPassphrase) initKeystore(keydir string) {
	idt.keystore = keystore.NewKeyStore(keydir, keystore.LightScryptN, keystore.LightScryptP)
}

// Read and decrypt Privatekey with getting passphrase
func (idt *identityPassphrase) readPrivateKey(pass string) error {
	path, err := parseKeystoreUrl(idt.defaultAccount.URL.String())
	if err != nil {
		return err
	}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	key, err := keystore.DecryptKey(file, pass)
	if err != nil {
		return err
	}
	idt.key = key
	idt.privateKey = key.PrivateKey
	return nil
}

// parsing key identity file and return path
// inbound param url looks like
// keystore:///users/user/home/.sonm/keystore/keyfile
// its return path - /users/user/home/.sonm/keystore/keyfile
func parseKeystoreUrl(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}
