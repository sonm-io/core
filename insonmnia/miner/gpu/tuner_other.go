// +build !cl

package gpu

import (
	"context"
)

// override all Tuner implementation with NilTuner if we building without GPU support

func newRadeonTuner(_ context.Context, _ *tunerOptions) (Tuner, error) {
	return NilTuner{}, nil
}

func newNvidiaTuner(_ context.Context, _ *tunerOptions) (Tuner, error) {
	return NilTuner{}, nil
}

func newNvidiaDockerTuner(_ context.Context, _ *tunerOptions) (Tuner, error) {
	return NilTuner{}, nil
}
