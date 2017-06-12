package main
import (
	"log"
	//"github.com/sonm-io/blockchain-api"
	"github.com/sonm-io/go-ethereum/ethclient"
	"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	"strings"
	"io/ioutil"
	//"github.com/sonm-io/go-ethereum/accounts/keystore"
	//"github.com/sonm-io/go-ethereum/accounts"
	"github.com/sonm-io/go-ethereum/common"
	//"time"
	"math/big"
	"fmt"
	//"os"
	"os/user"
	//"github.com/sonm-io/blockchain-api/go-build/MinWallet"
	"github.com/sonm-io/blockchain-api/go-build/SDT"
	//"github.com/sonm-io/go-ethereum/core/types"
)
func main (){
	//-------------------- 1 ------------------------------


/*
	usr, err := accounts.Accounts();
	fmt.Println(usr)
*/

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
	//return first





/*
	keyin := strings.NewReader(key)
	passphrase := "Metamorph9"

	json, err := ioutil.ReadAll(keyin)
		if err != nil {
			//return nil, err
			log.Fatalf("failed json", err)
		}
	key, err := keystore.DecryptKey(json, passphrase)
		if err != nil {
			//return nil, err
			log.Fatalf("failed key", err)
		}
*/

	//key=string(key)


	auth, err := bind.NewTransactor(strings.NewReader(key), "Metamorph9")
		if err != nil {
			log.Fatalf("Failed to create authorized transactor: %v", err)
		}

	//Create connection
	conn, err := ethclient.Dial("/home/jack/.rinkeby/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

		token, err := Token.NewSDT(common.HexToAddress("0x09e4a2de83220c6f92dcfdbaa8d22fe2a4a45943"), conn)
		if err != nil {
			log.Fatalf("Failed to instantiate a Token contract: %v", err)
		}
		name, err := token.Name(nil)
		if err != nil {
			log.Fatalf("Failed to retrieve token name: %v", err)
		}
		fmt.Println("Token name:", name)


		tx, err := token.Transfer(auth, common.HexToAddress("0x0000000000000000000000000000000000000000"), big.NewInt(1))

		if err != nil {
			log.Fatalf("Failed to request token transfer: %v", err)
		}
		// Need to do something about checking pending tx
		fmt.Printf("Transfer pending: 0x%x\n", tx.Hash())

}
