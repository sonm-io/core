package main

import (
	"fmt"
	"time"
	"github.com/sonm-io/fusrodah/miner"
	"github.com/ethereum/go-ethereum/whisper/whisperv2"
)

func main(){

	srv := miner.NewServer(nil)
	srv.Start()
	srv.Serve()

	var ip string
	ip = srv.GetHubIp()
	fmt.Println(ip)
}

func mainOld() {
	var ip string
	var filterId int

	srv := miner.NewServer(nil)
	srv.Start()

	done := make(chan bool, 1)

	filterId = srv.Frd.AddHandling(nil, nil, func(msg *whisperv2.Message) {
		ip = string(msg.Payload)

		srv.Frd.RemoveHandling(filterId)
		close(done)
	}, "minerDiscover")

	for {
		select {
		case <-done:
			fmt.Printf("Discovery complete, hub ip = %s\r\n", ip)
			return
		default:
			srv.Frd.Send(srv.GetPubKeyString(), nil, true, "hubDiscover")
			time.Sleep(time.Millisecond * 1000)
		}
	}
}
