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
