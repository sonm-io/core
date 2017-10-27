package main

import (
	"fmt"

	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/insonmnia/locator"
)

var (
	listenAddr  = flag.String("addr", ":9090", "Locator service listen addr")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Locator %s\r\n", version)
		return
	}

	cfg := locator.DefaultLocatorConfig()
	cfg.ListenAddr = *listenAddr

	lc := locator.NewLocator(cfg)
	fmt.Printf("Starting locator service at %s...\r\n", cfg.ListenAddr)
	if err := lc.Serve(); err != nil {
		fmt.Printf("Cannot start Locator service: %s\r\n", err.Error())
	}
}
