package main
//
//
//import (
//	"github.com/sonm-io/Fusrodah"
//	"github.com/sonm-io/go-ethereum/whisper/whisperv2"
//	"github.com/sonm-io/go-ethereum/crypto"
//	"fmt"
//)
//
//func main(){
//	done := make(chan struct{})
//	key := Fusrodah.Key{}
//	if !key.Load(){
//		key.Generate()
//		key.Save()
//	}
//
//	frd := Fusrodah.Fusrodah{Prv: key.Prv}
//	frd.Start("8000")
//
//	id1 := frd.AddHandling(nil, func(msg *whisperv2.Message) {
//		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
//
//		frd.Send(key.GetPubKeyString(), receivedPubKey, "miner", "discover")
//	}, "hub", "discover")
//
//	id2:= frd.AddHandling(&key.Prv.PublicKey, func(msg *whisperv2.Message) {
//		receivedPubKey := crypto.ToECDSAPub(msg.Payload)
//		frd.Send(Fusrodah.GetLocalIP(), receivedPubKey, "miner", "addr")
//	}, "hub", "addr")
//
//	fmt.Println(string(id1)+":::"+string(id2))
//
//	for {
//		close(done)
//	}
//	select {
//		case <- done:
//	}
//}