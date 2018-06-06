package password

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestInteractiveReader(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "TestInteractiveReader")
	require.NoError(t, err)
	f.Write([]byte("any\n"))
	f.Seek(0, 0)
	wBuf := &bytes.Buffer{}

	reader := NewInteractiveOSPasswordReader(f, wBuf)
	addr := common.HexToAddress("0x0000000000000000000000000000000000000001")
	pass, err := reader.ReadPassword(addr)
	require.NoError(t, err)
	require.Equal(t, pass, "any")
	f.Truncate(0)
	pass, err = reader.ReadPassword(addr)
	require.NoError(t, err)
	require.Equal(t, pass, "any")
	reader.ForgetPassword(addr)
	pass, err = reader.ReadPassword(addr)
	require.Error(t, err)
}
