package sonm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPUDevice_HashID(t *testing.T) {
	dev := &GPUDevice{
		ID:          "PCI:0001:0",
		VendorID:    100,
		VendorName:  "vendor",
		DeviceID:    200,
		DeviceName:  "name",
		MajorNumber: 123,
		MinorNumber: 234,
		Memory:      100500,
		Hash:        "123",
	}

	dev.FillHashID()
	v1 := dev.GetHash()

	dev.FillHashID()
	v2 := dev.GetHash()
	assert.Equal(t, v1, v2)

	dev.Hash = "wtfisthis"
	dev.FillHashID()
	v3 := dev.GetHash()

	assert.Equal(t, v2, v3)
}
