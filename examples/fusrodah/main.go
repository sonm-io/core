package main

import (
	"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"time"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	done := make(chan struct{})
	prv, _ := crypto.GenerateKey()

	frd := fusrodah.Fusrodah{
		Prv:   prv,
		Port:  ":30345",
		Enode: "enode://b0605764bd7c6a816c51325a9cb9d414277d639f420f9dc48b20d12c04c33391b0a99cc8c045d7ba4657de0c04e8bb3b0d4b072ca9779167a75761d7c3c18eb0@10.196.131.151:30348",
	}

	frd.Start()

	frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		fmt.Println(string(msg.Payload))
		if string(msg.Payload) == "Quit" {
			close(done)
		}
	}, "test")

	for{
		time.Sleep(3*time.Second)
		frd.Send("Quit", nil, true, "test")
	}

	select {
	case <-done:

	}
}
