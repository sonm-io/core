package optimus

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/linear"
)

type ModelConfig struct {
	Alpha          float64 `yaml:"alpha" default:"1e-3"`
	Regularization float64 `yaml:"regularization" default:"6.0"`
	MaxIterations  int     `yaml:"max_iterations" default:"1000"`
}

func (m *newModel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ty, err := typeofInterface(unmarshal)
	if err != nil {
		return err
	}

	switch ty {
	case "lls":
		cfg := ModelConfig{}
		if err := unmarshal(&cfg); err != nil {
			return err
		}

		*m = newLLSModel(cfg)
	default:
		return fmt.Errorf("unknown model: %s", ty)
	}

	return nil
}

type Model interface {
	Train(trainingSet [][]float64, expectation []float64) (TrainedModel, error)
}

type TrainedModel interface {
	Predict(vec []float64) (float64, error)
}

type newModel func() Model

func newLLSModel(cfg ModelConfig) newModel {
	return func() Model {
		return &llsModel{
			cfg:    cfg,
			output: ioutil.Discard,
		}
	}
}

type llsModel struct {
	cfg    ModelConfig
	output io.Writer
}

func (m *llsModel) Train(trainingSet [][]float64, expectation []float64) (TrainedModel, error) {
	model := linear.NewLeastSquares(base.BatchGA, m.cfg.Alpha, m.cfg.Regularization, m.cfg.MaxIterations, trainingSet, expectation)
	//model.Output = m.output
	if err := model.Learn(); err != nil {
		return nil, err
	}

	return &trainedLLSModel{
		model: model,
	}, nil
}

type trainedLLSModel struct {
	model *linear.LeastSquares
}

func (m *trainedLLSModel) Predict(vec []float64) (float64, error) {
	prediction, err := m.model.Predict(vec)
	if err != nil {
		return 0.0, err
	}

	if len(prediction) == 0 {
		return 0.0, fmt.Errorf("no prediction made")
	}

	return prediction[0], nil
}
