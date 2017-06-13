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
	//"math/big"
	"fmt"
	//"github.com/sonm-io/blockchain-api/go-build/MinWallet"
)
func main (){



	pass:=blockchain.ReadPwd()


	key:=blockchain.ReadKey()

  owner:=common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1")


	hd:=blockchain.GHome()

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
	h, err := blockchain.CreateHub(conn,auth)
		if err != nil {
			log.Fatalf("Failed to create hub : %v", err)
		}
    fmt.Println("tx:")
    fmt.Println(h)

    fmt.Println("Wait!")
    time.Sleep(15* time.Second)

    hAddr,err:=blockchain.GetHubAddr(conn,owner)
    if err != nil {
			log.Fatalf("Failed to create hub : %v", err)
		}
    fmt.Println("hub address:")
    hAdr:=hAddr.String()
    fmt.Println(hAdr)



	// Instantiate the contract and display its name
	//create tokens
	token, err := blockchain.GlueToken(conn)
	if err != nil {
		log.Fatalf("Failed to : %v", err)
	}

  name, err := token.Name(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve token name: %v", err)
	}
	fmt.Println("Token name:", name)




	//to :=
	//am :=
	//sent tokens
  tx, err:= blockchain.TransferToken(conn,auth,hAddr,1)
  if err != nil {
		log.Fatalf("Failed to do something: %v", err)
	}
  fmt.Println("tx:",tx)




// Check for registration
/*
  check,err:= blockchain.CheckHubs(conn,hAddr)
  if err != nil {
    log.Fatalf("Failed to do something: %v", err)
  }
  fmt.Println("check:",check)
*/


}
