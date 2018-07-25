// This package implements of NNLS model, which solves the non-negative least
// squares problem.
//
// See http://cmp.felk.cvut.cz/ftp/articles/franc/Franc-Hlavac-Navara-CAIP05.pdf
// for more details.

package optimus

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
)

type SCAKKTModel struct {
	MaxIterations int
	Log           *zap.SugaredLogger
}

func newSCAKKTModel() modelFactory {
	return func(log *zap.Logger) Model {
		return &SCAKKTModel{
			Log:           log.Sugar(),
			MaxIterations: 1e7,
		}
	}
}

func (m *SCAKKTModel) Train(trainingSet [][]float64, expectation []float64) (TrainedModel, error) {
	if len(trainingSet) == 0 {
		return nil, fmt.Errorf("training set is empty")
	}
	if len(trainingSet) != len(expectation) {
		return nil, fmt.Errorf("number of training set rows must be the same as the size of expectation vector")
	}

	// Addition of ⟨e⟩ vector to the training set is not required, because we
	// don't want to have a "default price".
	m.Log.Debugf("training M[%dx%d] -> E[%dx1] ...", len(trainingSet), len(trainingSet[0]), len(expectation))

	now := time.Now()
	a, numIterations, err := SCAKKT(trainingSet, expectation, 1e-9, m.MaxIterations)
	if err != nil {
		return nil, err
	}

	if numIterations == m.MaxIterations {
		return nil, fmt.Errorf("reached max number of iterations: %d", numIterations)
	}
	for id := range a {
		if math.IsNaN(a[id]) {
			return nil, fmt.Errorf("some of output parameters are NaN")
		}
	}

	trainedModel := &TrainedSCAKKTModel{
		a:             a,
		numIterations: numIterations,
	}

	m.Log.Debugf("training complete in %s and %d iterations: %s", time.Since(now), numIterations, trainedModel.String())

	return trainedModel, nil
}

type TrainedSCAKKTModel struct {
	a             []float64
	numIterations int
}

func (m *TrainedSCAKKTModel) Predict(x []float64) (float64, error) {
	if len(x) != len(m.a) { // TODO: Why not ">" ?
		return math.NaN(), errors.New("size of input vector must be the same as the size of training vector")
	}

	f := 0.0
	for id, xi := range x {
		f += xi * m.a[id]
	}

	return f, nil
}

func (m *TrainedSCAKKTModel) String() string {
	var parts []string

	for id := range m.a {
		if m.a[id] < 1e-6 {
			continue
		}

		parts = append(parts, fmt.Sprintf("%.6fx[%d]", m.a[id], id))
	}

	return "f(x) = aᵀx = " + strings.Join(parts, " + ")
}

// SCAKKT solves the non-negative least squares problem.
//
// Argument "A" represents input matrix. Each element of "A" must have the same
// length. Vector "b" represents a measurement or output vector.
// "Eps" is a tolerance for stopping iteration.
//
// The result tuple consists of three parameters. The first one returns
// coefficients of the fitted linear function. It will have the same length
// as elements of "A".
// The second value is the number of iterations performed.
// An error is returned if "A" and "b" are not the same length.
func SCAKKT(A [][]float64, b []float64, eps float64, maxIterations int) ([]float64, int, error) {
	m := len(A)
	n := len(A[0])
	if len(b) != m {
		return nil, 0, errors.New("both A, b must have the same length")
	}

	H, Hd := hessian(A, b)

	// Output vector.
	x := make([]float64, n)
	mu := make([]float64, n)
	for j := range mu {
		e := 0.
		for i, bi := range b {
			e -= bi * A[i][j]
		}
		mu[j] = e
	}

	nEps := -eps
	i := 1

	for ; i < maxIterations; i++ {
		ch := false
		for k, xk := range x {
			b := xk - mu[k]/Hd[k]
			if b < 0 {
				b = 0
			}

			if b == xk {
				continue
			}

			x[k] = b
			ch = true
			b -= xk

			for j, h := range H[k] {
				mu[j] += b * h
			}
		}
		if !ch {
			break
		}
		// εKKT criteria.
		for k, m := range mu {
			xk := x[k]
			if xk < 0 {
				break
			}
			if m < nEps {
				break
			}
			if xk > 0 && m > eps {
				break
			}

			return x, i, nil
		}
	}

	return x, i, nil
}

func hessian(A [][]float64, b []float64) ([][]float64, []float64) {
	n := len(A[0])

	// Hessian matrix. Defined as H = AᵀA.
	H := make([][]float64, n)
	// Copy of H diagonal.
	Hd := make([]float64, n)
	for i := range H {
		Hi := make([]float64, n)
		for j := 0; j < i; j++ {
			Hi[j] = H[j][i]
		}
		s := 0.0
		for k := range b {
			e := A[k][i]
			s += e * e
		}
		Hi[i] = s
		Hd[i] = s
		for j := i + 1; j < n; j++ {
			s := 0.0
			for k := range b {
				s += A[k][i] * A[k][j]
			}
			Hi[j] = s
		}
		H[i] = Hi
	}

	return H, Hd
}
