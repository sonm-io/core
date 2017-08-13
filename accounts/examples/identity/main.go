package main

import (
	"fmt"
	"github.com/sonm-io/core/accounts"
)

func main() {
	idt := accounts.NewIdentity()

	keydir := "/abs/path/to/eth/keystore"
	err := idt.Load(&keydir)
	if err != nil {
		fmt.Printf("Error during loading: %s", err)
		return
	}

	err = idt.Open("")
	if err != nil {
		fmt.Printf("Error during opening: %s", err)
		return
	}

	prv, err := idt.GetPrivateKey()
	if err != nil {
		fmt.Printf("Error during getting key: %s", err)
		return
	}
	fmt.Printf("PrivateKey: %s \r\n", prv)

}
