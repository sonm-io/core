package main

import (
	"fmt"
	"github.com/sonm-io/core/fusrodah/hub"
)

func main() {
	srv, err := hub.NewServer(nil, "123.123.123.123")
	if err != nil {
		fmt.Printf("Could not initialize server: %s\r\n", err)
		return
	}
	err = srv.Start()
	if err != nil {
		fmt.Printf("Could not start server: %s\r\n", err)
		return
	}
	srv.Serve()
	select {}
}
