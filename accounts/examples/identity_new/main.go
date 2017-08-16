package main

import (
	"fmt"
	"github.com/sonm-io/core/accounts"
)

func main() {
	var err error

	keydir := accounts.GetDefaultKeystoreDir()
	pass := ""

	idt := accounts.NewIdentity(keydir)

	err = idt.New(pass)
	if err != nil {
		fmt.Printf("Error during creating new account: %s", err)
		return
	}

	err = idt.Open(pass)
	if err != nil {
		fmt.Printf("Error during open: %s", err)
	}

	prv, err := idt.GetPrivateKey()
	if err != nil {
		fmt.Printf("Error during getting key: %s", err)
		return
	}
	fmt.Printf("PrivateKey: %s \r\n", prv)
}
