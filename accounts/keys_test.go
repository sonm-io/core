package accounts

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testKeyStorePath = "/tmp/sonm-test-keystore/"
)

func initTestKeyStore() error {
	return os.Mkdir(testKeyStorePath, 0755)
}

func initPassPhraser(t *testing.T) PassPhraser {
	passPhraser := NewMockPassPhraser(gomock.NewController(t))
	passPhraser.EXPECT().GetPassPhrase().AnyTimes().Return("testme", nil)

	return passPhraser
}

func flushTestKeyStore() {
	os.RemoveAll(testKeyStorePath)
}

func TestNewKeyOpener_Open(t *testing.T) {
	err := initTestKeyStore()
	assert.NoError(t, err)
	defer flushTestKeyStore()

	pf := initPassPhraser(t)
	K := NewKeyOpener(testKeyStorePath, pf)

	created, err := K.OpenKeystore()
	assert.NoError(t, err)
	assert.True(t, created, "Must create new Store reader empty dir")
}

func TestNewKeyOpener_NoDir(t *testing.T) {
	flushTestKeyStore()
	pf := initPassPhraser(t)
	K := NewKeyOpener(testKeyStorePath, pf)

	_, err := K.OpenKeystore()
	assert.EqualError(t, err, errNoKeystoreDir.Error(), "Must not be opened with non-existent dir")
}

func TestNewKeyOpener_GetKey(t *testing.T) {
	err := initTestKeyStore()
	assert.NoError(t, err)
	defer flushTestKeyStore()

	pf := initPassPhraser(t)
	K := NewKeyOpener(testKeyStorePath, pf)

	created, err := K.OpenKeystore()
	assert.NoError(t, err)
	assert.True(t, created)

	key, err := K.GetKey()
	assert.NoError(t, err)
	assert.NotNil(t, key)
}

func TestNewInteractivePassPhraser(t *testing.T) {
	r, w, err := os.Pipe()
	assert.NoError(t, err, "Cannot init os.Pipe")

	pf := interactivePassPhraser{
		reader: r,
		writer: w,
	}

	w.Write([]byte("testme"))
	pass, err := pf.GetPassPhrase()
	assert.NoError(t, err, "Cannot read password")

	assert.Equal(t, "testme", pass)
}
