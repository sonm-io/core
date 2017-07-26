package main

import (
	"time"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/sonm-io/core/fusrodah"
)

const testTopic = "myPrivateTopic"

func main() {
	done := make(chan interface{})

	fus := fusrodah.NewServer(
		nil,
		":30345",
		"enode://aef138dce2082a56f52a4a2293a29d0a44db902abfe4fc8be6eb53c8f8cbc37927d0511e310746307903fee330198cd7f52e5eefec58a25e89627708a0279709@172.16.1.189:30348",
	)

	fus.Start()

	var from *ecdsa.PublicKey = fus.GetMsgPublicKey()
	to := fus.GetMsgPublicKey()

	fus.AddHandling(to, from, func(msg *whisperv2.Message) {
		payload := string(msg.Payload)
		log.Info("Incoming message: ", "body", payload)
		if payload == "+exit" {
			close(done)
		}
	}, testTopic)

	log.Info("Node started")

	go func() {
		time.Sleep(100 * time.Second)
		log.Debug("FIRE EXIT COMMAND")
		fus.SendPrivateMsg("+exit", fus.GetMsgPublicKey(), testTopic)
	}()

	for {
		select {
		case <-done:
			log.Debug("Got exit signal")
			return
		default:
			time.Sleep(3 * time.Second)
			fus.SendPrivateMsg("Hello", fus.GetMsgPublicKey(), testTopic)

		}
	}
}
