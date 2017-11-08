package accounts

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	testKeyStorePath = "/tmp/sonm-test-keystore/"
)

func initTestKeyStore() error {
	return os.Mkdir(testKeyStorePath, 0755)
}

func flushTestKeyStore() {
	os.RemoveAll(testKeyStorePath)
}

func TestNewKeyOpener_Open(t *testing.T) {
	err := initTestKeyStore()
	assert.NoError(t, err)
	defer flushTestKeyStore()

	K := NewKeyOpener(testKeyStorePath, new(nullPassPhraser))

	created, err := K.OpenKeystore()
	assert.NoError(t, err)
	assert.True(t, created, "Must create new Store in empty dir")
}

func TestNewKeyOpener_NoDir(t *testing.T) {
	flushTestKeyStore()
	K := NewKeyOpener(testKeyStorePath, new(nullPassPhraser))

	_, err := K.OpenKeystore()
	assert.EqualError(t, err, errNoKeystoreDir.Error(), "Must not be opened with non-existent dir")
}

func TestNewKeyOpener_GetKey(t *testing.T) {
	err := initTestKeyStore()
	assert.NoError(t, err)
	defer flushTestKeyStore()

	K := NewKeyOpener(testKeyStorePath, new(nullPassPhraser))

	created, err := K.OpenKeystore()
	assert.NoError(t, err)
	assert.True(t, created)

	key, err := K.GetKey()
	assert.NoError(t, err)
	assert.NotNil(t, key)
}
