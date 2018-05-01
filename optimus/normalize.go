package optimus

import (
	"github.com/montanaflynn/stats"
	"github.com/pkg/errors"
)

var (
	ErrDegenerateVector = errors.New("all elements in the vector the same")
)

type Normalizer interface {
	Normalize(x float64) float64
	NormalizeBatch(x []float64)
	Denormalize(x float64) float64
}

type meanNormalizer struct {
	mean  float64
	scale float64
}

func newMeanNormalizer(values []float64) (Normalizer, error) {
	min, err := stats.Min(values)
	if err != nil {
		return nil, err
	}

	max, err := stats.Max(values)
	if err != nil {
		return nil, err
	}

	scale := max - min
	if scale == 0.0 {
		return nil, ErrDegenerateVector
	}

	mean, err := stats.Mean(values)
	if err != nil {
		return nil, err
	}

	m := &meanNormalizer{
		mean:  mean,
		scale: scale,
	}

	return m, nil
}

func (m *meanNormalizer) Normalize(x float64) float64 {
	return (x - m.mean) / m.scale
}

func (m *meanNormalizer) NormalizeBatch(x []float64) {
	for i, v := range x {
		x[i] = m.Normalize(v)
	}
}

func (m *meanNormalizer) Denormalize(x float64) float64 {
	return x*m.scale + m.mean
}

// Normalizer rescales the range of features to scale the range in [0, 1].
type normalizer struct {
	min   float64
	max   float64
	scale float64
}

func newNormalizer(vec ...float64) (*normalizer, error) {
	m := &normalizer{}

	if len(vec) > 0 {
		min, err := stats.Min(vec)
		if err != nil {
			return nil, err
		}

		max, err := stats.Max(vec)
		if err != nil {
			return nil, err
		}

		m.min = min
		m.max = max
		m.scale = max - min
	}

	return m, nil
}

func (m *normalizer) Add(x float64) {
	if x < m.min {
		m.min = x
	}

	if x > m.max {
		m.max = x
	}

	m.scale = m.max - m.min
}

func (m *normalizer) IsDegenerated() bool {
	return m.scale == 0.0
}

func (m *normalizer) Normalize(x float64) float64 {
	return (x - m.min) / m.scale
}

func (m *normalizer) NormalizeBatch(x []float64) {
	for i, v := range x {
		x[i] = m.Normalize(v)
	}
}

func (m *normalizer) Denormalize(x float64) float64 {
	return x*m.scale + m.min
}
