package password

import (
	"io"
	"os/user"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/howeyc/gopass"
	"github.com/zalando/go-keyring"
)

type InteractiveCliPasswordReader struct {
	reader      gopass.FdReader
	writer      io.Writer
	shouldCache bool
	cache       map[string]string
	mu          sync.Mutex
}

func NewInteractiveCliPasswordReader(reader gopass.FdReader, writer io.Writer, shouldCache bool) *InteractiveCliPasswordReader {
	return &InteractiveCliPasswordReader{
		reader:      reader,
		writer:      writer,
		shouldCache: shouldCache,
		cache:       map[string]string{},
	}
}

func (m *InteractiveCliPasswordReader) ReadPassword(address common.Address) (string, error) {
	addrHex := address.Hex()
	m.mu.Lock()
	defer m.mu.Unlock()

	if cached, ok := m.cache[addrHex]; ok {
		return cached, nil
	}

	pw, err := gopass.GetPasswdPrompt("\r\nKey passphrase: ", false, m.reader, m.writer)
	if err != nil {
		return "", err
	}

	if m.shouldCache {
		m.cache[addrHex] = string(pw)
	}

	return string(pw), nil
}

func (m *InteractiveCliPasswordReader) ForgetPassword(address common.Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, err := user.Current()
	if err != nil {
		return err
	}
	return keyring.Delete(sonmServicePrefix+address.Hex(), user.Username)
}
