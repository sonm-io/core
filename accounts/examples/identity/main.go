package main

import (
	"fmt"
	"github.com/sonm-io/core/accounts"
)

func main() {
	keydir := accounts.GetDefaultKeystoreDir()

	idt, err := accounts.NewIdentity(keydir)

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
