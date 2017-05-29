package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	//"bufio"

	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
  "github.com/sonm-io/blockchain-api/go-build/SDT"
	"github.com/sonm-io/blockchain-api/go-build/Factory"

	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
)
	// THIS IS HACK AND SHOULD BE REWRITTEN
	const key = `{"address":"fe36b232d4839fae8751fa10768126ee17a156c1","crypto":{"cipher":"aes-128-ctr","ciphertext":"b2f1390ba44929e2144a44b5f0bdcecb06060b5ef1e9b0d222ed0cd5340e2876","cipherparams":{"iv":"a33a90fc4d7a052db58be24bbfdc21a3"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"422328336107aeb54b4a152f4fae0d5f2fbca052fc7688d9516cd998cf790021"},"mac":"08f3fa22882b932ae2926f6bf5b1df2c0795720bd993b50d652cee189c00315c"},"id":"b36be1bf-6eb4-402e-8e26-86da65ae3156","version":3}`

func main() {
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial("/home/jack/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	// Instantiate the contract and display its name
	token, err := Token.NewSDT(common.HexToAddress("0x4b15b70e9e1ac7e7f7edd3f7c81cf8ec4e784cc0"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	name, err := token.Name(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve token name: %v", err)
	}
	fmt.Println("Token name:", name)


	// Create an authorized transactor and spend 1 unicorn
	// yes, this is hack too, need to rewrite it.
	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	tx, err := token.Transfer(auth, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(1))
	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}
	// Need to do something about checking pending tx
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())

	factory, err := Factory.NewFactory(common.HexToAddress("0xdc0b27895ba9316571799c4044109c452eb1bc14"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}

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
	 w:=hubof.String()
	fmt.Println("Wallet address is:", h)


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
