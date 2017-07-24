package main

import (
	"github.com/sonm-io/fusrodah/miner"
	//"github.com/ethereum/go-ethereum/crypto"
	"fmt"
)

func main() {
	//prv, _ := crypto.GenerateKey()

	srv := miner.NewServer(nil)

	srv.Start()
	srv.Serve()
	ip := srv.GeHubIp()
	fmt.Println(ip)
}
