package main

import (
	"fmt"

	"github.com/sonm-io/fusrodah/miner"
)

func main() {

	srv := miner.NewServer(nil)
	srv.Start()
	srv.Serve()

	ip := srv.GetHubIp()
	fmt.Println(ip)
}
