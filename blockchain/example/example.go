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
	"math/big"
)

const testPass = "QWEpoi123098"
var token_contract = common.StringToAddress("0xfaf800cad91426f026db07d254461cc707d10aa0")

const ethEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"


func main() {
	var err error

	ks := accounts.NewIdentity(accounts.GetDefaultKeystoreDir())

	err = ks.Open(testPass)
	if err != nil {
		if err == accounts.ErrWalletIsEmpty{
			err = ks.New(testPass)
			if err != nil {
				log.Fatal(err)
			}
		}
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
	fmt.Println(client)

	//token, err := api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), client)
	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), client)

	//fmt.Println(token)
	//totalSupply, err := token.TotalSupply(&bind.CallOpts{Pending: true})
	//if err != nil {
	//	log.Fatal("error via getting totalSupply(): ", err)
	//	return
	//}
	//fmt.Println("token Supply: ", totalSupply)

	auth.GasLimit = big.NewInt(200000)
	auth.GasPrice = big.NewInt(20000000000) // 20 gWei

	//tx, err := token.Approve(auth, common.HexToAddress("0x41BA7e0e1e661f7114f2F05AFd2536210c2ED351"), big.NewInt(1000000))
	tx, err := dealsContract.OpenDeal(auth, common.HexToAddress("0x41BA7e0e1e661f7114f2F05AFd2536210c2ED351"), big.NewInt(2341234), big.NewInt(10000))

	if err != nil {
		log.Fatalln(err)
	}
	log.Println(tx)
}
