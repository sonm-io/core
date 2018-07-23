// +build linux darwin freebsd

package input

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// LineSep is the separator for windows or unix systems
const LineSep = "\n"

// rawRead reads file with raw mode (without prompting to terminal).
func (i *UI) rawRead(f *os.File) (string, error) {

	// MakeRaw put the terminal connected to the given file descriptor
	// into raw mode
	fd := int(f.Fd())
	if !terminal.IsTerminal(fd) {
		return "", fmt.Errorf("file descriptor %d is not a terminal", fd)
	}

	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer terminal.Restore(fd, oldState)

	return i.rawReadline(f)
}
