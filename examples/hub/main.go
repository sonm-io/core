package main

import (
	"github.com/sonm-io/fusrodah/hub"
)

func main() {

	srv := hub.NewServer(nil, "123.123.123.123")

	srv.Start()

	srv.Serve()

	select {}
}
