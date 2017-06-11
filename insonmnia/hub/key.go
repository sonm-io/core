package hub

import (
	"crypto/ecdsa"
	"os/user"
	"fmt"
	"github.com/sonm-io/go-ethereum/crypto"
	"os"
)

type Key struct {
	prv 	*ecdsa.PrivateKey
}

func (key *Key) getKeyfilePath() string{
	usr, _ := user.Current()
	keyFolder := usr.HomeDir+"/"+".sonm/"
	os.Mkdir(keyFolder, 0755)
	keyFile := keyFolder+"hub"
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		key.createKeyFile()
	}
	return keyFile
}

func (*Key) createKeyFile() {
	usr, _ := user.Current()
	keyFolder := usr.HomeDir+"/"+".sonm/"
	os.Mkdir(keyFolder, 0755)
	keyFile := keyFolder+"hub"
	os.Create(keyFile)
}

func (key *Key) Generate(){
	key.prv, _ = crypto.GenerateKey()
}

func (key *Key) Load() bool{
	keyFile := key.getKeyfilePath()

	prv, err :=crypto.LoadECDSA(keyFile)
	if err != nil {
		fmt.Println(err)
		return false
	}

	key.prv = prv
	return true
}

func (key *Key) Save() bool{

	keyFile := key.getKeyfilePath()
	err := crypto.SaveECDSA(keyFile, key.prv)

	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}