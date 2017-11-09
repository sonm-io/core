package accounts

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var (
	testingDirectory         = "./.sonm"
	testingKeystoreDirectory = filepath.Join(testingDirectory, "keystore")
	testingKeyJson           = `{"address":"7ec8da94172d848ede642b6fdc62a6124a750f44","crypto":{"cipher":"aes-128-ctr","ciphertext":"06edce6914e6bada8a199f283b522b7e13b1ad49b665ef6eb9d9a94064c518a1","cipherparams":{"iv":"83abecb7314c0c2c37c2e1df6c3e6091"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":4096,"p":6,"r":8,"salt":"818c0ed85a2f7c3063e597fef582fe6d33f1f23f6c93c11acdfab528701e2099"},"mac":"8b63777dae08f61c9bf74e49cc3991835cebc3844021aada1e0db92cefe30935"},"id":"3dfacb22-aaea-4545-8134-6e697f54e971","version":3}`
	testingKeyPass           = ""
	testingKeyWrongPass      = "wrongPassword"
	testingKeyPrvString      = "32824833560431749722629613392302869613102870753837834504492459506999937601650"
)

func cleanTestingDirectory() {
	d, err := os.Open(testingKeystoreDirectory)
	if err != nil {
		return
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(testingKeystoreDirectory, name))
		if err != nil {
			return
		}
	}
	os.Remove(testingKeystoreDirectory)
	os.Remove(testingDirectory)
}

func initTestingKeystore() (Identity, error) {
	idt := NewIdentity(testingKeystoreDirectory)

	err := idt.Import([]byte(testingKeyJson), testingKeyPass)
	if err != nil {
		return idt, err
	}
	idt = NewIdentity(testingKeystoreDirectory)
	return idt, err
}

func TestNewIdentity(t *testing.T) {
	/* Success case - check object not null */
	idt := NewIdentity(testingKeystoreDirectory)
	idt_pass := idt.(*identityPassphrase)
	assert.NotNil(t, idt_pass.keystore)
}

func TestIdentityPassphrase_GetPrivateKey(t *testing.T) {
	/* init */
	idt, err := initTestingKeystore()
	defer cleanTestingDirectory()
	assert.NoError(t, err)

	// Success case - getting key
	idt.Open(testingKeyPass)
	prv, err := idt.GetPrivateKey()
	assert.NoError(t, err)
	assert.NotNil(t, prv)
}

func TestIdentityPassphrase_Open(t *testing.T) {
	/* init */
	cleanTestingDirectory()
	idt, err := initTestingKeystore()
	defer cleanTestingDirectory()
	assert.NoError(t, err)

	/* Success case - open existing account */
	idt = NewIdentity(testingKeystoreDirectory)
	err = idt.Open(testingKeyPass)
	assert.NoError(t, err)

	idtPrv, err := idt.GetPrivateKey()
	assert.NoError(t, err)

	assert.NotNil(t, idtPrv)
	assert.Equal(t, testingKeyPrvString, idtPrv.D.String())
}

func TestIdentityPassphrase_Open_WrongPassword(t *testing.T) {
	/* init */
	defer cleanTestingDirectory()
	idt, err := initTestingKeystore()
	assert.NoError(t, err)

	/* Wrong case - wrong password */
	err = idt.Open(testingKeyWrongPass)
	assert.Error(t, err)
}

func TestIdentityPassphrase_Open_WrongWallets(t *testing.T) {
	/* init */
	cleanTestingDirectory()
	idt := NewIdentity(testingDirectory)

	/* Wrong case - keystore doesn't exist any wallets */
	idt = NewIdentity(testingKeystoreDirectory)
	err := idt.Open(testingKeyPass)
	assert.Error(t, err)
}

func TestIdentityPassphrase_New(t *testing.T) {
	/* init */
	cleanTestingDirectory()
	defer cleanTestingDirectory()

	/* Success case - creating new Account in keystore */
	idt := NewIdentity(testingKeystoreDirectory)
	err := idt.New(testingKeyPass)
	assert.NoError(t, err)

	err = idt.Open(testingKeyPass)
	assert.NoError(t, err)

	idtPrv, err := idt.GetPrivateKey()
	assert.NoError(t, err)

	assert.NotNil(t, idtPrv)
}

func TestIdentityPassphrase_Import(t *testing.T) {
	/* init */
	cleanTestingDirectory()
	defer cleanTestingDirectory()

	idt := NewIdentity(testingKeystoreDirectory)
	err := idt.Import([]byte(testingKeyJson), testingKeyPass)
	assert.NoError(t, err)
}

func TestIdentityPassphrase_ImportECDSA(t *testing.T) {
	/* init */
	cleanTestingDirectory()
	defer cleanTestingDirectory()
	prv, _ := crypto.GenerateKey()

	/* Success case - simple import */
	idt := NewIdentity(testingKeystoreDirectory)
	err := idt.ImportECDSA(prv, testingKeyPass)
	assert.NoError(t, err)

	err = idt.Open(testingKeyPass)
	assert.NoError(t, err)

	idtPrv, err := idt.GetPrivateKey()
	assert.NoError(t, err)
	assert.Equal(t, crypto.FromECDSA(prv), crypto.FromECDSA(idtPrv))

	// load imported key
	idt = NewIdentity(testingKeystoreDirectory)
	err = idt.Open(testingKeyPass)
	assert.NoError(t, err)

	idtPrv, err = idt.GetPrivateKey()
	assert.NoError(t, err)

	assert.Equal(t, prv, idtPrv)

	/* Wrong case - importing already imported key */
	err = idt.ImportECDSA(prv, testingKeyPass)
	assert.Error(t, err)
}
