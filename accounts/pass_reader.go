package accounts

import (
	"io"
	"os"

	"github.com/howeyc/gopass"
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

// interactivePassPhraser implements the PassPhrase which allows to
// read passphrase on terminal interactively
type interactivePassPhraser struct {
	reader gopass.FdReader
	writer io.Writer
}

func (pf *interactivePassPhraser) GetPassPhrase() (string, error) {
	// TODO: pass prompt text as param, it will make prompt more obvious for user.
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

type staticPassPhraser struct {
	p string
}

func (pf *staticPassPhraser) GetPassPhrase() (string, error) { return pf.p, nil }

// NewStaticPassPhraser inits pass phrase reader with pre-defined pass.
func NewStaticPassPhraser(p string) PassPhraser {
	return &staticPassPhraser{p: p}
}
