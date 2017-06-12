package main
import (
	"log"
	"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"strings"
	"io/ioutil"
	"github.com/sonm-io/go-ethereum/accounts/keystore"
)
func main (){
	//-------------------- 1 ------------------------------
	keyin := strings.NewReader(key)
	passphrase := "cotic"

	json, err := ioutil.ReadAll(keyin)
		if err != nil {
			return nil, err
		}
	key, err := keystore.DecryptKey(json, passphrase)
		if err != nil {
			return nil, err
		}

	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

	//Create connection
	conn, err := ethclient.Dial("/home/cotic/.rinkeby/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

	//-------------------- 2 ------------------------------

	//Create Hub
	h, err := blockchainApi.CreateHub(conn)  //(conn)
		if err != nil {
			log.Fatalf("Failed to create hub : %v", err)
		}

	//create hub wallet
	hubWallet, err := blockchainApi.GlueHubWallet(conn, h)
		if err != nil {
			log.Fatalf("Failed to create hub wallet: %v", err)
		}

	//-------------------- 3 ------------------------------

}

