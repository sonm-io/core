package optimus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	values := []float64{
		0.0,
		1.0,
		5.0,
		10.0,
	}
	normalizer, err := newNormalizer(values...)

	require.NotNil(t, normalizer)
	require.NoError(t, err)

	normalizer.NormalizeBatch(values)

	assert.Equal(t, 0.0, values[0])
	assert.Equal(t, 0.1, values[1])
	assert.Equal(t, 0.5, values[2])
	assert.Equal(t, 1.0, values[3])
}

func TestMeanNormalize(t *testing.T) {
	values := []float64{
		0.0,
		1.0,
		5.0,
		10.0,
	}
	normalizer, err := newMeanNormalizer(values)

	require.NotNil(t, normalizer)
	require.NoError(t, err)

	normalizer.NormalizeBatch(values)

	assert.Equal(t, -0.4, values[0])
	assert.Equal(t, -0.3, values[1])
	assert.Equal(t, +0.1, values[2])
	assert.Equal(t, +0.6, values[3])
}

func TestMeanNormalizeDegeneratedVector(t *testing.T) {
	values := []float64{
		1.0,
		1.0,
		1.0,
		1.0,
	}
	normalizer, err := newMeanNormalizer(values)

	require.Nil(t, normalizer)
	require.EqualError(t, err, ErrDegenerateVector.Error())
}
