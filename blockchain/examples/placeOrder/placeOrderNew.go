package main

import (
	"context"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/blockchain"
)

const (
	hexKey = "a000000000000000000000000000000000000000000000000000000000000000"
)

func main() {
	client, err := blockchain.NewClient("https://sidechain-dev.sonm.com")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	prv, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Fatalln(err)
		return
	}

	api, err := blockchain.NewAPI()
	if err != nil {
		log.Fatalln(err)
		return
	}

	tx, err := api.SideToken().Approve(context.TODO(), prv, "0x0", big.NewInt(0))
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	rec, err := blockchain.WaitTransactionReceipt(context.TODO(), client, time.Duration(3*time.Second), tx)
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	log.Println(rec)
}
