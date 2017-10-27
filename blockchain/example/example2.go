package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/blockchain/tsc/api"
	"github.com/sonm-io/core/blockchain/utils"
	"log"
)

func main() {

	const ethEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

	client, err := utils.InitEthClient(ethEndpoint)
	if err != nil {
		return
	}

	token, err := api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), client)

	//fmt.Println(token)
	totalSupply, err := token.TotalSupply(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatal("error via getting totalSupply(): ", err)
		return
	}
	fmt.Println("token Supply: ", totalSupply)

	balance, err := token.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress("0x41BA7e0e1e661f7114f2F05AFd2536210c2ED351"))

	fmt.Println("BALANCE:", balance)

	// DEALS EX

	hubDeals, err := blockchain.GetHubDeals("0x41BA7e0e1e661f7114f2F05AFd2536210c2ED351")
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println("HubDeals:", hubDeals)

	clientDeals, err := blockchain.GetClientDeals("0x41BA7e0e1e661f7114f2F05AFd2536210c2ED351")
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println("ClientDeals:", clientDeals)

	amount, err := blockchain.GetDealAmount()
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println("DealAmount:", amount)
}
