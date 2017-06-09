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

)

//-------INIT ZONE--------------------------------------------------------------

	// THIS IS HACK AND SHOULD BE REWRITTEN
//	const key = `{"address":"fe36b232d4839fae8751fa10768126ee17a156c1","crypto":{"cipher":"aes-128-ctr","ciphertext":"b2f1390ba44929e2144a44b5f0bdcecb06060b5ef1e9b0d222ed0cd5340e2876","cipherparams":{"iv":"a33a90fc4d7a052db58be24bbfdc21a3"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"422328336107aeb54b4a152f4fae0d5f2fbca052fc7688d9516cd998cf790021"},"mac":"08f3fa22882b932ae2926f6bf5b1df2c0795720bd993b50d652cee189c00315c"},"id":"b36be1bf-6eb4-402e-8e26-86da65ae3156","version":3}`

func packageblockchain() {
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

	//Define factory
	factory, err := Factory.NewFactory(common.HexToAddress("0xDC0B27895bA9316571799C4044109c452Eb1bC14"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}

	//Put correct addresses into Factory
	// auth, dao-address, Whitelist address as @params
	tx, err = factory.ChangeAdresses(auth,common.HexToAddress("0xFE36B232D4839FAe8751fa10768126ee17A156c1"),common.HexToAddress("0x4d98d99e9b74d66fc2b4ac49070422b0f514339b"))
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

	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0x4d98d99e9b74d66fc2b4ac49070422b0f514339b"), conn)
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


//Registry Miner in Whitelist



//Define MinerWallet
mw, err := MinWallet.NewMinerWallet(mb, conn)
if err != nil {
	log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
}

//Register MinerWallet
tx, err = mw.Registration(auth,big.NewInt(1))
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

balance, err := token.BalanceOf(&bind.CallOpts{Pending: true},wb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
bal:=balance

fmt.Printf("Balance of Hub", bal)

//Balance of Miner
balance, err = token.BalanceOf(&bind.CallOpts{Pending: true},mb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
bal=balance

fmt.Printf("Balance of Miner", bal)

//Make some initial supplyment to hubwallet

tx, err = token.Transfer(auth, wb, big.NewInt(10))
if err != nil {
	log.Fatalf("Failed to request token transfer: %v", err)
}
// Need to do something about checking pending tx
fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())

//Transfer some as a payout

//Register HubWallet
tx, err = hw.Transfer(auth,mb,big.NewInt(2))
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
////////////////////////////////
balance, err = token.BalanceOf(&bind.CallOpts{Pending: true},mb)
if err != nil {
	log.Fatalf("Failed to request token balance: %v", err)
}
// Need to do something about checking pending tx
bal=balance

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
// HubWalletCaller is an auto generated read-only Go binding around an Ethereum contract.

//Create a HubWallet
//type HubWalletCaller struct {
//	contract *bind.BoundContract // Generic contract wrapper for the low level calls
//}
//func NewHubWallet(address common.Address, backend bind.ContractBackend) (*HubWallet, error) {
//	contract, err := bindHubWallet(address, backend, backend)
//	if err != nil {
//		return nil, err
//	}
//	return &HubWallet{HubWalletCaller: HubWalletCaller{contract: contract}, HubWalletTransactor: HubWalletTransactor{contract: contract}}, nil
//}
//type HubWallet struct {
//	HubWalletCaller     // Read-only binding to the contract
//	HubWalletTransactor // Write-only binding to the contract
//}
//type HubWalletTransactor struct {
//	contract *bind.BoundContract // Generic contract wrapper for the low level calls
//}
//
//func bindHubWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
//	parsed, err := abi.JSON(strings.NewReader(HubWalletABI))
//	if err != nil {
//		return nil, err
//	}
//	return bind.NewBoundContract(address, parsed, caller, transactor), nil
//}
//const HubWalletABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"currentPhase\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezePeriod\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"gulag\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"suspect\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"sharesTokenAddress\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"genesisTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"PayDay\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"freezeQuote\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenTime\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"frozenFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockPercent\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"DAO\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_to\",\"type\":\"address\"},{\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockedFunds\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"rehub\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"Factory\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[],\"name\":\"Registration\",\"outputs\":[{\"name\":\"success\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"payable\":false,\"type\":\"function\"},{\"inputs\":[{\"name\":\"_hubowner\",\"type\":\"address\"},{\"name\":\"_dao\",\"type\":\"address\"},{\"name\":\"_whitelist\",\"type\":\"address\"},{\"name\":\"sharesAddress\",\"type\":\"address\"}],\"payable\":false,\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"newPhase\",\"type\":\"uint8\"}],\"name\":\"LogPhaseSwitch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"pass\",\"type\":\"string\"}],\"name\":\"LogPass\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"val\",\"type\":\"uint256\"}],\"name\":\"ToVal\",\"type\":\"event\"}]"
//// Whitelist is an auto generated Go binding around an Ethereum contract.
//type Whitelist struct {
//	WhitelistCaller     // Read-only binding to the contract
//	WhitelistTransactor // Write-only binding to the contract
//}
//
//// WhitelistCaller is an auto generated read-only Go binding around an Ethereum contract.
//type WhitelistCaller struct {
//	contract *bind.BoundContract // Generic contract wrapper for the low level calls
//}
//// WhitelistTransactor is an auto generated write-only Go binding around an Ethereum contract.
//type WhitelistTransactor struct {
//	contract *bind.BoundContract // Generic contract wrapper for the low level calls
//}
//// NewWhitelist creates a new instance of Whitelist, bound to a specific deployed contract.
//func NewWhitelist(address common.Address, backend bind.ContractBackend) (*Whitelist, error) {
//	contract, err := bindWhitelist(address, backend, backend)
//	if err != nil {
//		return nil, err
//	}
//	return &Whitelist{WhitelistCaller: WhitelistCaller{contract: contract}, WhitelistTransactor: WhitelistTransactor{contract: contract}}, nil
//}
//// bindWhitelist binds a generic wrapper to an already deployed contract.
//func bindWhitelist(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor) (*bind.BoundContract, error) {
//	parsed, err := abi.JSON(strings.NewReader(WhitelistABI))
//	if err != nil {
//		return nil, err
//	}
//	return bind.NewBoundContract(address, parsed, caller, transactor), nil
//}
//// WhitelistABI is the input ABI used to generate the binding from.
const WhitelistABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"},{\"name\":\"stakeShare\",\"type\":\"uint256\"}],\"name\":\"RegisterMin\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"}],\"name\":\"UnRegisterMiner\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\"},{\"name\":\"wallet\",\"type\":\"address\"},{\"name\":\"time\",\"type\":\"uint64\"}],\"name\":\"RegisterHub\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"type\":\"function\"}]"