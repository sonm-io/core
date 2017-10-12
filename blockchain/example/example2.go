package main

import (
	"github.com/sonm-io/core/blockchain"
	"log"
	"fmt"
)

func main(){

	hubDeals, err := blockchain.GetHubDeals("0xc533ca655314ff1407312637fb3b82000aa1e154")
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println("HubDeals:", hubDeals)

	clientDeals, err := blockchain.GetClientDeals("0xc533ca655314ff1407312637fb3b82000aa1e154")
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
