package Fusrodah

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
	"os/user"
)

type Key struct {
	Prv *ecdsa.PrivateKey
}

func (key *Key) getKeyfilePath() string {
	usr, _ := user.Current()
	keyFolder := usr.HomeDir + "/" + ".sonm/"
	os.Mkdir(keyFolder, 0755)
	keyFile := keyFolder + "hub"
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		key.createKeyFile()
	}
	return keyFile
}

func (*Key) createKeyFile() {
	usr, _ := user.Current()
	keyFolder := usr.HomeDir + "/" + ".sonm/"
	os.Mkdir(keyFolder, 0755)
	keyFile := keyFolder + "hub"
	os.Create(keyFile)
}

func (key *Key) Generate() {
	key.Prv, _ = crypto.GenerateKey()
}

func (key *Key) Load() bool {
	keyFile := key.getKeyfilePath()

	prv, err := crypto.LoadECDSA(keyFile)
	if err != nil {
		fmt.Println(err)
		return false
	}

	key.Prv = prv
	return true
}

func (key *Key) Save() bool {

	keyFile := key.getKeyfilePath()
	err := crypto.SaveECDSA(keyFile, key.Prv)

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func (key *Key) GetPubKey() *ecdsa.PublicKey {
	pkBytes := crypto.FromECDSA(key.Prv)
	pk := crypto.ToECDSAPub(pkBytes)
	return pk
}

func (key *Key) GetPubKeyString() string {
	pkString := string(crypto.FromECDSAPub(&key.Prv.PublicKey))
	return pkString
}
