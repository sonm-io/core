package blockchainApi
import (
	"fmt"
	"log"
	//"math/big"
	"strings"
	"github.com/sonm-io/go-ethereum/common"
	"github.com/sonm-io/go-ethereum/ethclient"
  	"github.com/sonm-io/blockchain-api/go-build/SDT"
	"github.com/sonm-io/blockchain-api/go-build/Factory"
	"github.com/sonm-io/blockchain-api/go-build/Whitelist"
	"github.com/sonm-io/blockchain-api/go-build/HubWallet"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"encoding/json"
	"io/ioutil"
	"os/user"
	"github.com/sonm-io/go-ethereum/core/types"
	//"github.com/ipfs/go-ipfs/repo/config"
	"math/big"
)
//----ServicesSupporters Allocation---------------------------------------------

//For rinkeby testnet
const confFile = ".rinkeby/keystore/"

//create json for writing KEY
type MessageJson struct {
	//Key       string     `json:"Key"`
	}
//Reading KEY
func readKey() MessageJson{
	//usr, err := user.Current();
	//file, err := ioutil.ReadFile(usr.HomeDir+"/"+confFile)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//var m MessageJson
	//err = json.Unmarshal(file, &m)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//for directory list
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
	first := files[0]
	return first
}

type PasswordJson struct {
	Password		string	`json:"Password"`
}

//Reading user password
// ВОПРОС - Это возвращает JSON структуру или строку?
func readPwd() PasswordJson{
	usr, err := user.Current();
	// User password file JSON should be in root of home directory
	file, err := ioutil.ReadFile(usr.HomeDir+"/")
	if err != nil {
		fmt.Println(err)
	}

	var m PasswordJson
	err = json.Unmarshal(file, &m)
	if err != nil {
		fmt.Println(err)
	}
	return m
}

//--Services Getters-----------------------------
/*
	Those functions allows someone behind library gets
	conn and auth for further interaction

*/

