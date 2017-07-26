package main

import (
	"github.com/sonm-io/core/blockchain-api"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
	"log"
	"strings"
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

	// Transfer 5 tokens to hub address to be sure in deposit
	tx, err := blockchain.TransferToken(conn, auth, hAddr, 5)
	if err != nil {
		log.Fatalf("Failed to transfer token hub: %v", err)
	}
	fmt.Println("tx:", tx)

	time.Sleep(10 * time.Second)

	bal, err := blockchain.GetBalance(conn, hAddr)
	if err != nil {
		log.Fatalf("Failed to get balance: %v", err)
	}
	fmt.Println("bal:", bal)

	regHubwl, err := blockchain.RegisterHub(conn, auth, hAddr) // registration in whitelist
	if err != nil {
		log.Fatalf("Failed to register hub in whitelist: %v", err)
	}
	fmt.Println("regHubwl:", regHubwl)

	time.Sleep(10 * time.Second)

	checklistHubWl, err := blockchain.CheckHubs(conn, hAddr)
	if err != nil {
		log.Fatalf("Failed to check hubs: %v", err)
	}
	fmt.Println("checklistHubWl:", checklistHubWl)

}
