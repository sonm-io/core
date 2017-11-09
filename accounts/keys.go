package accounts

import (
	"crypto/ecdsa"
	"errors"
	"io"
	"os"

	"github.com/howeyc/gopass"
)

var (
	errNoKeystoreDir = errors.New("keystore directory does not exists")
)

// PassPhraser is interface for retrieving
// pass phrase for Eth keys
//
// If you want to retrieve pass phrases reader different ways
// (e.g: from file, from env variables, interactively from terminal)
// you must implement PassPhraser reader a different way and pass it to
// KeyOpener instance
type PassPhraser interface {
	GetPassPhrase() (string, error)
}

// KeyOpener is interface for loading Eth keys
type KeyOpener interface {
	// GetPassPhraser return PassPhraser interface
	// that provides pass phrase for loaded keys
	GetPassPhraser() PassPhraser
	// OpenKeystore opens key storage.
	// bool param is true if keystore was not existed and was created
	OpenKeystore() (bool, error)
	// GetKey returns private key from opened storage
	GetKey() (*ecdsa.PrivateKey, error)
}

// defaultKeyOpener implements KeyOpener interface
type defaultKeyOpener struct {
	keyDirPath string
	idt        Identity
	pf         PassPhraser
}

func (o *defaultKeyOpener) GetPassPhraser() PassPhraser {
	return o.pf
}

func (o *defaultKeyOpener) OpenKeystore() (bool, error) {
	var err error
	var idt Identity

	defer func() {
		if err == nil {
			o.idt = idt
		}
	}()

	if !hasDir(o.keyDirPath) {
		return false, errNoKeystoreDir
	}

	passPhrase, err := o.pf.GetPassPhrase()
	if err != nil {
		return false, err
	}

	idt = NewIdentity(o.keyDirPath)
	err = idt.Open(passPhrase)
	if err == nil {
		return false, nil
	}

	if err == ErrWalletIsEmpty {
		_, err = o.createNewKey(idt)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (o *defaultKeyOpener) GetKey() (*ecdsa.PrivateKey, error) {
	if o.idt == nil {
		return nil, ErrWalletNotOpen
	}

	return o.idt.GetPrivateKey()
}

func (o *defaultKeyOpener) createNewKey(idt Identity) (*ecdsa.PrivateKey, error) {
	passPhrease, err := o.pf.GetPassPhrase()
	if err != nil {
		return nil, err
	}

	err = idt.New(passPhrease)
	if err != nil {
		return nil, err
	}

	err = idt.Open(passPhrease)
	if err != nil {
		return nil, err
	}

	return idt.GetPrivateKey()
}

func hasDir(p string) bool {
	if _, err := os.Stat(p); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// NewKeyOpener returns KeyOpener that able to open keys
func NewKeyOpener(keyDir string, pf PassPhraser) KeyOpener {
	ko := &defaultKeyOpener{
		keyDirPath: keyDir,
		idt:        nil,
		pf:         pf,
	}

	return ko
}

// interactivePassPhraser implements the PassPhrase which allows to
// read passphrase on terminal interactively
type interactivePassPhraser struct {
	reader gopass.FdReader
	writer io.Writer
}

func (pf *interactivePassPhraser) GetPassPhrase() (string, error) {
	pw, err := gopass.GetPasswdPrompt("\r\nKey passphrase: ", false, pf.reader, pf.writer)
	if err != nil {
		return "", err
	}

	return string(pw), nil
}

// NewInteractivePassPhraser implements PassPhraser that prompts user for pass-phrase
// and read it from terminal's Stdin
func NewInteractivePassPhraser() PassPhraser {
	return &interactivePassPhraser{
		reader: os.Stdin,
		writer: os.Stdout,
	}
}