//Establish Connection to geth IPC
// Create an IPC based RPC connection to a remote node
func cnct() {
	// NOTE there is should be wildcard but not username.
	// Try ~/.rinkeby/geth.ipc
	conn, err := ethclient.Dial("/home/cotic/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//return connection obj
  	return conn
}

// Create an authorized transactor
func getAuth() {

	key:= readKey()
	pass:=readPwd()

	auth, err := bind.NewTransactor(strings.NewReader(key), pass)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	return auth
}

//---Defines Binds-----------------------------------------

/*
	Those should be internal functions for internal usage (but not for sure)

*/


//Token Defines
func GlueToken(conn ethclient.Client) (*Token.SDT) {
	// Instantiate the contract
	token, err := Token.NewSDT(common.HexToAddress("0x8016a9f651a4393a608d57d096c464f9115763ea"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	return token
}

func GlueFactory(conn ethclient.Client) (*Factory.Factory) {
	//Define factory
	factory, err := Factory.NewFactory(common.HexToAddress("0x389166c28d119d85f3cd9711e250d856075bd774"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Factory contract: %v", err)
	}
	return factory
}

func GlueWhitelist(conn ethclient.Client) (*Whitelist.Whitelist)  {
	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0xad30096e883f7cc6c1653043751d9ddfe2914a87"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}
	return whitelist
}

func GlueHubWallet(conn ethclient.Client, wb common.Address) (*Hubwallet.HubWallet, error)  {
	//Define HubWallet

	hw, err := Hubwallet.NewHubWallet(wb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a HubWallet contract: %v", err)
	}
	return hw
}

func GlueMinWallet(conn ethclient.Client, mb common.Address) (*MinWallet.MinerWallet, error) {
	//Define MinerWallet
	mw, err := MinWallet.NewMinerWallet(mb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
	}
	return mw
}


func main(){}
// sdt was here???
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
func getBalance(conn ethclient.Client, mb common.Address) (*types.Transaction) {
	token:=GlueToken(conn)
	bal, err := token.BalanceOf(&bind.CallOpts{Pending: true},mb)
	if err != nil {
		log.Fatalf("Failed to request token balance: %v", err)
	}
	return bal
}


func HubTransfer(conn ethclient.Client, auth *bind.TransactOpts, wb common.Address, to common.Address,amount big.Int) (*types.Transaction)  {
	hw:=GlueHubWallet(conn,wb)
	am = big.NewInt(amount *10^17)

	tx, err := hw.Transfer(auth,to,am)
	if err != nil {
		log.Fatalf("Failed to request hub transfer: %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx
}

func WhiteListCall (conn ethclient.Client,)(){
	wl:= GlueWhitelist(conn)
	dp, err := wl.WhitelistCaller()
	if err != nil{
		log.Fatalf("Failed whiteList: %v", err)
	}
	return dp
}
func WhiteListTransactor (conn ethclient.Client,)(){
	wl:= GlueWhitelist(conn)
	dp, err := wl.WhitelistTransactor()
	if err != nil{
		log.Fatalf("Failed whiteList: %v", err)
	}
	return dp
}
func CreateMiner (conn ethclient.Client)(){
	factory := GlueFactory(conn)
	rc, err := factory.FactoryTransactor.CreateMiner()
	if err!= nil{ log.Fatal("Failed to create miner")}
	return  rc

}
func CreateHub (conn ethclient.Client)(){
	factory := GlueFactory(conn)
	chub, err := factory.FactoryTransactor.CreateHub()
	if err!= nil{ log.Fatal("Failed to create hub")}
	return  chub

}
func RegisterMiner (auth *bind.TransactOpts, adr common.Address, stake big.Int)(){
	rm, err := GlueMinWallet(auth, adr)
	stk := big.NewInt(stake * 10^17)
	dp, err := rm.Registration(auth,stk)
	if err != nil {
		log.Fatal("Failed register miner")
	}
	return dp
}
func RegisterHub (auth *bind.TransactOpts, adr common.Address, stake big.Int)(){
	rh, err := GlueHubWallet(auth, adr)
	dp, err := rh.Registration(auth)
	if err != nil {
		log.Fatal("Failed register miner")
	}
	return dp
}
func TransferFunds(hw Hubwallet.HubWallet, auth *bind.TransactOpts , mb common.Address) (*types.Transaction){
	tx, err := hw.Transfer(auth,mb,big.NewInt(2 * 10^17))
	if err != nil {
		log.Fatalf("Failed to request hub transfer: %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())

	return tx
}
func PullingMoney (mw MinWallet.MinerWallet, auth *bind.TransactOpts, wb common.Address) (*types.Transaction) {
	tx, err := mw.PullMoney(auth,wb)
	if err != nil {
		log.Fatalf("Failed to request : %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx
}

func hPayDay ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction){
	ghw, err := GlueHubWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add hub wallet: %v", err)
	}
	tx, err := ghw.PayDay(auth) //auth
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx
}


func mPayDay ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction){
	gmw,err := GlueMinWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add miner wallet: %v", err)
	}
	tx, err := gmw.PayDay(auth)
	if err != nil {
		log.Fatalf("Failed to pay you your money: %v", err)
	}
	return tx
}

func hWithdraw ( auth *bind.TransactOpts, mb common.Address) (*types.Transaction){
	hw,err := GlueMinWallet(auth,mb)
	if err != nil {
		log.Fatalf("Failed to add miner wallet: %v", err)
	}
	tx,err := hw.Withdraw(auth)
	if err != nil {
		log.Fatalf("Failed to withdraw from hub: %v", err)
	}
	return tx
}

func hSuspect ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	ghs,err := GlueHubWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add suspected hub: %v", err)
	}
	tx, err := ghs.Suspect(auth)
	if err != nil {
		log.Fatalf("Failed to all your funds are belong to us: %v", err)
	}
	return tx
}


func mSuspect ( auth bind.TransactOpts, mb common.Address) (*types.Transaction, error){
	gms,err := GlueMinWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add suspected miner: %v", err)
	}
	tx, err := gms.Suspect(auth)
	if err != nil {
		log.Fatalf("Failed to all your funds are belong to us: %v", err)
	}
	return tx
}


func hGulag ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	ggh, err := GlueHubWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add punished hub: %v", err)
	}
	tx, err := ggh.Gulag(auth)
	if err != nil {
		log.Fatalf("Failed to anathemize hub: %v", err)
	}
	return tx
}

func mGulag ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	gmh, err := GlueMinWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add punished miner: %v", err)
	}
	tx, err := gmh.Gulag(auth)
	if err != nil {
		log.Fatalf("Failed to anathemize miner: %v", err)
	}
	return tx
}

func hRehub ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	grh, err := GlueHubWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add this hub to saints list: %v", err)
	}
	tx, err := grh.Rehub(auth)
	if err != nil {
		log.Fatalf("Failed to forgive this hub: %v", err)
	}
	return tx
}

func mRehub ( auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	grm, err := GlueMinWallet(auth, mb)
	if err != nil {
		log.Fatalf("Failed to add this miner to saints list: %v", err)
	}
	tx, err := grm.Rehub(auth)
	if err != nil {
		log.Fatalf("Failed to forgive this miner: %v", err)
	}
	return tx
}

func CheckMiners (auth *bind.TransactOpts, mb common.Address) (*types.Transaction, error){
	cm := GlueWhitelist(auth)
	tx, err := cm.RegistredMiners(&bind.CallOpts{Pending: true}, mb)

	if err != nil {
		log.Fatalf("Failed to retrieve miner wallet: %v", err)
	}
	return tx
}

func CheckHubs (auth *bind.TransactOpts , mb common.Address) (*types.Transaction, error){
	ch := GlueWhitelist(auth)
	tx, err := ch.RegistredHubs(&bind.CallOpts{Pending: true}, mb)
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}
	return tx
}

