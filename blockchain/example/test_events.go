package main

import (
	"log"

	"github.com/sonm-io/core/blockchain"
)

const defaultEthEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

//const defaultEthEndpoint string = "/Users/anton/Library/Ethereum/rinkeby/geth.ipc"

func main() {
	bch, err := blockchain.NewBlockchainAPI(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	//Addr := "41ba7e0e1e661f7114f2f05afd2536210c2ed351"
	dealsIds, err := bch.GetOpenedDeal(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(dealsIds)
}
