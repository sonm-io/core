package optimus

import (
	"fmt"

	"go.uber.org/zap"
)

type RegressionModelFactory interface {
	Config() interface{}
	Create(log *zap.SugaredLogger) Model
}

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

type regressionModelFactory struct {
	RegressionModelFactory
}

func (m *regressionModelFactory) MarshalYAML() (interface{}, error) {
	return m.Config(), nil
}

func (m *regressionModelFactory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ty, err := typeofInterface(unmarshal)
	if err != nil {
		return err
	}

	factory := regressionFactory(ty)
	if factory == nil {
		return fmt.Errorf("unknown regression model: %s", ty)
	}

	if err := unmarshal(factory.Config()); err != nil {
		return err
	}

	m.RegressionModelFactory = factory

	return nil
}

func regressionFactory(ty string) RegressionModelFactory {
	switch ty {
	case "lls":
		return &llsModelConfig{}
	case "nnls":
		return &SCAKKTModel{}
	default:
		return nil
	}
}
