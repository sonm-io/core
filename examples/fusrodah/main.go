package main

import (
	"fmt"
	"github.com/sonm-io/fusrodah/fusrodah"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"time"
)

func main() {
	done := make(chan struct{})
	//prv, _ := crypto.GenerateKey()

	frd := fusrodah.Fusrodah{
		Prv:   nil,
		Port:  ":30346",
		Enode: "enode://bda98fd8e7b8a377f8964d98ac71f5b2f8df0c8401dc62437905deca1b71a582aa6a3c57e3d3b4092e3d444f9b751bd54abaa536dfdf057ed8f31684bdac19b2@10.196.131.151:30348",
	}

	frd.Start()

	frd.AddHandling(nil, func(msg *whisperv2.Message) {
		fmt.Println(string(msg.Payload))
		if string(msg.Payload) == "Quit" {
			close(done)
		}
	}, "test")

	for{
		time.Sleep(3*time.Second)
		//frd.Send("Quit", nil, "test")
	}

	select {
	case <-done:

	}
}
