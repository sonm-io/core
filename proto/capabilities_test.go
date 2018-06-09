package sonm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNetFlags_ToBoolSlice(t *testing.T) {
	slice := make([]bool, MinNetFlagsCount)
	netFlags := &NetFlags{}
	require.Equal(t, netFlags.ToBoolSlice(), slice)

	netFlags.Flags = 6
	slice[1] = true
	slice[2] = true
	require.Equal(t, netFlags.ToBoolSlice(), slice)

}

func TestNetFlags_FromBoolSlice(t *testing.T) {
	slice := make([]bool, MinNetFlagsCount)
	slice[1] = true
	slice[2] = true

	require.Equal(t, NetFlagsFromBoolSlice(slice).Flags, uint64(6))
}

func TestNetFlags_GetFlags(t *testing.T) {
	netFlags := &NetFlags{Flags: 7}
	require.Equal(t, netFlags.GetOutbound(), true)
	require.Equal(t, netFlags.GetIncoming(), true)
	require.Equal(t, netFlags.GetOverlay(), true)
	netFlags = &NetFlags{Flags: 0}
	require.Equal(t, netFlags.GetOutbound(), false)
	require.Equal(t, netFlags.GetIncoming(), false)
	require.Equal(t, netFlags.GetOverlay(), false)
	netFlags = &NetFlags{Flags: 1}
	require.Equal(t, netFlags.GetOutbound(), false)
	require.Equal(t, netFlags.GetIncoming(), false)
	require.Equal(t, netFlags.GetOverlay(), true)
}

func TestNetFlags_SetFlags(t *testing.T) {
	netFlags := &NetFlags{}
	require.Equal(t, netFlags.SetOutbound(true).GetOutbound(), true)
	require.Equal(t, netFlags.SetOverlay(true).GetOverlay(), true)
	require.Equal(t, netFlags.SetIncoming(true).GetIncoming(), true)
	require.Equal(t, netFlags.SetOutbound(false).GetOutbound(), false)
	require.Equal(t, netFlags.SetOverlay(false).GetOverlay(), false)
	require.Equal(t, netFlags.SetIncoming(false).GetIncoming(), false)
}

func TestNetFlags_ConverseImplication(t *testing.T) {
	tests := []struct {
		lhs    uint64
		rhs    uint64
		result bool
	}{
		{7, 7, true},
		{7, 5, true},
		{7, 3, true},
		{7, 0, true},
		{0, 0, true},
		{6, 4, true},
		{6, 2, true},
		{5, 1, true},
		{4, 4, true},
		{4, 0, true},
		{0, 7, false},
		{0, 6, false},
		{0, 1, false},
		{2, 4, false},
		{2, 6, false},
		{5, 6, false},
	}

	for _, test := range tests {
		lhs := &NetFlags{Flags: test.lhs}
		rhs := &NetFlags{Flags: test.rhs}
		require.Equal(t, lhs.ConverseImplication(rhs), test.result)
	}
}

func TestNetFlagsMirror(t *testing.T) {
	flags := &NetFlags{uint64(6)}
	flagsSlice := flags.ToBoolSlice()
	require.Equal(t, NetFlagsFromBoolSlice(flagsSlice), flags)
}
