package main

import (
	"fmt"
	"github.com/sonm-io/core/fusrodah/hub"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	prv, _ := crypto.GenerateKey()

	srv, err := hub.NewServer(prv, "123.123.123.123")
	if err != nil {
		fmt.Printf("Could not initialize server: %s\r\n", err)
		return
	}
	err = srv.Start()
	if err != nil {
		fmt.Printf("Could not start server: %s\r\n", err)
		return
	}
	srv.Serve()
	select {}
}
