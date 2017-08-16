package main

import (
	"fmt"
	"github.com/sonm-io/core/accounts"
)

func main() {
	var err error

	keydir := accounts.GetDefaultKeystoreDir()

	idt := accounts.NewIdentity(keydir)

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
