package metricsStructs

import (
	"time"
	"encoding/json"
)

// MetricsHub describe some
type MetricsHub struct {
	HubAddress          string        `json:"hubAddress"`
	HubPing             string        `json:"hubPing"`
	HubService          string        `json:"hubService"`
	HubStack            string        `json:"hubStack"`
	CreationDate        string        `json:"creationDate"`
	HubLifetime         time.Time     `json:"hubLifeTime"`
	PayDay              string        `json:"payDay"`
	FreezeTime          time.Time     `json:"freezeTime"`
	AmountFreezeTime    int 		  `json:"amountFreezeTime"`
	TransferLimit       string 		  `json:"transferLimit"`
	SuspectStatus       bool 		  `json:"suspectStatus"`
	DayLimit            time.Time 	  `json:"dayLimit"`
	AvailabilityPresale bool 		  `json:"availabilityPresale"`
	SpeedConfirm        time.Time 	  `json:"speedConfirm"`
}
type MetricsMiner struct {
	MinAddress   string
	MinPing      string
	MinStack     string
	CreationDate time.Time
	MinService   string
}

// ToJSON serialise struct
func (m *MetricsHub) ToJSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

//FromJSON de_serialise struct
func (m *MetricsHub) FromJSON(b []byte) error {
	err := json.Unmarshal(b, &m)
	return err
}

//func (m *MetricsHub) MarshalJson() ([]byte, error) {
//	return []byte (`{"hubAddress": "` + m.HubAddress + `"}`), nil
//}
