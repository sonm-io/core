package main


import (
	//"github.com/sokel/tsc/api"
	"github.com/sonm-io/core/accounts"
	"fmt"
	"log"
	"github.com/sonm-io/core/blockchain/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/sonm-io/core/blockchain/tsc/api"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain/tsc"
)

const testPass = ""
var token_contract = common.StringToAddress("0xfaf800cad91426f026db07d254461cc707d10aa0")

const ethEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"


func main() {
	var err error

	ks := accounts.NewIdentity(accounts.GetDefaultKeystoreDir())
	err = ks.New(testPass)
	if err != nil {
		log.Fatal(err)
	}

	err = ks.Open(testPass)
	if err != nil {
		log.Fatal(err)
	}

	prv, err := ks.GetPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(prv)
	fmt.Println(auth)

	client, err := utils.InitEthClient(ethEndpoint)
	if err != nil {
		return
	}
	fmt.Print(client)

	token, err := api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), client)

	fmt.Println(token)
	totalSupply, err := token.TotalSupply(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal("error via getting totalSupply(): ", err)
		return
	}
	fmt.Println(totalSupply)

}
