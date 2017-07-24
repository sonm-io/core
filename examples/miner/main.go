package main

import (
	"github.com/sonm-io/fusrodah/miner"
	//"github.com/ethereum/go-ethereum/crypto"
	"fmt"
	"time"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	//prv, _ := crypto.GenerateKey()
	var ip string
	var filterId int
	srv := miner.NewServer(nil)

	srv.Start()

	//var hubPubKeyString *ecdsa.PublicKey
	c := make(chan bool, 1)
	var hubRecievedPubKey *ecdsa.PublicKey


	filterId = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		hubRecievedPubKey = crypto.ToECDSAPub(msg.Payload)

		fmt.Println(string(msg.Payload))

		srv.Frd.RemoveHandling(filterId)

		filterId = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
			ip = string(msg.Payload)
			fmt.Println(ip)

			c <- true
		}, "minerAddr")

		i := 0
		for i < 10{
			i += 1
			srv.Frd.Send(srv.GetPubKeyString(), nil, true,"hubAddr")
			fmt.Println("DISC #2 SENDED")
			time.Sleep(time.Millisecond * 1000)
		}

	}, "minerDiscover")

	//for {
	//fmt.Println(srv.GetPubKeyString())
	i := 0
	for i < 10{
		i += 1
		srv.Frd.Send(srv.GetPubKeyString(), nil, true, "hubDiscover")
		fmt.Println("DISC #1 SENDED")
		time.Sleep(time.Millisecond * 1000)
	}



	select {
	case <-c:

	}
}
