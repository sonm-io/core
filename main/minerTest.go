package main
//
//import (
//	"github.com/sonm-io/Fusrodah"
//	"crypto/ecdsa"
//	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
//	"github.com/sonm-io/go-ethereum/crypto"
//	"fmt"
//)
//
//func main() {
//	var address string
//
//	c := make(chan bool, 1)
//
//	key := Fusrodah.Key{}
//	if !key.Load(){
//		key.Generate()
//		key.Save()
//	}
//	frd := Fusrodah.Fusrodah{Prv: key.Prv}
//	frd.Start("8001")
//
//	var hubPubKeyString *ecdsa.PublicKey
//
//	go func() {
//
//		defer frd.Send(key.GetPubKeyString(), nil, "hub", "discover")
//		frd.AddHandling(nil, func(msg *whisperv2.Message) {
//			hubPubKeyString = crypto.ToECDSAPub(msg.Payload)
//			c <- true
//		}, "miner", "discover")
//
//	}()
//
//	<- c
//
//	go func() {
//
//		defer frd.Send(key.GetPubKeyString(), hubPubKeyString, "hub", "addr")
//		frd.AddHandling(&key.Prv.PublicKey, func(msg *whisperv2.Message) {
//			address = string(msg.Payload)
//		}, "miner", "addr")
//		c <- true
//	}()
//
//	<- c
//
//	fmt.Println(address)
//	frd.Stop()
//}
