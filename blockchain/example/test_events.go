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
	dealsIds, err := bch.GetOpenedDeal(nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(dealsIds)
}
