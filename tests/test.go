package main
import (
	"log"
	"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"strings"
	//"io/ioutil"
	//"github.com/sonm-io/go-ethereum/accounts/keystore"
	"github.com/sonm-io/go-ethereum/common"
	"time"
	"math/big"
	"fmt"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
)
func main (){



	pass:=blockchainApi.ReadPwd()


	key:=blockchainApi.ReadKey()


	hd:=blockchainApi.GHome()

	conn, err := ethclient.Dial(hd+"/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}



	auth, err := bind.NewTransactor(strings.NewReader(key), pass)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}


	//-------------------- 2 ------------------------------
	//--HUB INIT--//
	//Create Hub wallet
	h, err := blockchainApi.CreateHub(conn,auth)
		if err != nil {
			log.Fatalf("Failed to create hub : %v", err)
		}


	// Instantiate the contract and display its name
	//create tokens
	ct, err := blockchainApi.GlueToken(conn)
	if err != nil {
		log.Fatalf("Failed to : %v", err)
	}

	fmt.Println("Token name:", ct.Name(nil))
	//to :=
	//am :=
	//sent tokens
	
}
