package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	//"bufio"
	"time"

	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
  "github.com/sonm-io/blockchain-api/go-build/SDT"
	"github.com/sonm-io/blockchain-api/go-build/Factory"
	"github.com/sonm-io/blockchain-api/go-build/Whitelist"
	"github.com/sonm-io/blockchain-api/go-build/HubWallet"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"encoding/json"
	"os"
)

//-------INIT ZONE--------------------------------------------------------------

// THIS IS HACK AND SHOULD BE REWRITTEN
const key = `{"address":"fe36b232d4839fae8751fa10768126ee17a156c1","crypto":{"cipher":"aes-128-ctr","ciphertext":"b2f1390ba44929e2144a44b5f0bdcecb06060b5ef1e9b0d222ed0cd5340e2876","cipherparams":{"iv":"a33a90fc4d7a052db58be24bbfdc21a3"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"422328336107aeb54b4a152f4fae0d5f2fbca052fc7688d9516cd998cf790021"},"mac":"08f3fa22882b932ae2926f6bf5b1df2c0795720bd993b50d652cee189c00315c"},"id":"b36be1bf-6eb4-402e-8e26-86da65ae3156","version":3}`

//create json for writing password
type MessageJson struct {
	Key       string     `json:"Key"`
	}
func main() {
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial("/home/cotic/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	// Instantiate the contract and display its name
	token, err := Token.NewSDT(common.HexToAddress("0x09e4a2de83220c6f92dcfdbaa8d22fe2a4a45943"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	name, err := token.Name(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve token name: %v", err)
	}
	fmt.Println("Token name:", name)

	//write json

	file, err := os.Create("Password.json")

	if err != nil{
		// handle the error here
		log.Fatal("Failed to create Password.json")
	}
	_ = json.NewEncoder(os.Stdout).Encode(
		MessageJson{key},
	)
	defer file.Close()


	// Create an authorized transactor and spend 1 unicorn
	// yes, this is hack too, need to rewrite it.
	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	tx, err := token.Transfer(auth, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(1 * 10^17))

	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}
	// Need to do something about checking pending tx
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())

	//Define factory
	factory, err := Factory.NewFactory(common.HexToAddress("0xfadcd0e54a6bb4c8e1130b4da6022bb29540c1a1"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}

	//Put correct addresses into Factory
	// auth, dao-address, Whitelist address as @params
	tx, err = factory.ChangeAdresses(auth,common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1"),common.HexToAddress("0x833865a1379b9750c8a00b407bd6e2f08e465153"))
	if err != nil {
		log.Fatalf("Failed to request change address: %v", err)
	}
	fmt.Println("ChangeAdresses pending: 0x%x\n", tx.Hash())


	time.Sleep(250 * time.Millisecond)

	//-----------PayLoad----------------------------------------------------------

	//Creates HUB

	tx, err = factory.CreateHub(auth)
	if err != nil {
		log.Fatalf("Failed to request hub creation: %v", err)
	}
	fmt.Println("CreateHub pending: 0x%x\n", tx.Hash())

	// Don't even wait, check its presence in the local pending state
	time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

	//Request info about hubs
	hubof, err := factory.HubOf(&bind.CallOpts{Pending: true}, common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1"))
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}

	 wb:=hubof
	 w:=hubof.String()

	fmt.Println("Wallet address is:", w)


	//Registry Hub in Whitelist

	//Add some tokens
	tx, err = token.Transfer(auth, wb, big.NewInt(15 * 10^17))
	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}
	// Need to do something about checking pending tx
	fmt.Printf("Transfer to hub wallet pending: 0x%x\n", tx.Hash())

	time.Sleep(250 * time.Millisecond)

	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0x833865a1379b9750c8a00b407bd6e2f08e465153"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}

	//Define HubWallet
	hw, err := Hubwallet.NewHubWallet(wb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a HubWallet contract: %v", err)
	}

	//Register HubWallet
	tx, err = hw.Registration(auth)
	if err != nil {
		log.Fatalf("Failed to request hub registration: %v", err)
	}
	fmt.Println("Registration pending: 0x%x\n", tx.Hash())

	// Don't even wait, check its presence in the local pending state
	time.Sleep(250 * time.Millisecond)

	//Request info about hubs
	wRegistredHubs, err := whitelist.RegistredHubs(&bind.CallOpts{Pending: true},wb)
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}

	 wf:=wRegistredHubs


	fmt.Println("Wallet address is:", wf)


//----------MINER INIT----------------------------------------------------------

//Creates MINER

tx, err = factory.CreateMiner(auth)
if err != nil {
	log.Fatalf("Failed to request Miner creation: %v", err)
}
fmt.Println("createMiner pending: 0x%x\n", tx.Hash())

// Don't even wait, check its presence in the local pending state
time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

//Request info about miners from factory
minerof, err := factory.MinerOf(&bind.CallOpts{Pending: true}, common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1"))
if err != nil {
	log.Fatalf("Failed to retrieve miner wallet: %v", err)
}

 mb:=minerof
 m:=minerof.String()

fmt.Println("Miner Wallet address is:", m)
time.Sleep(250 * time.Millisecond)

//Registry Miner in Whitelist

//Adding some funds to quote
//Add some tokens
tx, err = token.Transfer(auth, mb, big.NewInt(15 * 10^17))
if err != nil {
	log.Fatalf("Failed to request token transfer: %v", err)
}
// Need to do something about checking pending tx
fmt.Printf("Transfer to MinerWallet wallet pending: 0x%x\n", tx.Hash())

time.Sleep(250 * time.Millisecond)
time.Sleep(250 * time.Millisecond)
time.Sleep(250 * time.Millisecond)

//Define MinerWallet
mw, err := MinWallet.NewMinerWallet(mb, conn)
if err != nil {
	log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
}

//Register MinerWallet
tx, err = mw.Registration(auth,big.NewInt(2))
if err != nil {
	log.Fatalf("Failed to request miner registration: %v", err)
}
fmt.Println("Registration pending: 0x%x\n", tx.Hash())

// Don't even wait, check its presence in the local pending state
time.Sleep(250 * time.Millisecond)

//Check info about miners
wRegistredMiners, err := whitelist.RegistredMiners(&bind.CallOpts{Pending: true},mb)
if err != nil {
	log.Fatalf("Failed to retrieve miner wallet: %v", err)
}

 mf:=wRegistredMiners

fmt.Println("Wallet address is:", mf)

//------Transfers---------------------------------------------------------------

//Hub is registred, we should transfer him money

//First of all - we will try to get balance of Hub

bal, err := token.BalanceOf(&bind.CallOpts{Pending: true},wb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
//bal:=tx

fmt.Printf("Balance of Hub", bal)

//Balance of Miner

bal, err = token.BalanceOf(&bind.CallOpts{Pending: true},mb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
//bal=tx

fmt.Printf("Balance of Miner", bal)

/*
//Make some initial supplyment to hubwallet

tx, err = token.Transfer(auth, wb, big.NewInt(10))
if err != nil {
	log.Fatalf("Failed to request token transfer: %v", err)
}
// Need to do something about checking pending tx
fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())
*/

//Transfer some as a payout

//Transfer funds from hub side
tx, err = hw.Transfer(auth,mb,big.NewInt(2 * 10^17))
if err != nil {
	log.Fatalf("Failed to request hub transfer: %v", err)
}
fmt.Println(" pending: 0x%x\n", tx.Hash())

// Don't even wait, check its presence in the local pending state
time.Sleep(250 * time.Millisecond)

//Pull payment from MinerSide

tx, err = mw.PullMoney(auth,wb)
if err != nil {
	log.Fatalf("Failed to request : %v", err)
}
fmt.Println(" pending: 0x%x\n", tx.Hash())

// Don't even wait, check its presence in the local pending state
time.Sleep(250 * time.Millisecond)

bal, err = token.BalanceOf(&bind.CallOpts{Pending: true},mb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
//bal=tx

fmt.Printf("Balance of Miner", bal)


/*
// Withdraw to main
tx, err = mw.Withdraw(auth)
if err != nil {
	log.Fatalf("Failed to request : %v", err)
}
fmt.Println(" pending: 0x%x\n", tx.Hash())

// Don't even wait, check its presence in the local pending state
time.Sleep(250 * time.Millisecond)
*/


//Something wrong with sessions bindings, it is a go-ethereum bug again. Probably need to fix in the future
/*
	// Wrap the Token contract instance into a session
t_session := &token.SDTSession{
	Contract: token,
	CallOpts: bind.CallOpts{
		Pending: true,
	},
	TransactOpts: bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		GasLimit: big.NewInt(3141592),
	},
}
// Call the previous methods without the option parameters


		name = t_session.Name()
		fmt.Println("Token name:", name)
		/*
		tx = t_session.Transfer("0x0000000000000000000000000000000000000000"), big.NewInt(1))
		fmt.Println("Transaction pending:", tx)
		*/
}

