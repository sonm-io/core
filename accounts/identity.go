package accounts

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"
	"io/ioutil"
	url2 "net/url"
	"path/filepath"
)

// Identity interface uses for auth and detect all objects in network
// source implementation going to go-ethereum accounting
// its need to be storing wallets in one dir and opened it by passphrase
type Identity interface {
	// return *ecdsa.PrivateKey, it include PublicKey and ethereum Address shortly
	GetPrivateKey() (*ecdsa.PrivateKey, error)


	// created new account in keystore
	// WARN: not open created account
	New(passphrase string) error

	// open loading account
	// use this after load()
	// passphrase may be blank - eg. passphrase=""
	Open(passphrase string) error
}

// Implementation of Identity interface
// working trough KeystorePassphrase from go-ethereum
type identityPassphrase struct {
	Identity

	keystore *keystore.KeyStore

	defaultWallet  accounts.Wallet
	defaultAccount accounts.Account

	passphrase string

	privateKey *ecdsa.PrivateKey
	key        *keystore.Key
}

// Create new instance of identity
// this implementation works though passphrase
func NewIdentity(keydir string) (idt *identityPassphrase) {
	idt = &identityPassphrase{}
	idt.load(keydir)
	return idt
}

func (idt *identityPassphrase) Open(passphrase string) (err error) {

	wallets := idt.keystore.Wallets()

	fmt.Println(wallets)

	if len(wallets) == 0 {
		return errors.New("Doesn't have any wallets in the store")
	}
	idt.defaultWallet = wallets[0]

	accs := idt.defaultWallet.Accounts()
	if len(accs) == 0 {
		return errors.New("Doesn't have any accounts in the wallet")
	}
	idt.defaultAccount = accs[0]

	return idt.readPrivateKey(passphrase)
}

func (idt *identityPassphrase) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	if idt.privateKey == nil {
		return nil, errors.New("Wallet is not open now")
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

// load keystore
func (idt *identityPassphrase) load(keystoreDir string) {
	idt.initKeystore(keystoreDir)
}

// Keystore initialization
// init keystore at homedir while keydir params is nil
func (idt *identityPassphrase) initKeystore(keydir string) {
	idt.keystore = keystore.NewKeyStore(keydir, keystore.LightScryptN, keystore.LightScryptP)
}

// Read and decrypt Privatekey with getting passphrase
func (idt *identityPassphrase) readPrivateKey(pass string) (err error) {
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
func parseKeystoreUrl(url string) (string, error) {
	fmt.Println(url)
	u, err := url2.Parse(url)
	if err != nil {
		return "", err
	}
	return u.Path, nil
}

// return default keystore directory stored in in `.sonm` directory
// if any error occurred .sonm directory will be in working dir
func GetDefaultKeystoreDir() string {
	rootDir, err := util.GetUserHomeDir()
	if err != nil {
		rootDir = ""
	}
	return filepath.Join(rootDir, ".sonm", "keystore")
}
