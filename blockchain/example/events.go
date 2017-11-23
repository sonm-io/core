package main

import (
	"log"

	"github.com/sonm-io/core/blockchain"
)

func main() {
	bch, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	dealsIds, err := bch.GetOpenedDeal("", "")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("OpenedDeals: ", dealsIds)

	dealsIds, err = bch.GetAcceptedDeal("", "")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("AcceptedDeals: ", dealsIds)

	dealsIds, err = bch.GetClosedDeal("", "")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("ClosedDeals: ", dealsIds)
}
