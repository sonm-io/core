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
	cli    *InteractiveCliPasswordReader
}

func NewInteractiveOSPasswordReader(reader gopass.FdReader, writer io.Writer) *InteractiveOSPasswordReader {
	return &InteractiveOSPasswordReader{
		reader: reader,
		writer: writer,
		cli:    NewInteractiveCliPasswordReader(reader, writer, false),
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
			pw, err := m.cli.ReadPassword(address)
			if err != nil {
				return "", err
			}
			keyring.Set(sonmServicePrefix+address.Hex(), user.Username, string(pw))
			return string(pw), nil
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
