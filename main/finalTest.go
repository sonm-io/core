package main
import (
	"log"
	"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"time"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"fmt"
)
func main (){
	//Create connection
	conn, err := ethclient.Dial("/home/cotic/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//Request info about hubs
	hubof, err := factory.HubOf(&bind.CallOpts{Pending: true}, common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1"))
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}

	wb:=hubof
	w:=hubof.String()
	//Create Hub Wallet
	hubWallet, err := blockchainApi.GlueHubWallet(conn, wb)
	if err != nil {
		log.Fatalf("Failed to create hub wallet: %v", err)
	}
	return hubWallet

	//Transfer token to HubWallet
	token, err := blockchainApi.HubTransfer()

	time.Sleep(250 * time.Millisecond)
	//Test balance
	bal, err = blockchainApi(&bind.CallOpts{Pending: true},mb)
	if err != nil {
		log.Fatalf("Failed to request token balance: %v", err)
	}
	// Need to do something about checking pending tx
	//bal=tx

	fmt.Printf("Balance of Miner", bal)

	////Define whitelist
	//whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0x833865a1379b9750c8a00b407bd6e2f08e465153"), conn)
	whitelist, err := blockchainApi.WhiteListCall(conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}
	return whitelist

}

