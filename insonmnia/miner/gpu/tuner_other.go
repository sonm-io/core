// +build !cl darwin

package gpu

import (
	"context"
)

// override all Tuner implementation with NilTuner if we building without GPU support

func newRadeonTuner(_ context.Context, opts ...Option) (Tuner, error) {
	return NilTuner{}, nil
}

func newNvidiaTuner(_ context.Context, opts ...Option) (Tuner, error) {
	return NilTuner{}, nil
}
