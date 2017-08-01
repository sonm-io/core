package main

import (
	"github.com/sonm-io/core/fusrodah/hub"
	"fmt"
)

func main() {
	srv, err := hub.NewServer(nil, "123.123.123.123")
	if err != nil {
		fmt.Errorf("Could not initialize server: %s", err)
		return
	}
	err = srv.Start()
	if err != nil {
		fmt.Errorf("Could not start server: %s", err)
		return
	}
	srv.Serve()
	select {}
}
