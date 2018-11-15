// +build !cl darwin

package gpu

import "context"

func newNvidiaMetricsHandler(ctx context.Context) (MetricsHandler, error) {
	return nilMetricsHandler{}, nil
}

func newRadeonMetricsHandler(ctx context.Context) (MetricsHandler, error) {
	return nilMetricsHandler{}, nil
}
