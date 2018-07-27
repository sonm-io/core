package optimus

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/linear"
	"go.uber.org/zap"
)

type llsModelConfig struct {
	Alpha          float64 `yaml:"alpha" default:"1e-3"`
	Regularization float64 `yaml:"regularization" default:"6.0"`
	MaxIterations  int     `yaml:"max_iterations" default:"1000"`
}

func (m *llsModelConfig) Config() interface{} {
	return m
}

func (m *llsModelConfig) Create(log *zap.SugaredLogger) Model {
	return &llsModel{
		cfg:    *m,
		output: ioutil.Discard,
	}
}

type llsModel struct {
	cfg    llsModelConfig
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
