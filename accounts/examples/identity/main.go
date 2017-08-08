package main

import (
	"github.com/sonm-io/core/accounts"
	"fmt"
)

func main(){
	idt, err := accounts.NewIdentity("/absolute/path/to/Ethereum", "")
	if err != nil {
		fmt.Printf("Error during initialization: %s", err)
		return
	}

	fmt.Printf("PrivateKey: %s \r\n", idt.PrivateKey)
	fmt.Printf("TransactionOpts: %s \r\n", idt.TransactionOpts)
}
