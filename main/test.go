package main

import (
	"github.com/sonm-io/metricsStructs"
	"fmt"
)

/**
 /--------TEST--------/
 THIS FUNCTION FOR TEST
 /--------------------/
*/
func main() {
	v := metricsStructs.MetricsHub{"hubAddress", "HubPing", "HubService", "HubStack", "date"}
	v.HubAddress = "CurrentlyHub address"
	fmt.Print(v.HubAddress)
}
