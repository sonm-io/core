package connor

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestApplyEnvTemplate(t *testing.T) {
	m1 := map[string]string{
		"a":     "b",
		"c":     "123",
		"param": "command",
	}

	dealID := sonm.NewBigIntFromInt(123)

	// no templates in map, should not be changed
	m1 = applyEnvTemplate(m1, dealID)
	assert.Len(t, m1, 3)
	assert.Equal(t, "b", m1["a"])
	assert.Equal(t, "123", m1["c"])
	assert.Equal(t, "command", m1["param"])

	m2 := map[string]string{
		"deal":   "{DEAL_ID}",
		"worker": "sonm_{DEAL_ID}",
		"foo":    "{DEAL_ID}_{DEAL_ID}_{DEAL_ID}{DEAL_ID}_c{DEAL_ID}",
	}

	// map's values have templates, should be changed
	m2 = applyEnvTemplate(m2, dealID)
	assert.Len(t, m2, 3)

	assert.Equal(t, "123", m2["deal"])
	assert.Equal(t, "sonm_123", m2["worker"])
	assert.Equal(t, "123_123_123123_c123", m2["foo"])

}
