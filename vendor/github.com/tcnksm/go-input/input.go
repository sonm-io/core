/*
Package input reads user input at the console. http://github.com/tcnksm/go-input

  ui := &input.UI{
      Writer: os.Stdout,
      Reader: os.Stdin,
  }

  query := "What is your name?"
  name, err := ui.Ask(query, &input.Options{
      Default: "tcnksm",
      Required: true,
      Loop:     true,
  })
*/
package input

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	// defaultWriter and defaultReader is default val for UI.Writer
	// and UI.Reader.
	defaultWriter = os.Stdout
	defaultReader = os.Stdin

	// defualtMaskVal is default mask val for read
	defaultMaskVal = "*"
)

var (
	// Errs are error returned by input functions.
	// It's useful for handling error from outside of input functions.
	ErrEmpty       = errors.New("default value is not provided but input is empty")
	ErrNotNumber   = errors.New("input must be number")
	ErrOutOfRange  = errors.New("input is out of range")
	ErrInterrupted = errors.New("interrupted")
)

// UI is user-interface of input and output.
type UI struct {
	// Writer is where output is written. For example a query
	// to the user will be written here. By default, it's os.Stdout.
	Writer io.Writer

	// Reader is source of input. By default, it's os.Stdin.
	Reader io.Reader

	// mask is option for read function
	mask    bool
	maskVal string

	bReader *bufio.Reader

	once sync.Once
}

// DefaultUI returns default UI. It outputs to stdout and intputs from stdin.
func DefaultUI() *UI {
	return &UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
}

// setDefault sets the default value for UI struct.
func (i *UI) setDefault() {
	// Set the default writer & reader if not provided
	if i.Writer == nil {
		i.Writer = defaultWriter
	}

	if i.Reader == nil {
		i.Reader = defaultReader
	}

	if i.bReader == nil {
		i.bReader = bufio.NewReader(i.Reader)
	}
}

// ValidateFunc is function to validate the user input.
//
// The following example shows validating the user input is
// 'Y' or 'n' when asking yes or no question.
type ValidateFunc func(string) error

// Options is structure contains option for input functions.
type Options struct {
	// Default is the default value which is used when no thing
	// is input.
	Default string

	// Loop loops asking user to input until getting valid input.
	Loop bool

	// Required returns error when input is empty.
	Required bool

	// HideDefault hides default var output.
	HideDefault bool

	// HideOrder hides order comment ('Enter a value')
	HideOrder bool

	// Hide hides user input is prompting console.
	Hide bool

	// Mask hides user input and will be matched by MaskVal
	// on the screen. By default, MaskVal is asterisk(*).
	Mask bool

	// MaskDefault hides default value. By default, MaskVal is asterisk(*).
	MaskDefault bool

	// MaskVal is a value which is used for masking user input.
	// By default, MaskVal is asterisk(*).
	MaskVal string

	// ValidateFunc is function to do extra validation of user
	// input string. By default, it does nothing (just returns nil).
	ValidateFunc ValidateFunc
}

// validateFunc returns ValidateFunc. If it's specified by
// user it returns it. If not returns default function.
func (o *Options) validateFunc() ValidateFunc {
	if o.ValidateFunc == nil {
		return defaultValidateFunc
	}

	return o.ValidateFunc
}

// defaultValidateFunc is default ValidateFunc which does
// nothing.
func defaultValidateFunc(input string) error {
	return nil
}

// readOpts returns readOptions from given the Options.
func (o *Options) readOpts() *readOptions {
	var mask bool
	var maskVal string

	// Hide input and prompt nothing on screen.
	if o.Hide {
		mask = true
	}

	// Mask input and prompt default maskVal.
	if o.Mask {
		mask = true
		maskVal = defaultMaskVal
	}

	// Mask input and prompt custom maskVal.
	if o.MaskVal != "" {
		maskVal = o.MaskVal
	}

	return &readOptions{
		mask:    mask,
		maskVal: maskVal,
	}
}

// maskString is used to mask string which should not be displayed.
func maskString(s string) string {
	if len(s) < 3 {
		return "*******"
	}

	return s[:3] + "****"
}
