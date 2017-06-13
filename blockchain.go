package blockchainApi
import (
	"fmt"
	"log"
	//"math/big"
	"strings"
	"github.com/sonm-io/go-ethereum/common"
 //"github.com/sonm-io/go-ethereum/crypto"
	//"github.com/sonm-io/go-ethereum/ethclient"
  	"github.com/sonm-io/blockchain-api/go-build/SDT"
	"github.com/sonm-io/blockchain-api/go-build/Factory"
	"github.com/sonm-io/blockchain-api/go-build/Whitelist"
	"github.com/sonm-io/blockchain-api/go-build/HubWallet"
	"github.com/sonm-io/blockchain-api/go-build/MinWallet"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	//"encoding/json"
	"io/ioutil"
	"os/user"
	"github.com/sonm-io/go-ethereum/core/types"
	//"github.com/ipfs/go-ipfs/repo/config"
	"math/big"
	"math"
)
//----ServicesSupporters Allocation---------------------------------------------

func getPath() string {
usr, err := user.Current()
	if err != nil {
			log.Fatal("cant get user", err )
	}
//	fmt.Println( usr.HomeDir )

	// home directory
	hd:=usr.HomeDir

	//conf for keystore
	const confFile = "/.rinkeby/keystore/"

	confPath:= hd+confFile
	return confPath
}

	func gHome() string {
		usr, err := user.Current()
			if err != nil {
					log.Fatal("cant get user", err )
			}
		//	fmt.Println( usr.HomeDir )

			// home directory
			hd:=usr.HomeDir

			return hd
	}

	func readKey() string {

		confPath:=getPath()

	files, err := ioutil.ReadDir(confPath)
	if err != nil {
		log.Fatal("can't read dir", err)
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}
	first := files[0]
	fName:= first.Name()

	keyf, err:=ioutil.ReadFile(confPath+fName)
	if err != nil {
        log.Fatalf("can't read the file", err)
    }


		key:=string(keyf)

	fmt.Println("key")
	fmt.Println(key)

	return key
}


	func readPwd() string {

		hd:=gHome()

		npass:="/pass.txt"

		hnpass:=hd+npass

		passf, err:=ioutil.ReadFile(hnpass)
		if err != nil {
	        log.Fatalf("can't read the file", err)
	    }


			pass:=string(passf)
			pass=strings.TrimRight(pass,"\n")

		//	fmt.Println("password:")
		//	fmt.Println(pass)

			return pass

	}



//--Services Getters-----------------------------
/*
	Those functions allows someone behind library gets
	conn and auth for further interaction

*/






/*
//Establish Connection to geth IPC
// Create an IPC based RPC connection to a remote node
func cnct()  *ethclient.Client{
	hd:=gHome()
	conn, err := ethclient.Dial(hd+"/.rinkeby/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//return connection obj
  	return conn
}
*/



/*
// Create an authorized transactor
func getAuth() *bind.TransactOpts {

	key:= readKey()
	pass:=readPwd()

	auth, err := bind.NewTransactor(strings.NewReader(key), pass)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	return auth
}
*/

//---Defines Binds-----------------------------------------

/*
	Those should be internal functions for internal usage (but not for sure)

*/


