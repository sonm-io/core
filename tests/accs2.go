package main
import (
	"log"
	//"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"strings"
	"io/ioutil"
	"github.com/sonm-io/go-ethereum/accounts/keystore"
	//"github.com/sonm-io/go-ethereum/common"
	//"time"
	//"math/big"
	"fmt"
	//"github.com/sonm-io/blockchain-api/go-build/MinWallet"
)
func main (){
	//-------------------- 1 ------------------------------
	keyin := strings.NewReader(key)
	passphrase := "Metamorph9"

	json, err := ioutil.ReadAll(keyin)
		if err != nil {
			//return nil, err
			log.Fatalf("failed json", err)
		}
	key, err := keystore.DecryptKey(json, passphrase)
		if err != nil {
			//return nil, err
			log.Fatalf("failed key", err)
		}

	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

	//Create connection
	conn, err := ethclient.Dial("/home/jack/.rinkeby/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}


}
