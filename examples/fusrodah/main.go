package main

import (
	"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
)

func main() {
	done := make(chan struct{})
	//prv, _ := crypto.GenerateKey()

	frd := fusrodah.Fusrodah{
		Prv:   nil,
		Port:  ":30346",
		Enode: "enode://8614cbcc79eaede3f26731c0002e768f15adbcdb5f7dab2961536959d36e580fd9183075dc89a3732805f4ce83a9bfb0612f5bc9ad61c01beebad0dea52dd4f9@192.168.10.51:30348",
	}

	frd.Start()

	frd.AddHandling(nil, func(msg *whisperv2.Message) {
		fmt.Println(string(msg.Payload))
		if string(msg.Payload) == "Quit" {
			close(done)
		}
	}, "test")

	//frd.Send("Quit", nil, "test")

	select {
	case <-done:

	}
}
