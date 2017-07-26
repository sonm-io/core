package main

import (
	"github.com/sonm-io/core/blockchain-api"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/ethclient"
	"log"
	"strings"

	"github.com/sonm-io/go-ethereum/common"
	"time"

	"fmt"
)

func main() {

	pass := blockchain.ReadPwd()
	fmt.Println("pass:", pass)

	key := blockchain.ReadKey()
	owner := common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1")
	hd := blockchain.GHome()

	conn, err := ethclient.Dial(hd + "/.rinkeby/geth.ipc")
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
	h, err := blockchain.CreateHub(conn, auth)
	if err != nil {
		log.Fatalf("Failed to create hub : %v", err)
	}
	fmt.Println("tx:")
	fmt.Println(h)

	fmt.Println("Wait!")
	time.Sleep(10 * time.Second)

	hAddr, err := blockchain.GetHubAddr(conn, owner)
	if err != nil {
		log.Fatalf("Failed to create hub : %v", err)
	}
	fmt.Println("hub address:")
	hAdr := hAddr.String()
	fmt.Println(hAdr)
}
