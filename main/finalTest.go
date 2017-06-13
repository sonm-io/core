package main
import (
	"log"
	"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"strings"
	"io/ioutil"
	"github.com/sonm-io/go-ethereum/accounts/keystore"
	"github.com/sonm-io/go-ethereum/common"
	"time"
	"math/big"
	"fmt"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
)
func main (){
	//-------------------- 1 ------------------------------


	pass:=blockchainApi.readPwd()


	key:=blockchainApi.readKey()


	hd:=blockchainApi.gHome()

	conn, err := ethclient.Dial(hd+"/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}



	auth, err := bind.NewTransactor(strings.NewReader(key), pass)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}


/*
	keyin := strings.NewReader(key)
	passphrase := "Metamorph9"

	json, err := ioutil.ReadAll(keyin)
		if err != nil {
			return nil, err
		}
	key, err := keystore.DecryptKey(json, passphrase)
		if err != nil {
			return nil, err
		}

	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

	//Create connection
	conn, err := ethclient.Dial("/home/jack/.rinkeby/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}
		*/

	//-------------------- 2 ------------------------------
	//--HUB INIT--//
	//Create Hub
	h, err := blockchainApi.CreateHub(conn)
		if err != nil {
			log.Fatalf("Failed to create hub : %v", err)
		}

	//create hub wallet
	hw, err := blockchainApi.GlueHubWallet(conn, h)
		if err != nil {
			log.Fatalf("Failed to create hub wallet: %v", err)
		}
	fmt.Println("CreateHub", hw)

	time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

	// Instantiate the contract and display its name
	//create tokens
	ct, err := blockchainApi.GlueToken(conn)
	if err != nil {
		log.Fatalf("Failed to : %v", err)
	}
	wb:= common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1")
	mb := common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1")
	fmt.Println("Token name:", ct)
	//to :=
	//am :=
	//sent tokens
	st, err := blockchainApi.HubTransfer(conn, auth, wb, to, am)
		if err != nil {
			log.Fatalf("Failed to sent tokens to hub wallet: %v", err)
		}

	fmt.Println("Sent tokens: 0x%x\n", st.Hash())

	time.Sleep(250 * time.Millisecond)

	bal, err := blockchainApi.GetBalance(conn, mb)
		if err != nil {
			log.Fatalf("Failed to request token balance: %v", err)
		}
	fmt.Println("Balance:", bal.Hash())

	time.Sleep(250 * time.Millisecond)

	//registered hub in whitelist
	stake := big.Int{}
	stk := big.NewInt(stake * 10^17)

	wl, err := blockchainApi.RegisterHub(auth, wb, stk)
		if err != nil {
			log.Fatalf("Failed to  %v", err)
		}
	time.Sleep(250 * time.Millisecond)

	//Request info about hubs
	checkHubsWl, err := blockchainApi.CheckHubs(conn, wb)
	wf:=checkHubsWl
	fmt.Println("Wallet address is:", wf)

	//////////////////////////////////////////////////////////////////
	//----------------------MINER INIT------------------------------//
	//Create Miner
	m, err := blockchainApi.CreateMiner(conn)
		if err != nil {
			log.Fatalf("Failed to request Miner creation: %v", err)
		}
	fmt.Println("Created miner name:", m)
	//create hub wallet

	mw, err := blockchainApi.GlueMinWallet(conn, h)
		if err != nil {
			log.Fatalf("Failed to create min wallet: %v", err)
		}
	fmt.Println("CreateHub", mw)
	//create tokens
	ct, err := blockchainApi.GlueToken(conn)
		if err != nil {
			log.Fatalf("Failed to : %v", err)
		}
	fmt.Println("Token name:", ct)

	//sent tokens
	st, err := blockchainApi.HubTransfer(conn, auth, wb, to, am)
		if err != nil {
			log.Fatalf("Failed to sent tokens to hub wallet: %v", err)
		}


	time.Sleep(250 * time.Millisecond)

	bal, err := blockchainApi.GetBalance(conn, mb)
		if err != nil {
			log.Fatalf("Failed to request token balance: %v", err)
		}

	time.Sleep(250 * time.Millisecond)

	wl, err := blockchainApi.RegisterMiner(auth, wb, stk)
		if err != nil {
			log.Fatalf("Failed to  %v", err)
		}
	time.Sleep(250 * time.Millisecond)

	//Request info about hubs
	checkMinWl, err := blockchainApi.CheckHubs(conn, wb)
		wf:=checkMinWl
	fmt.Println("Wallet address is:", wf)
	//-------------------- 3 ------------------------------
	st, err := blockchainApi.HubTransfer(conn, auth, wb, to, am)
		if err != nil {
			log.Fatalf("Failed to sent tokens to hub wallet: %v", err)
		}
	wm:= MinWallet.MinerWallet{}
	pm, err := blockchainApi.PullingMoney(wm, auth, wb)
		if err != nil {
			log.Fatalf("Failed to miner`a pulling money: %v", err)
		}
	fmt.Println("Pulling money is:", pm)
	time.Sleep(250 * time.Millisecond)

	balMin := blockchainApi.GetBalance(conn, mb)
	balHub := blockchainApi.GetBalance(conn, wb)
}
