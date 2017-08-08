package metrics

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// HubMetrics describe some
type HubMetrics struct {
	// HubAddress records the hub address
	HubAddress string `json:"hubAddress"`
	// HubPing uses pings to diagnose the system and figure out the uplink speed
	// between the hub and the source. Determines whether there are any packets lost between the source and the hub.
	HubPing string `json:"hubPing"`
	// HubService: the amount of services, available to the hub (not yet implemented).
	HubService string `json:"hubService"`
	// HubStack: how much are the participants holding in their wallets.
	HubStack string `json:"hubStack"`
	// CreationDate: date, on which the hub was activated (registered).
	// Determines for how long the hub was been registered (in order to assess the activity levels for that period).
	CreationDate time.Time `json:"creationDate"`
	// PayDay sets the amount of money that the hub can pay out
	PayDay float64 `json:"payDay"`
	// TransferLimit sets the transfer limit for the hub.
	TransferLimit float64 `json:"transferLimit"`
	// HubLifetime determines the lifetime of the hub.
	HubLifetime time.Duration `json:"hubLifeTime"`
	// SpeedConfirm attribute determines the response time for the hub, which in turn influences the activity probability for the hub.
	SpeedConfirm time.Time `json:"speedConfirm"`
	// FreezeTime the overall amount of time the hub spent being frozen.
	FreezeTime time.Time `json:"freezeTime"`
	// DayLimit the limit on the amount of money that the hub can send.
	DayLimit time.Time `json:"dayLimit"`
	// AmountFreezeTime how many times the hub was frozen.
	AmountFreezeTime int `json:"amountFreezeTime"`
	// SuspectStatus this status becomes true if the hub is suspected to be involved with fraud.
	SuspectStatus bool `json:"suspectStatus"`
	// AvailabilityPresale attribute that shows whether or not the hub has presale tokens.
	AvailabilityPresale bool `json:"availabilityPresale"`
}

// ToJSON serialise struct
func (m *HubMetrics) ToJSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

//FromJSON de_serialise struct
func (m *HubMetrics) FromJSON(b []byte) error {
	err := json.Unmarshal(b, &m)
	return err
}

func (m *HubMetrics) LoadFromFile(path string) error {
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

func (m *HubMetrics) SaveToFile(path string) error {
	data := m.ToJSON()
	return ioutil.WriteFile(path, data, 0600)
}

type MinerMetrics struct {
	MinAddress   string
	MinPing      string
	MinStack     string
	CreationDate time.Time
	MinService   string
}
