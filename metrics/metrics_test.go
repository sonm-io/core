package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testFile = "config_test.json"

func TestMetricsHubFileOperations(t *testing.T) {
	os.Remove(testFile)

	loc, _ := time.LoadLocation("UTC")
	now := time.Now().In(loc)

	m := &HubMetrics{
		HubAddress:          "123",
		HubPing:             "q23",
		HubService:          "service",
		HubStack:            "hub stack",
		CreationDate:        "12.12.12",
		PayDay:              "12.14.2012",
		TransferLimit:       "1234.42",
		HubLifetime:         now,
		SpeedConfirm:        now,
		FreezeTime:          now,
		DayLimit:            now,
		AmountFreezeTime:    3,
		SuspectStatus:       true,
		AvailabilityPresale: true,
	}

	err := m.SaveToFile(testFile)
	assert.Nil(t, err)

	m2 := &HubMetrics{}
	err = m2.LoadFromFile(testFile)
	assert.Nil(t, err)

	assert.Equal(t, m.HubPing, m2.HubPing)
	assert.Equal(t, m.HubAddress, m2.HubAddress)
	assert.Equal(t, m.HubService, m2.HubService)
	assert.Equal(t, m.HubStack, m2.HubStack)
	assert.Equal(t, m.CreationDate, m2.CreationDate)
	assert.Equal(t, m.PayDay, m2.PayDay)
	assert.Equal(t, m.TransferLimit, m2.TransferLimit)
	assert.Equal(t, m.HubLifetime, m2.HubLifetime)
	assert.Equal(t, m.SpeedConfirm, m2.SpeedConfirm)
	assert.Equal(t, m.FreezeTime, m2.FreezeTime)
	assert.Equal(t, m.DayLimit, m2.DayLimit)
	assert.Equal(t, m.AmountFreezeTime, m2.AmountFreezeTime)
	assert.Equal(t, m.SuspectStatus, m2.SuspectStatus)
	assert.Equal(t, m.AvailabilityPresale, m2.AvailabilityPresale)
}
