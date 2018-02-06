package gpu

import (
	"fmt"
	"testing"

	"strconv"

	"github.com/stretchr/testify/assert"
)

func TestMatchDeviceName(t *testing.T) {
	tests := []struct {
		name  string
		num   int64
		match bool
	}{
		{name: "card0", num: 0, match: true},
		{name: "card2", num: 2, match: true},
		{name: "card10", num: 10, match: true},
		{name: "card", match: false},
		{name: "ccard0", match: false},
		{name: "cardX", match: false},
		{name: "controlD64", match: false},
		{name: "renderD129", match: false},
	}

	for _, tt := range tests {
		m := devDriCardNameRe.FindStringSubmatch(tt.name)
		match := m != nil && len(m) == 3

		if tt.match {
			num, _ := strconv.ParseInt(m[2], 10, 64)
			assert.True(t, match)
			assert.Equal(t, tt.num, num, tt.name)
		} else {
			assert.False(t, match, fmt.Sprintf("value %v must not be parsed", tt.name))
		}
	}
}

func TestParseSysClassValue(t *testing.T) {
	tests := []struct {
		in       string
		out      uint64
		mustFail bool
	}{
		{in: "0x0", out: 0},
		{in: "0x1", out: 1},
		{in: "0xff", out: 255},
		{in: "fff", out: 4095},
		{in: "fff\r", out: 4095},
		{in: "fff\r\n", out: 4095},
		{in: "", mustFail: true},
		{in: "\r", mustFail: true},
		{in: "\t", mustFail: true},
		{in: "0xp1d0r", mustFail: true},
	}

	for _, tt := range tests {
		out, err := parseSysClassValue([]byte(tt.in))
		if tt.mustFail {
			assert.Error(t, err)
		} else {
			assert.Equal(t, tt.out, out, fmt.Sprintf("expect %s == %d", tt.in, out))
		}
	}
}

func TestParsePCISlotName(t *testing.T) {
	tests := []struct {
		in   string
		out  string
		isOK bool
	}{
		{in: "", out: "", isOK: false},
		{in: "aaabbb", out: "", isOK: false},
		{in: "aaa=bbb", out: "", isOK: false},
		{in: "PCI_SLOT_NAME", out: "", isOK: false},
		{in: "PCI_SLOT_NAME=", out: "", isOK: true},
		{in: "PCI_SLOT_NAME=0000:01:00.0", out: "0000:01:00.0", isOK: true},
	}

	for _, tt := range tests {
		out, ok := parsePCISlotName(tt.in)
		assert.Equal(t, ok, tt.isOK)
		assert.Equal(t, out, tt.out)
	}
}
