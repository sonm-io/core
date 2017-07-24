package main

import (
	//"fmt"
	"github.com/sonm-io/fusrodah/hub"
	"github.com/sonm-io/fusrodah/util"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"fmt"
)

func main() {
	done := make(chan struct{})

	//prv, _ := crypto.GenerateKey()

	srv := hub.NewServer(nil, "123.123.123.123")


	srv.Start()


	//srv.Serve()
	srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #1")
		srv.Frd.Send(srv.GetPubKeyString(), receivedPubKey, "minerDiscover")
	}, "hubDiscover")

	srv.Frd.AddHandling(&srv.PrivateKey.PublicKey, nil, func(msg *whisperv2.Message) {
		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
		fmt.Println("DISCOVERY RESPONSE #2")
		srv.Frd.Send(util.GetLocalIP(), receivedPubKey, "miner", "addr")
	}, "hub", "addr")

	srv.Frd.Send("", nil, "hubDiscover")


	select {
	case <-done:

	}
}
