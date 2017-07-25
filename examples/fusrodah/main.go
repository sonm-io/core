package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/fusrodah/fusrodah"
)

const testTopic = "testme"

func main() {
	done := make(chan struct{})
	prv, _ := crypto.GenerateKey()

	frd := fusrodah.Fusrodah{
		Prv:  prv,
		Port: ":30345",
		// Enode is bootnode p2p addr
		Enode: "enode://12428dbfae0929ef05a6bfe2798db9ad830b9d65f74aea98285ffa105196ebd6b823ed4911fc5b65b3a4868b23f47f1e577e48b4b833f30cc1fbdaa0ff4d21bf@172.16.1.189:30348",
	}

	frd.Start()

	frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		fmt.Printf("Incoming message: %s\r\n", string(msg.Payload))
		if string(msg.Payload) == "Quit" {
			close(done)
		}
	}, testTopic)

	for {
		select {
		case <-done:
			fmt.Println("Quit command received, finishing...")
			return
		default:
			time.Sleep(3 * time.Second)
			frd.Send("healthz", nil, true, testTopic)
		}
	}
}
