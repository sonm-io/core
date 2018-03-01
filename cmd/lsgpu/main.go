package main

import (
	"fmt"
	"os"

	"github.com/sonm-io/core/insonmnia/miner/gpu"
)

var appVersion string

func main() {
	fmt.Printf("sonm lspgu %s\r\n", appVersion)
	cards, err := gpu.CollectDRICardDevices()
	if err != nil {
		fmt.Printf("cannot collect card devces: %v\r\n", err)
		os.Exit(1)
	}

	for _, card := range cards {
		fmt.Printf("Card: %s\r\n", card.Path)

		fmt.Printf("   vid=%d did=%d\r\n", card.VendorID, card.DeviceID)
		fmt.Printf("   maj=%d min=%d\r\n", card.Major, card.Minor)
		fmt.Printf("   PCI=%s\r\n", card.PCIBusID)

		for _, rel := range card.Devices {
			fmt.Printf("      %s\r\n", rel)
		}
	}
}
