package accounts

import (
	"crypto/ecdsa"
	"strings"
	"github.com/pkg/errors"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

type Identity struct {
	accman *accounts.Manager

	defaultWallet  accounts.Wallet
	defaultAccount accounts.Account

	passphrase		string

	TransactionOpts *bind.TransactOpts
	PrivateKey      *ecdsa.PrivateKey
}

func NewIdentity(datadir string, passphrase string) (idt *Identity, err error) {
	identity := Identity{}

	err = identity.initAccountManager(datadir)
	if err != nil {
		return nil, err
	}

	err = identity.setDefaultWallet(nil)
	if err != nil {
		return nil, err
	}
	identity.setDefaultAccount()

	err = identity.readPrivateKey(passphrase)
	if err != nil {
		return nil, err
	}

	identity.setTransactOpts()

	return &identity, nil
}

func (idt *Identity) initAccountManager(datadir string) (err error) {
	am, err := NewManager(&AccountManagerConfig{
		DataDir:           datadir,
		NoUSB:             true,
		UseLightweightKDF: false,
		KeyStoreDir:       datadir + "/keystore",
	})
	if err != nil {
		return err
	}
	idt.accman = am
	return err
}

func (idt *Identity) setDefaultWallet(walletString *string) (err error) {
	if walletString == nil {
		wallets := idt.accman.Wallets()
		if len(wallets) > 0{
			idt.defaultWallet = wallets[0]
		}
		return errors.New("Doesn't have any wallets in the store")
	} else {
		idt.defaultWallet, err = idt.accman.Wallet(*walletString)
		if err != nil {
			return err
		}
		return err
	}
}

func (idt *Identity) setDefaultAccount() {
	idt.defaultAccount = idt.defaultWallet.Accounts()[0]
}

func parseUrl(url string) ([]string, error) {
	parts := strings.Split(url, "://")
	if len(parts) != 2 || parts[0] == "" {
		err := errors.New("Error while parsing url keystore string")
		return nil, err
	}
	return parts, nil
}

func (idt *Identity) readPrivateKey(pass string) (err error) {
	parts, err := parseUrl(idt.defaultAccount.URL.String())
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
	idt.PrivateKey = key.PrivateKey
	return nil
}

func (idt *Identity) setTransactOpts(){
	idt.TransactionOpts = bind.NewKeyedTransactor(idt.PrivateKey)
}

