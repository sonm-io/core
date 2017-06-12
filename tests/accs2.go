package main
import (
	"log"
	//"github.com/sonm-io/blockchain-api"
	//"github.com/sonm-io/go-ethereum/ethclient"
	//"github.com/sonm-io/go-ethereum/accounts/abi/bind"
	//"strings"
	"io/ioutil"
	//"github.com/sonm-io/go-ethereum/accounts/keystore"
	//"github.com/sonm-io/go-ethereum/accounts"
	//"github.com/sonm-io/go-ethereum/common"
	//"time"
	//"math/big"
	"fmt"
	//"os"
	"os/user"
	//"github.com/sonm-io/blockchain-api/go-build/MinWallet"
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
	fmt.Println("key")
	fmt.Println(first)
	//return first




/*
	dir, err := os.Open(confPath)
	 if err != nil {
		 		log.Fatalf("can't open path", err)
			 return

	 }
	 defer dir.Close()

	 fileInfos, err := dir.Readdir(-1)
	 if err != nil {
		 	log.Fatalf("can't read dir", err)
			 return
	 }
	 for _, fi := range fileInfos {
			 fmt.Println(fi.Name())
			 first:=fi[0]
	 }

	 first:=fi[0]
	 fmt.Println(first)
*/

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

/*
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

}
