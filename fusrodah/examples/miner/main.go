package main

import (
	"fmt"

	"github.com/sonm-io/core/fusrodah/miner"
	"time"
)

func main() {

	srv := miner.NewServer(nil)
	srv.Start()
	srv.Serve()

	for{
		select {
		default:
			time.Sleep(1*time.Second)
			fmt.Println(srv.GetHubIp())
		}
	}
}
