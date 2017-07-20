package metricsStructs

import (
	"time"
	//"fmt"

)

type MetricsHub struct {
	HubAddress string `json:"hubAddress"`
	HubPing string `json: "hubPing"`
	HubService string `json: "hubService"`
	HubStack string `json: "hubStack"`
	CreationDate string `json: "creationDate"`

	//HubLifetime time.Time
	//PayDay string
	//FreezeTime time.Time
	//AmountFreezeTime int
	//TransferLimit string
	//SuspectStatus bool
	//DayLimit time.Time
	//AvailabilityPresale bool
	//SpeedConfirm time.Time
}

type MetricsMin struct {
	MinAddress string
	MinPing string
	MinStack string
	CreationDate time.Time
	MinService string
}

func (metricsHub *MetricsHub) StructureMetrics  () {
	//v := MetricsHub{"hubAddress", "HubPing", "HubService", "HubStack", time.Now()}
	//v.HubAddress = "CurrentlyHub address"
	//fmt.Print(v.HubAddress)
}
