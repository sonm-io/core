package accounts

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Identity interface uses for auth and detect all objects in network
// source implementation going to go-ethereum accounting
// its need to be storing wallets in one dir and opened it by passphrase
type Identity interface {
	// return *ecdsa.PrivateKey, it include PublicKey and ethereum Address shortly
	GetPrivateKey() (*ecdsa.PrivateKey, error)

	// open any keystore and seek account in this keystore
	// DANG: this not getting ready account for using, use Open() for setup account and GetPrivateKey() for getting this
	Load(keydir string) error

	// open loading account
	// use this after Load()
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
func NewIdentity(keydir string) (idt *identityPassphrase, err error) {
	idt = &identityPassphrase{}
	err = idt.Load(keydir)
	if err != nil {
		return nil, err
	}
	return idt, nil
}

// Load keystore
// this implementation
func (idt *identityPassphrase) Load(keystoreDir string) (err error) {
	err = idt.initKeystore(keystoreDir)
	if err != nil {
		return err
	}
	return nil
}

func (idt *identityPassphrase) Open(pass string) (err error) {
	return idt.readPrivateKey(pass)
}

func (idt *identityPassphrase) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	if idt.privateKey == nil {
		return nil, errors.New("Wallet is not open now")
	}
	return idt.privateKey, nil
}

// Keystore initialization
// init keystore at homedir while keydir params is nil
func (idt *identityPassphrase) initKeystore(keydir string) (err error) {

	fmt.Println(keydir)

	idt.keystore = keystore.NewKeyStore(keydir, keystore.LightScryptN, keystore.LightScryptP)

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
	return nil
}

// Read and decrypt Privatekey with getting passphrase
func (idt *identityPassphrase) readPrivateKey(pass string) (err error) {
	parts, err := parseKeystoreUrl(idt.defaultAccount.URL.String())
	if err != nil {
		return err
	}
	file, err := ioutil.ReadFile(parts[1])
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

func parseKeystoreUrl(url string) ([]string, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 || parts[0] == "" {
		err := errors.New("Error while parsing url keystore string")
		return nil, err
	}
	return parts, nil
}

// return default keystore directory stored in in `.sonm` directory
// if any error occurred .sonm directory will be in working dir
func GetDefaultKeystoreDir() string{
	rootDir, err := util.GetUserHomeDir()
	if err != nil {
		rootDir = ""
	}
	return filepath.Join(rootDir, ".sonm", "keystore")
}