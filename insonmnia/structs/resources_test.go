package structs

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskResourcesRestrictCPUs(t *testing.T) {
	resources, err := NewTaskResources(&sonm.TaskResourceRequirements{
		CPUCores: 2,
	})

	require.NotNil(t, resources)
	require.NoError(t, err)

	containerResources := resources.ToContainerResources("")

	assert.Equal(t, int64(100000), containerResources.CPUPeriod)
	assert.Equal(t, int64(200000), containerResources.CPUQuota)
}

func TestValidateSlot_SingleGPU(t *testing.T) {
	s := &sonm.Resources{
		GpuCount: sonm.GPUCount_SINGLE_GPU,
	}

	err := ValidateResources(s)
	assert.EqualError(t, err, ErrUnsupportedSingleGPU.Error())
}
