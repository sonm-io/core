package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
  "github.com/sonm-io/blockchain-api/go-build/SDT"

	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
)

	const key = `paste the contents of your *testnet* key json here`

func main() {
	// Create an IPC based RPC connection to a remote node
	conn, err := ethclient.Dial("/home/jack/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	// Instantiate the contract and display its name
	token, err := Token.NewSDT(common.HexToAddress("0x31ac2908b1f981519b7ab992b46eaa41566b3c0a"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	name, err := token.Name(nil)
	if err != nil {
		log.Fatalf("Failed to retrieve token name: %v", err)
	}
	fmt.Println("Token name:", name)

	// Create an authorized transactor and spend 1 unicorn
	auth, err := bind.NewTransactor(strings.NewReader(key), "my awesome super secret password")
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	tx, err := Token.Transfer(auth, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(1))
	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}
	fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())


}
