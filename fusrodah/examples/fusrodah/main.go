package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/fusrodah"
)

const testTopic = "testme"
const quitCommand = "Quit"

func main() {
	frd, _ := fusrodah.NewServer(nil, ":30345", []string{common.BootNodeAddr, common.SecondBootNodeAddr})
	frd.Start()

	done := make(chan struct{})
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
			frd.Send("healthz", true, testTopic)
		}
	}
}
