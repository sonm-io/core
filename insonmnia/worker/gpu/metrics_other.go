// +build !cl darwin

package gpu

func newNvidiaMetricsHandler() (MetricsHandler, error) {
	return nilMetricsHandler{}, nil
}

func newRadeonMetricsHandler() (MetricsHandler, error) {
	return nilMetricsHandler{}, nil
}
