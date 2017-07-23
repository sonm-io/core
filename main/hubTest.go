package main

import (
	"github.com/sonm-io/Fusrodah/hub"
	"github.com/sonm-io/go-ethereum/crypto"
	"fmt"
	"github.com/sonm-io/Fusrodah"
	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
)

func main(){
	done := make(chan struct{})

	//prv, _ := crypto.GenerateKey()

	srv := hub.NewServer(nil, "123.123.123.123")

	srv.Frd.Send("", nil, "hubDiscover")

	srv.Start()
	srv.Frd.AddHandling(nil, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #1")
		srv.Frd.Send(srv.GetPubKeyString(), receivedPubKey, "miner", "discover")
	}, "hubDiscover")

	srv.Frd.AddHandling(&srv.PrivateKey.PublicKey, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #2")
		srv.Frd.Send(Fusrodah.GetLocalIP(), receivedPubKey, "miner", "addr")
	}, "hub", "addr")


	select {
	case <-done:

	}
}