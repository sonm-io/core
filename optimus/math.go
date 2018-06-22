package optimus

import (
	"math"
)

type sigmoid func(x float64) float64

type sigmoidConfig struct {
	Alpha float64 `yaml:"α" default:"10.0"`
	Delta float64 `yaml:"∂" default:"43200.0"`
}

func newSigmoid(cfg sigmoidConfig) sigmoid {
	return func(x float64) float64 {
		return 1 - (1 / (1 + math.Exp(-cfg.Alpha*(x-cfg.Delta)/cfg.Delta)))
	}
}
