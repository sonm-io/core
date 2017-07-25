package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/fusrodah/common"
	"github.com/sonm-io/fusrodah/fusrodah"
)

const testTopic = "testme"
const quitCommand = "Quit"

func main() {
	done := make(chan struct{})
	prv, _ := crypto.GenerateKey()

	frd := fusrodah.Fusrodah{
		Prv:  prv,
		Port: ":30345",
		// Enode is bootnode p2p addr
		Enode: common.BootNodeAddr,
	}

	frd.Start()

	frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		fmt.Printf("Incoming message: %s\r\n", string(msg.Payload))
		if string(msg.Payload) == quitCommand {
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
