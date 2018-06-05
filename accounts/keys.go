package accounts

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"
	"github.com/sonm-io/core/util"
)

const (
	HomeConfigDir       = ".sonm"
	DefaultKeystorePath = "keystore"
)

var (
	errNoKeystoreDir = errors.New("keystore directory does not exist")
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
	// Bool param is true if keystore did not exist and was created.
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

	if err := util.DirectoryExists(o.keyDirPath); err != nil {
		return false, err
	}

	passPhrase, err := o.pf.GetPassPhrase()
	if err != nil {
		return false, err
	}

	idt = NewIdentity(o.keyDirPath)
	err = idt.Open(passPhrase)
	if err == nil {
		return false, nil
	} else {
		if err == ErrWalletIsEmpty {
			// pass passPhrase to createKey method to prevent double pass phrase reading
			_, err = o.createNewKey(idt, passPhrase)
			if err == nil {
				return true, nil
			}
		}

		return false, err
	}
}

func (o *defaultKeyOpener) GetKey() (*ecdsa.PrivateKey, error) {
	if o.idt == nil {
		return nil, ErrWalletNotOpen
	}

	return o.idt.GetPrivateKey()
}

func (o *defaultKeyOpener) createNewKey(idt Identity, passPhrease string) (*ecdsa.PrivateKey, error) {
	err := idt.New(passPhrease)
	if err != nil {
		return nil, err
	}

	err = idt.Open(passPhrease)
	if err != nil {
		return nil, err
	}

	return idt.GetPrivateKey()
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

// staticPassPhraser implements PassPhraser interface
// by holding already-known pass phrase received from any
// external source.
type staticPassPhraser struct {
	p string
}

func (pf *staticPassPhraser) GetPassPhrase() (string, error) { return pf.p, nil }

func NewStaticPassPhraser(p string) PassPhraser {
	return &staticPassPhraser{p: p}
}

// KeyStorager interface describe an item that must know something about
// a path to the keystore and a passphrase
type KeyStorager interface {
	// KeyStore returns path to key store
	KeyStore() string
	// PassPhrase returns passphrase for keystore
	PassPhrase() string
}

// Printer interface describe anything that can print
// something somehow on a something.
type Printer interface {
	Printf(format string, i ...interface{})
}

// silentPrinter implements Printer interface but prints nothing.
type silentPrinter struct{}

func (sp *silentPrinter) Printf(format string, i ...interface{}) {}

type fmtPrinter struct{}

func (fp *fmtPrinter) Printf(format string, i ...interface{}) {
	fmt.Printf(format, i...)
}

func NewFmtPrinter() Printer {
	return new(fmtPrinter)
}

// NewSilentPrinter returns new printer which can prints nothing
func NewSilentPrinter() Printer { return new(silentPrinter) }

// DefaultKeyOpener return KeyOpener configured for using with pre-defined pass-phrase or
// retrieve pass-phrase interactively
func DefaultKeyOpener(p Printer, keyDir, passPhrase string) (KeyOpener, error) {
	var err error
	keyDir, err = homedir.Expand(keyDir)
	if err != nil {
		return nil, err
	}

	// Use default key store dir if not specified in config.
	if keyDir == "" {
		keyDir, err = GetDefaultKeyStoreDir()
		if err != nil {
			return nil, err
		}
	}

	p.Printf("Using %s as KeyStore directory\r\n", keyDir)

	if err := util.DirectoryExists(keyDir); err != nil {
		p.Printf("KeyStore directory does not exist, try to create it...\r\n")
		err = os.MkdirAll(keyDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	// ask for pass-phrase if not specified in config
	var pf PassPhraser
	if passPhrase == "" {
		pf = NewInteractivePassPhraser()
	} else {
		pf = NewStaticPassPhraser(passPhrase)
	}

	ko := NewKeyOpener(keyDir, pf)
	return ko, nil
}

func GetDefaultKeyStoreDir() (string, error) {
	home, err := util.GetUserHomeDir()
	if err != nil {
		return "", err
	}

	keyDir := path.Join(home, HomeConfigDir, DefaultKeystorePath)
	return keyDir, nil
}

// todo: move this code onto (c *EthConfig) LoadKey(options ...Option)
func LoadKeys(keystore, passphrase string, options ...Option) (*ecdsa.PrivateKey, error) {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	ko, err := DefaultKeyOpener(opts.printer, keystore, passphrase)
	if err != nil {
		return nil, err
	}

	_, err = ko.OpenKeystore()
	if err != nil {
		return nil, err
	}

	return ko.GetKey()
}

type EthConfig struct {
	Passphrase string `required:"false" default:"" yaml:"pass_phrase"`
	Keystore   string `required:"false" default:"" yaml:"key_store"`
}

func (c *EthConfig) LoadKey(options ...Option) (*ecdsa.PrivateKey, error) {
	key, err := LoadKeys(c.Keystore, c.Passphrase, options...)
	if err != nil {
		return nil, err
	}

	return key, nil
}

type options struct {
	printer Printer
}

func newOptions() *options {
	return &options{
		printer: NewFmtPrinter(),
	}
}

type Option func(o *options)

func Silent() Option {
	return func(o *options) {
		o.printer = NewSilentPrinter()
	}
}
