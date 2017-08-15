package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/fusrodah/miner"
)

func main() {
	prv, _ := crypto.GenerateKey()

	srv, err := miner.NewServer(prv)
	if err != nil {
		fmt.Printf("Error while initialize instanse: %s \r\n", err)
		return
	}
	err = srv.Start()
	if err != nil {
		fmt.Printf("Error while start instanse: %s \r\n", err)
		return
	}

	srv.Serve()

	fmt.Println(srv.GetHub())
}
