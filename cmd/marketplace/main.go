package main

import (
	"fmt"

	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/insonmnia/marketplace"
)

var (
	listenAddr  = flag.String("addr", ":9095", "Marketplace service listen address")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Marketplace %s\r\n", version)
		return
	}

	mp := marketplace.NewMarketplace(*listenAddr)
	fmt.Printf("Starting Marketplace service at %s...\r\n", *listenAddr)
	if err := mp.Serve(); err != nil {
		fmt.Printf("Cannot start Markerplace service: %s\r\n", err)
	}
}
