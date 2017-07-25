package metricsStructs

import (
	"time"
	"encoding/json"
	"io/ioutil"
)

// MetricsHub describe some
type MetricsHub struct {
	// records the hub address
	HubAddress          string        `json:"hubAddress"`
	//uses pings to diagnose the system and figure out the uplink speed
	// between the hub and the source. Determines whether there are any packets lost between the source and the hub.
	HubPing             string        `json:"hubPing"`
	//the amount of services, available to the hub (not yet implemented).
	HubService          string        `json:"hubService"`
	//the attribute determines how much are the participants holding in their wallets.
	HubStack            string        `json:"hubStack"`
	//date, on which the hub was activated (registered).
	// Determines for how long the hub was been registered (in order to assess the activity levels for that period).
	CreationDate        string        `json:"creationDate"`
	//this attribute sets the amount of money that the hub can pay out
	PayDay              string        `json:"payDay"`
	//this function sets the transfer limit for the hub.
	TransferLimit       string        `json:"transferLimit"`
	//determines the lifetime of the hub.
	HubLifetime         time.Time     `json:"hubLifeTime"`
	//this sttribute determines the response time for the hub, which in turn influences the activity probability for the hub.
	SpeedConfirm        time.Time     `json:"speedConfirm"`
	//the overall amount of time the hub spent being frozen.
	FreezeTime          time.Time     `json:"freezeTime"`
	//the limit on the amount of money that the hub can send.
	DayLimit            time.Time     `json:"dayLimit"`
	// how many times the hub was frozen.
	AmountFreezeTime    int           `json:"amountFreezeTime"`
	//this status becomes true if the hub is suspected to be involved with fraud.
	SuspectStatus       bool          `json:"suspectStatus"`
	//attribute that shows whether or not the hub has presale tokens.
	AvailabilityPresale bool          `json:"availabilityPresale"`
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

func (m *MetricsHub) LoadFromFile(path string) error {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = m.FromJSON(file)
	if err != nil {
		return err
	}

	return nil
}

func (m *MetricsHub) SaveToFile(path string) error {
	data := m.ToJSON()
	return ioutil.WriteFile(path, data, 0600)
}

