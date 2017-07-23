package main

import (
	"github.com/sonm-io/Fusrodah/mainer"
	//"github.com/sonm-io/go-ethereum/crypto"
	"fmt"
)

func main(){
	//prv, _ := crypto.GenerateKey()

	srv := mainer.NewServer(nil)

	srv.Start()
	srv.Serve()
	ip := srv.GeHubIp()
	fmt.Println(ip)
}
