package optimus

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/linear"
	"go.uber.org/zap"
)

// Model represents a ML model that can be trained using provided training set
// with some expectation.
type Model interface {
	// Train runs the training process, returning the trained model on success.
	// The "trainingSet" argument is a MxN matrix, while "expectation" must be
	// a M-length vector.
	Train(trainingSet [][]float64, expectation []float64) (TrainedModel, error)
}

type TrainedModel interface {
	Predict(vec []float64) (float64, error)
}

type llsModelConfig struct {
	Alpha          float64 `yaml:"alpha" default:"1e-3"`
	Regularization float64 `yaml:"regularization" default:"6.0"`
	MaxIterations  int     `yaml:"max_iterations" default:"1000"`
}

type modelFactory func(log *zap.Logger) Model

// NewModelFactory constructs a new model factory using provided unmarshaller.
func newModelFactory(cfgUnmarshal func(interface{}) error) (modelFactory, error) {
	ty, err := typeofInterface(cfgUnmarshal)
	if err != nil {
		return nil, err
	}

	switch ty {
	case "lls":
		cfg := llsModelConfig{}
		if err := cfgUnmarshal(&cfg); err != nil {
			return nil, err
		}

		return newLLSModelFactory(cfg), nil
	case "nnls":
		return newSCAKKTModel(), nil
	default:
		return nil, fmt.Errorf("unknown model: %s", ty)
	}
}

func (m *modelFactory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	factory, err := newModelFactory(unmarshal)
	if err != nil {
		return err
	}

	*m = factory

	return nil
}

func newLLSModelFactory(cfg llsModelConfig) modelFactory {
	return func(log *zap.Logger) Model {
		return &llsModel{
			cfg:    cfg,
			output: ioutil.Discard,
		}
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
