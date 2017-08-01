//go:generate go run generate-api.go

package api

import (
	"fmt"
	"log"
	"strings"

	"io/ioutil"
	"os/user"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/contracts/api/Factory"
	"github.com/sonm-io/core/contracts/api/HubWallet"
	"github.com/sonm-io/core/contracts/api/MinerWallet"
	"github.com/sonm-io/core/contracts/api/SonmDummyToken"
	"github.com/sonm-io/core/contracts/api/Whitelist"

	"math"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

//----ServicesSupporters Allocation---------------------------------------------

func getPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("cant get user", err)
	}
	//	fmt.Println( usr.HomeDir )

	// home directory
	hd := usr.HomeDir

	//conf for keystore
	const confFile = "/.rinkeby/keystore/"

	confPath := hd + confFile
	return confPath
}

func GHome() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("cant get user", err)
	}
	//	fmt.Println( usr.HomeDir )

	// home directory
	hd := usr.HomeDir

	return hd
}

func ReadKey() string {

	confPath := getPath()

	files, err := ioutil.ReadDir(confPath)
	if err != nil {
		log.Fatal("can't read dir", err)
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
	first := files[0]
	fName := first.Name()

	keyf, err := ioutil.ReadFile(confPath + fName)
	if err != nil {
		log.Fatalf("can't read the file", err)
	}

	key := string(keyf)

	fmt.Println("key")
	fmt.Println(key)

	return key
}

func ReadPwd() string {

	hd := GHome()

	npass := "/pass.txt"

	hnpass := hd + npass

	passf, err := ioutil.ReadFile(hnpass)
	if err != nil {
		log.Fatalf("can't read the file", err)
	}

	pass := string(passf)
	pass = strings.TrimRight(pass, "\n")

	//	fmt.Println("password:")
	//	fmt.Println(pass)

	return pass

}

//Token Defines
func GlueToken(conn bind.ContractBackend) (*SonmDummyToken.SonmDummyToken, error) {
	// Instantiate the contract
	token, err := SonmDummyToken.NewSonmDummyToken(common.HexToAddress("0x8016a9f651a4393a608d57d096c464f9115763ea"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	return token, err
}

func GlueFactory(conn bind.ContractBackend) (*Factory.Factory, error) {
	//Define factory
	factory, err := Factory.NewFactory(common.HexToAddress("0x389166c28d119d85f3cd9711e250d856075bd774"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}
	return factory, err
}

func GlueWhitelist(conn bind.ContractBackend) (*Whitelist.Whitelist, error) {
	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0xad30096e883f7cc6c1653043751d9ddfe2914a87"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}
	return whitelist, err
}

func GlueHubWallet(conn bind.ContractBackend, wb common.Address) (*HubWallet.HubWallet, error) {
	//Define HubWallet

	hw, err := HubWallet.NewHubWallet(wb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a HubWallet contract: %v", err)
	}
	return hw, err
}

func GlueMinWallet(conn bind.ContractBackend, mb common.Address) (*MinerWallet.MinerWallet, error) {
	//Define MinerWallet
	mw, err := MinerWallet.NewMinerWallet(mb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
	}
	return mw, err
}

//-------------------------------------------------------------------------

//--MAIN LIBRARY-----------------------------------------------------------

/*
		HOW THIS SHOULD WORK?

		MainProgram				ThisLibrary
		A<---------------->B

	First A program wants to get Auth and Connection,
	So it will ask for getAuth and cnct functions from b
	and should store those objects inside.

	Therefore A want to interact with smart contracts functions from Blockchain,
	so it is call such functions from B with conn and auth in parameters.
*/
/*
Example
*/
func GetBalance(conn bind.ContractBackend, mb common.Address) (*big.Int, error) {
	token, err := GlueToken(conn)
	bal, err := token.BalanceOf(&bind.CallOpts{Pending: true}, mb)
	if err != nil {
		log.Fatalf("Failed to request token balance: %v", err)
	}
	return bal, err
}

func TransferToken(conn bind.ContractBackend, auth *bind.TransactOpts, to common.Address, amount float64) (*types.Transaction, error) {
	token, err := GlueToken(conn)
	if err != nil {
		log.Fatalf("Failed to glue to HubWallet: %v", err)
	}
	//dec:=big.NewInt(10^17)
	dec := math.Pow(10, 18)
	di := int64(dec)
	am := int64(amount)
	am = am * di

	amb := big.NewInt(am)

	tx, err := token.Transfer(auth, to, amb)
	if err != nil {
		log.Fatalf("Failed to request token transfer: %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx, err
}

func HubTransfer(conn bind.ContractBackend, auth *bind.TransactOpts, wb common.Address, to common.Address, amount float64) (*types.Transaction, error) {
	hw, err := GlueHubWallet(conn, wb)
	if err != nil {
		log.Fatalf("Failed to glue to HubWallet: %v", err)
	}
	//dec:=big.NewInt(10^17)
	dec := math.Pow(10, 18)
	di := int64(dec)
	am := int64(amount)
	am = am * di

	amb := big.NewInt(am)

	tx, err := hw.Transfer(auth, to, amb)
	if err != nil {
		log.Fatalf("Failed to request hub transfer: %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx, err
}

func CreateMiner(conn bind.ContractBackend, auth *bind.TransactOpts) (*types.Transaction, error) {
	factory, err := GlueFactory(conn)

	tx, err := factory.CreateMiner(auth)
	if err != nil {
		log.Fatalf("Failed to request hub creation: %v", err)
	}
	fmt.Println("CreateMiner pending: 0x%x\n", tx.Hash())

	// Don't even wait, check its presence in the local pending state
	time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

	return tx, err
}

func CreateHub(conn bind.ContractBackend, auth *bind.TransactOpts) (*types.Transaction, error) {
	factory, err := GlueFactory(conn)

	tx, err := factory.CreateHub(auth)
	if err != nil {
		log.Fatalf("Failed to request hub creation: %v", err)
	}
	fmt.Println("CreateHub pending: 0x%x\n", tx.Hash())

	// Don't even wait, check its presence in the local pending state
	time.Sleep(250 * time.Millisecond) // Allow it to be processed by the local node :P

	return tx, err
}

func RegisterMiner(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address, stake float64) (*types.Transaction, error) {
	rm, err := GlueMinWallet(conn, adr)

	dec := math.Pow(10, 18)
	di := int64(dec)
	stk := int64(stake)
	stk = stk * di

	stb := big.NewInt(stk)

	tx, err := rm.Registration(auth, stb)
	if err != nil {
		log.Fatal("Failed register miner")
	}
	return tx, err
}

func RegisterHub(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	rm, err := GlueHubWallet(conn, adr)

	tx, err := rm.Registration(auth)
	if err != nil {
		log.Fatal("Failed register hub")
	}
	return tx, err
}

func PullingMoney(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address, from common.Address) (*types.Transaction, error) {
	mw, err := GlueMinWallet(conn, adr)
	tx, err := mw.PullMoney(auth, from)
	if err != nil {
		log.Fatalf("Failed to request : %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx, err
}

//func balanceMiner (mw MinWallet.MinerWallet, wb common.Address, backend *bind.ContractBackend) () {
//	token,err := Token.NewSDT(wb,backend)
//		if err != nil {
//			log.Fatalf("Failed to instantiate a Token contract: %v", err)
//		}
//	bal, err := token.BalanceOf(&bind.CallOpts{Pending: true},wb)
//}

func hPayDay(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	ghw, err := GlueHubWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := ghw.PayDay(auth) //auth
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx, err
}

func mPayDay(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	gmw, err := GlueMinWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := gmw.PayDay(auth) //auth
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx, err
}

func hWithdraw(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	hw, err := GlueHubWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := hw.Withdraw(auth) //auth
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx, err
}

func mWithdraw(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	mw, err := GlueMinWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := mw.Withdraw(auth) //auth
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx, err
}

func hSuspect(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	hw, err := GlueHubWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := hw.Suspect(auth)
	if err != nil {
		log.Fatalf("Failed to suspect: %v", err)
	}
	return tx, err
}

func mSuspect(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	mw, err := GlueMinWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := mw.Suspect(auth)
	if err != nil {
		log.Fatalf("Failed to suspect: %v", err)
	}
	return tx, err
}

func hGulag(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	hw, err := GlueHubWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := hw.Gulag(auth)
	if err != nil {
		log.Fatalf("Failed to gulag: %v", err)
	}
	return tx, err
}

func mGulag(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	mw, err := GlueMinWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := mw.Gulag(auth)
	if err != nil {
		log.Fatalf("Failed to gulag: %v", err)
	}
	return tx, err
}

func hRehub(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	hw, err := GlueHubWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := hw.Rehub(auth)
	if err != nil {
		log.Fatalf("Failed to rehub: %v", err)
	}
	return tx, err
}

func mRehub(conn bind.ContractBackend, auth *bind.TransactOpts, adr common.Address) (*types.Transaction, error) {
	mw, err := GlueMinWallet(conn, adr)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := mw.Rehub(auth)
	if err != nil {
		log.Fatalf("Failed to rehub: %v", err)
	}
	return tx, err
}

func CheckMiners(conn bind.ContractBackend, mb common.Address) (bool, error) {
	w, err := GlueWhitelist(conn)
	state, err := w.RegistredMiners(&bind.CallOpts{Pending: true}, mb)

	if err != nil {
		log.Fatalf("Failed to retrieve miner wallet: %v", err)
	}
	return state, err
}

func CheckHubs(conn bind.ContractBackend, mb common.Address) (bool, error) {
	//wRegisteredHubs
	w, err := GlueWhitelist(conn)
	state, err := w.RegistredHubs(&bind.CallOpts{Pending: true}, mb)
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}
	return state, err
}

func GetHubAddr(conn bind.ContractBackend, owner common.Address) (common.Address, error) {
	f, err := GlueFactory(conn)
	addr, err := f.Hubs(&bind.CallOpts{Pending: true}, owner)
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}
	return addr, err
}

func GetMinAddr(conn bind.ContractBackend, owner common.Address) (common.Address, error) {
	f, err := GlueFactory(conn)
	addr, err := f.Miners(&bind.CallOpts{Pending: true}, owner)
	if err != nil {
		log.Fatalf("Failed to retrieve miner wallet: %v", err)
	}
	return addr, err
}
