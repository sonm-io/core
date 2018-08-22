package antifraud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	var f flags = SkipBlacklisting
	assert.True(t, f.SkipBlacklist())

	f = AllChecks
	assert.False(t, f.SkipBlacklist())
}
