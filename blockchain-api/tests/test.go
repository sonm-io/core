package main

import (
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
)

func main() {

	pass := blockchain.ReadPwd()
	fmt.Println("pass:", pass)

	key := blockchain.ReadKey()
	// Need to fix it to wildcard
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

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	hAddr, err := blockchain.GetHubAddr(conn, owner)
	if err != nil {
		log.Fatalf("Failed to create hub : %v", err)
	}
	fmt.Println("hub address:")
	hAdr := hAddr.String()
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

	tx, err := blockchain.TransferToken(conn, auth, hAddr, 5)
	if err != nil {
		log.Fatalf("Failed to transfer token: %v", err)
	}
	fmt.Println("tx:", tx)

	fmt.Println("Wait 10 sec!")
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

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	checklistHubWl, err := blockchain.CheckHubs(conn, hAddr)
	if err != nil {
		log.Fatalf("Failed to check hubs: %v", err)
	}
	fmt.Println("checklistHubWl:", checklistHubWl)

	//-------------------- 2 ------------------------------
	//--MINER INIT--//
	//Create Min wallet
	m, err := blockchain.CreateMiner(conn, auth)
	if err != nil {
		log.Fatalf("Failed to create miner : %v", err)
	}
	fmt.Println("tx:")
	fmt.Println(m)

	mAddr, err := blockchain.GetMinAddr(conn, owner)
	if err != nil {
		log.Fatalf("Failed to create miner : %v", err)
	}
	fmt.Println("miner address:")
	mAdr := mAddr.String()
	fmt.Println(mAdr)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	tx, err = blockchain.TransferToken(conn, auth, mAddr, 5)
	if err != nil {
		log.Fatalf("Failed to transfer token miner: %v", err)
	}
	fmt.Println("tx:", tx)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	regMinwl, err := blockchain.RegisterMiner(conn, auth, mAddr, 1) // registration in whitelist
	if err != nil {
		log.Fatalf("Failed to register miner in whitelist: %v", err)
	}
	fmt.Println("regMinwl:", regMinwl)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	checkregistrMinerWl, err := blockchain.CheckMiners(conn, mAddr)
	if err != nil {
		log.Fatalf("Failed to check Miner: %v", err)
	}
	fmt.Println("checkregistrWl:", checkregistrMinerWl)

	//-------------------- 3 ------------------------------
	//--END--//
	//Hub sent money to miner

	hubToMin, err := blockchain.HubTransfer(conn, auth, hAddr, mAddr, 2)
	if err != nil {
		log.Fatalf("Failed to hub transfer to miner wallet: %v", err)
	}
	fmt.Println("hubToMin:", hubToMin)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	minerPullMoney, err := blockchain.PullingMoney(conn, auth, mAddr, hAddr)
	if err != nil {
		log.Fatalf("Failed to miner pulling money from hub: %v", err)
	}
	fmt.Println("minerPullMoney:", minerPullMoney)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	fmt.Println("Wait 10 sec!")
	time.Sleep(10 * time.Second)

	getBalanceHub, err := blockchain.GetBalance(conn, hAddr)
	if err != nil {
		log.Fatalf("Failed to get balance for hub: %v", err)
	}

	b_dec := big.NewInt(1000000000000000000)
	//b_dec=1000000000000000000
	balHub := new(big.Int)
	balHub.Div(getBalanceHub, b_dec)
	fmt.Println("getBalanceHub:", balHub)

	getBalanceMiner, err := blockchain.GetBalance(conn, mAddr)
	if err != nil {
		log.Fatalf("Failed to get balance for miner: %v", err)
	}

	balMin := new(big.Int)
	balMin.Div(getBalanceMiner, b_dec)

	fmt.Println("getBalanceMiner:", balMin)

}
