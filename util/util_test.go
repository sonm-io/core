package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		mustErr  bool
	}{
		{
			input:    "192.168.0.1:10001",
			expected: "10001",
			mustErr:  false,
		},
		{
			input:    ":10002",
			expected: "10002",
			mustErr:  false,
		},
		{
			input:    "192.168.0.1",
			expected: "",
			mustErr:  true,
		},
		{
			input:    "192.168.0.1:qwer",
			expected: "",
			mustErr:  true,
		},
		{
			input:    "192.168.0.1:99999",
			expected: "",
			mustErr:  true,
		},
	}

	for _, tt := range tests {
		port, err := ParseEndpointPort(tt.input)
		assert.Equal(t, tt.expected, port)
		if tt.mustErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestParseTaskID(t *testing.T) {
	tests := []struct {
		in       string
		task     string
		hub      string
		mustFail bool
	}{
		{
			in:       "aaa@bbb",
			task:     "aaa",
			hub:      "bbb",
			mustFail: false,
		},
		{
			in:       "aaa@",
			mustFail: true,
		},
		{
			in:       "@bbb",
			mustFail: true,
		},
		{
			in:       "@",
			mustFail: true,
		},
		{
			in:       "",
			mustFail: true,
		},
	}

	for _, tt := range tests {
		task, hub, err := ParseTaskID(tt.in)
		assert.True(t, (err != nil) == tt.mustFail)
		assert.Equal(t, tt.task, task)
		assert.Equal(t, tt.hub, hub)
	}
}
