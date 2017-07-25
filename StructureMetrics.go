package metricsStructs

import (
	"time"
	"encoding/json"
	"io/ioutil"
)

// MetricsHub describe some
type MetricsHub struct {
	HubAddress          string        `json:"hubAddress"`
	HubPing             string        `json:"hubPing"`
	HubService          string        `json:"hubService"`
	HubStack            string        `json:"hubStack"`
	CreationDate        string        `json:"creationDate"`
	PayDay              string        `json:"payDay"`
	TransferLimit       string        `json:"transferLimit"`
	HubLifetime         time.Time     `json:"hubLifeTime"`
	SpeedConfirm        time.Time     `json:"speedConfirm"`
	FreezeTime          time.Time     `json:"freezeTime"`
	DayLimit            time.Time     `json:"dayLimit"`
	AmountFreezeTime    int           `json:"amountFreezeTime"`
	SuspectStatus       bool          `json:"suspectStatus"`
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

//тест для формирования структуры: пишу файл- проверяю,чтобы не было ош,
// сравнивать структуру по полям
