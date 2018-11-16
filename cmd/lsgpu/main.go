package main

import (
	"fmt"
	"os"

	"github.com/sonm-io/core/insonmnia/worker/gpu"
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
		if m, err := card.Metrics(); err == nil {
			fmt.Printf(" t = %.1f (fan = %.1f%%)\r\n", m.Temperature, m.Fan)
			fmt.Printf(" pow = %.1f\r\n", m.Power)
		} else {
			fmt.Printf(" metrics is not available\r\n")
		}

		fmt.Printf("   vid=%d did=%d\r\n", card.VendorID, card.DeviceID)
		fmt.Printf("   maj=%d min=%d\r\n", card.Major, card.Minor)
		fmt.Printf("   PCI=%s\r\n", card.PCIBusID)

		for _, rel := range card.Devices {
			fmt.Printf("      %s\r\n", rel)
		}
	}
}
