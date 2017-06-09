package blockchainApi

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
  	"github.com/sonm-io/blockchain-api/go-build/SDT"
	"github.com/sonm-io/blockchain-api/go-build/Factory"
	"github.com/sonm-io/blockchain-api/go-build/Whitelist"
	"github.com/sonm-io/blockchain-api/go-build/HubWallet"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"

	"github.com/sonm-io/go-ethereum/accounts/abi"
	"encoding/json"
	"os"
	"io/ioutil"
	"os/user"
	"github.com/sonm-io/go-ethereum/core/types"
)

//----ServicesSupporters Allocation---------------------------------------------


//For rinkeby testnet
const confFile = ".rinkeby/keystore/key.json"

//create json for writing KEY
type MessageJson struct {
	Key       string     `json:"Key"`
	}


//Reading KEY
func readKey() MessageJson{
	usr, err := user.Current();
	file, err := ioutil.ReadFile(usr.HomeDir+"/"+confFile)
	if err != nil {
		fmt.Println(err)
	}

	var m MessageJson
	err = json.Unmarshal(file, &m)
	if err != nil {
		fmt.Println(err)
	}
	return m
}

type PasswordJson struct {
	Password		string	`json:"Password"`
}

//Reading user password
// ВОПРОС - Это возвращает JSON структуру или строку?
func readPwd() PasswordJson{
	usr, err := user.Current();
	// User password file JSON should be in root of home directory
	file, err := ioutil.ReadFile(usr.HomeDir+"/")
	if err != nil {
		fmt.Println(err)
	}

	var m PasswordJson
	err = json.Unmarshal(file, &m)
	if err != nil {
		fmt.Println(err)
	}
	return m
}


//--Services Getters-----------------------------
/*
	Those functions allows someone behind library gets
	conn and auth for further interaction

*/

//Establish Connection to geth IPC
// Create an IPC based RPC connection to a remote node
func cnct() {
	// NOTE there is should be wildcard but not username.
	// Try ~/.rinkevy/geth.ipc
conn, err := ethclient.Dial("/home/cotic/.rinkeby/geth.ipc")
if err != nil {
	log.Fatalf("Failed to connect to the Ethereum client: %v", err)
}
	//return connectiion obj
  return conn
}

// Create an authorized transactor
func getAuth() {

	key:=readKey()
	pass:=readPwd()

auth, err := bind.NewTransactor(strings.NewReader(key), pass)
if err != nil {
	log.Fatalf("Failed to create authorized transactor: %v", err)
}
	return auth
}

//-----------------------------------------------------------


//---Defines Binds-----------------------------------------

/*
	Those should be internal functions for internal usage (but not for sure)

*/


//Token Defines
func GlueToken(conn *ethclient.Client)  {
	// Instantiate the contract
	token, err := Token.NewSDT(common.HexToAddress("0x09e4a2de83220c6f92dcfdbaa8d22fe2a4a45943"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}

}

func GlueFactory(conn *ethclient.Client)  {
	//Define factory
	factory, err := Factory.NewFactory(common.HexToAddress("0xfadcd0e54a6bb4c8e1130b4da6022bb29540c1a1"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}
}

func GlueWhitelist(conn *ethclient.Client)  {
	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0x833865a1379b9750c8a00b407bd6e2f08e465153"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}
}

func GlueHubWallet(conn *ethclient.Client, wb common.Address)  {
	//Define HubWallet
	hw, err := Hubwallet.NewHubWallet(wb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a HubWallet contract: %v", err)
	}
}

func GlueMinWallet(conn *ethclient.Client, wb common.Address)  {
	//Define MinerWallet
	mw, err := MinWallet.NewMinerWallet(mb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
	}
	}
//-------------------------------------------------------------------------

//--MAIN LIBRARY-----------------------------------------------------------

/*
Example
*/









}