//Token Defines
func GlueToken(conn bind.ContractBackend) (*Token.SDT, error) {
	// Instantiate the contract
	token, err := Token.NewSDT(common.HexToAddress("0x8016a9f651a4393a608d57d096c464f9115763ea"), conn)
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

func GlueWhitelist(conn bind.ContractBackend) (*Whitelist.Whitelist, error)  {
	//Define whitelist
	whitelist, err := Whitelist.NewWhitelist(common.HexToAddress("0xad30096e883f7cc6c1653043751d9ddfe2914a87"), conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a Whitelist contract: %v", err)
	}
	return whitelist, err
}

func GlueHubWallet(conn bind.ContractBackend, wb common.Address) (*Hubwallet.HubWallet, error)  {
	//Define HubWallet

	hw, err := Hubwallet.NewHubWallet(wb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a HubWallet contract: %v", err)
	}
	return hw, err
}

func GlueMinWallet(conn bind.ContractBackend, mb common.Address) (*MinWallet.MinerWallet, error) {
	//Define MinerWallet
	mw, err := MinWallet.NewMinerWallet(mb, conn)
	if err != nil {
		log.Fatalf("Failed to instantiate a MinWallet contract: %v", err)
	}
	return mw, err
}



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
func GetBalance(conn bind.ContractBackend, mb common.Address) (*big.Int, error) {
	token, err :=GlueToken(conn)
	bal, err := token.BalanceOf(&bind.CallOpts{Pending: true}, mb)
	if err != nil {
		log.Fatalf("Failed to request token balance: %v", err)
	}
	return bal, err
}


func HubTransfer(conn bind.ContractBackend, auth *bind.TransactOpts, wb common.Address, to common.Address,amount *big.Int) (*types.Transaction, error)  {
	hw,err:=GlueHubWallet(conn,wb)
	if err != nil {
		log.Fatalf("Failed to glue to HubWallet: %v", err)
	}
	//dec:=big.NewInt(10^17)
	dec:=math.Pow(10,17)
	di:= int64(dec)

	db:=big.NewInt(di)

	am := amount * db


	tx, err := hw.Transfer(auth,to,am)
	if err != nil {
		log.Fatalf("Failed to request hub transfer: %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx, err
}

//I don't even know what is it and who did that
/*func WhiteListCall (conn ethclient.Client,)(){
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
*/


func CreateMiner (conn bind.ContractBackend)(MinWallet.MinerWallet, error){
	factory := GlueFactory(conn)
	rc, err := factory.FactoryTransactor.CreateMiner()
	if err!= nil{ log.Fatal("Failed to create miner")}
	return  rc

}
func CreateHub (conn bind.ContractBackend)(Hubwallet.HubWallet, error){
	factory := GlueFactory(conn)
	chub, err := factory.FactoryTransactor.CreateHub()
	return  chub, err

}
func RegisterMiner (auth *bind.TransactOpts, adr common.Address, stake big.Int)(error){
	rm, err := GlueMinWallet(auth, adr)
	stk := big.NewInt(stake * (10^17))
	dp, err := rm.Registration(auth,stk)
	if err != nil {
		log.Fatal("Failed register miner")
	}
	return dp
}
func RegisterHub (auth *bind.TransactOpts, adr common.Address, stake big.Int)(error){
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
func PullingMoney (mw MinWallet.MinerWallet, auth *bind.TransactOpts, wb common.Address) (*types.Transaction, error) {
	tx, err := mw.PullMoney(auth,wb)
	if err != nil {
		log.Fatalf("Failed to request : %v", err)
	}
	fmt.Println(" pending: 0x%x\n", tx.Hash())
	return tx
}
//func balanceMiner (mw MinWallet.MinerWallet, wb common.Address, backend *bind.ContractBackend) () {
//	token,err := Token.NewSDT(wb,backend)
//		if err != nil {
//			log.Fatalf("Failed to instantiate a Token contract: %v", err)
//		}
//	bal, err := token.BalanceOf(&bind.CallOpts{Pending: true},wb)
//}

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

func CheckMiners (conn bind.ContractBackend, mb common.Address) (*types.Transaction, error){
	cm := GlueWhitelist(conn)
	tx, err := cm.RegistredMiners(&bind.CallOpts{Pending: true}, mb)

	if err != nil {
		log.Fatalf("Failed to retrieve miner wallet: %v", err)
	}
	return tx
}

func CheckHubs (conn bind.ContractBackend , mb common.Address) (*types.Transaction, error){
	//wRegisteredHubs
	ch := GlueWhitelist(conn)
	tx, err := ch.RegistredHubs(&bind.CallOpts{Pending: true}, mb)
	if err != nil {
		log.Fatalf("Failed to retrieve hubs wallet: %v", err)
	}
	return tx
}
