package accounts

import (
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeystoreDir = "/tmp/test/"
)

func TestGenerateFirst(t *testing.T) {
	k, err := NewMultiKeystore(NewKeystoreConfig(testKeystoreDir), NewStaticPassPhraser("test"))
	require.NoError(t, err)
	defer func() { os.RemoveAll(testKeystoreDir) }()

	first, err := k.Generate()
	require.NoError(t, err)
	assert.NotNil(t, first)

	defaultKey, err := k.GetDefault()
	require.NoError(t, err)
	assert.Equal(t, defaultKey, first)

	randomKey, err := k.Generate()
	require.NoError(t, err)
	assert.NotNil(t, randomKey)

	yetAnotherDefault, err := k.GetDefault()
	require.NoError(t, err)
	assert.NotEqual(t, randomKey, yetAnotherDefault)
}

func TestSetDefault(t *testing.T) {
	k, err := NewMultiKeystore(NewKeystoreConfig(testKeystoreDir), NewStaticPassPhraser("test"))
	require.NoError(t, err)

	defer func() { os.RemoveAll(testKeystoreDir) }()

	for i := 0; i < 5; i++ {
		_, err := k.Generate()
		require.NoError(t, err)
	}

	ls := k.List()
	assert.Len(t, ls, 5)

	defaultAddr := ls[3].Address
	k.SetDefault(defaultAddr)

	priv, err := k.GetDefault()
	require.NoError(t, err)

	assert.Equal(t, crypto.PubkeyToAddress(priv.PublicKey), defaultAddr)
}

type panicPassphrase struct{}

func (panicPassphrase) GetPassPhrase() (string, error) { panic("test failed") }

func TestAlreadyKnownPasswords(t *testing.T) {
	k, err := NewMultiKeystore(NewKeystoreConfig(testKeystoreDir), NewStaticPassPhraser("test"))
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(testKeystoreDir)
	}()

	key1, err := k.Generate()
	require.NoError(t, err)
	key2, err := k.Generate()
	require.NoError(t, err)
	key3, err := k.Generate()
	require.NoError(t, err)

	addr1, addr2, addr3 := crypto.PubkeyToAddress(key1.PublicKey), crypto.PubkeyToAddress(key2.PublicKey), crypto.PubkeyToAddress(key3.PublicKey)

	anotherK, err := NewMultiKeystore(
		&KeystoreConfig{
			KeyDir: testKeystoreDir,
			PassPhrases: map[string]string{
				addr1.Hex(): "test",
				addr2.Hex(): "test",
				addr3.Hex(): "test",
			},
		},
		// panicPassphrase panics if called, should never happens.
		panicPassphrase{},
	)
	require.NoError(t, err)

	res1, err := anotherK.GetKeyByAddress(addr1)
	require.NoError(t, err)
	assert.Equal(t, res1, key1)

	res2, err := anotherK.GetKeyByAddress(addr2)
	require.NoError(t, err)
	assert.Equal(t, res2, key2)

	res3, err := anotherK.GetKeyByAddress(addr3)
	require.NoError(t, err)
	assert.Equal(t, res3, key3)
}
