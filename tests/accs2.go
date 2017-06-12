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
	//"bytes"
	//"bufio"
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
	type PasswordJson struct {
		Password		string	`json:"Password"`
	}

	npass:="/pass.json"

	//Reading user password
	// ВОПРОС - Это возвращает JSON структуру или строку?
	func readPwd() PasswordJson{
		usr, err := user.Current();
		// User password file JSON should be in root of home directory
		file, err := ioutil.ReadFile(usr.HomeDir+npass)
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

	pass:=readPwd()
*/

	npass:="/pass.txt"

	hnpass:=hd+npass

	passf, err:=ioutil.ReadFile(hnpass)
	if err != nil {
        log.Fatalf("can't read the file", err)
    }


		pass:=string(passf)

		fmt.Println("password:")
		fmt.Println(pass)


		//pass = strings.NewReader(pass)



/*
	reader := bufio.NewReader(os.Stdin)
fmt.Print("Enter text: ")
text, _ := reader.ReadString('\n')
fmt.Println(text)

	text=string(text)
*/


	qwe:="Metamorph9"
	qwe=pass
	fmt.Println(qwe)

	auth, err := bind.NewTransactor(strings.NewReader(key), qwe)
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
