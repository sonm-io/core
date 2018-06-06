package password

import (
	"io"
	"os/user"

	"github.com/ethereum/go-ethereum/common"
	"github.com/howeyc/gopass"
	"github.com/zalando/go-keyring"
)

const (
	sonmServicePrefix = "sonmKeychain_"
)

type InteractiveOSPasswordReader struct {
	reader gopass.FdReader
	writer io.Writer
}

func NewInteractiveOSPasswordReader(reader gopass.FdReader, writer io.Writer) *InteractiveOSPasswordReader {
	return &InteractiveOSPasswordReader{
		reader: reader,
		writer: writer,
	}
}

func (m *InteractiveOSPasswordReader) ReadPassword(address common.Address) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	secret, err := keyring.Get(sonmServicePrefix+address.Hex(), user.Username)
	if err != nil {
		if err == keyring.ErrNotFound {
			return m.readFromCli(address, user)
		}
		return "", err
	}
	return secret, nil
}

func (m *InteractiveOSPasswordReader) ForgetPassword(address common.Address) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	return keyring.Delete(sonmServicePrefix+address.Hex(), user.Username)
}

func (m *InteractiveOSPasswordReader) readFromCli(address common.Address, user *user.User) (string, error) {
	pw, err := gopass.GetPasswdPrompt("\r\nKey passphrase: ", false, m.reader, m.writer)
	if err != nil {
		return "", err
	}

	keyring.Set(sonmServicePrefix+address.Hex(), user.Username, string(pw))

	return string(pw), nil
}
