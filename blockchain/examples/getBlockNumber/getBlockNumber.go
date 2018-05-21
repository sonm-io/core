package main

import (
	"context"
	"log"
	"os"

	"github.com/sonm-io/core/blockchain"
)

func main() {
	client, err := blockchain.NewClient("https://sidechain-dev.sonm.com")
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	client.GetLastBlock(context.TODO())
}
