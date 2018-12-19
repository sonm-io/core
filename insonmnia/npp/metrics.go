package npp

import (
	"net"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	prometheusIO "github.com/prometheus/client_model/go"
	"github.com/rcrowley/go-metrics"
	"go.uber.org/atomic"
)

type ListenerMetrics struct {
	RendezvousAddr       net.Addr
	NumConnectionsDirect uint64
	NumConnectionsNAT    uint64
	NumConnectionsRelay  uint64
}

type listeneMetrics struct {
	NumConnectionsDirect *atomic.Uint64
	NumConnectionsNAT    *atomic.Uint64
	NumConnectionsRelay  *atomic.Uint64
}

func newListenerMetrics() *listeneMetrics {
	return &listeneMetrics{
		NumConnectionsDirect: atomic.NewUint64(0),
		NumConnectionsNAT:    atomic.NewUint64(0),
		NumConnectionsRelay:  atomic.NewUint64(0),
	}
}

type dialMetrics struct {
	// NumAttempts describes the total number of attempts to connect
	// to a remote address using NPP dialer.
	NumAttempts *meterWrapper
	// NumSuccess describes the total number of successful attempts to connect
	// to a remote address using NPP dialer no matter which method was
	// successful.
	NumSuccess *meterWrapper
	// NumFailed describes the total number of failed attempts to connect to
	// a remote address.
	NumFailed *meterWrapper
	// UsingTCPDirectHistogram describes the distribution of connect times for
	// successful connection attempts using direct TCP connection.
	UsingTCPDirectHistogram *histogramWrapper
	// UsingNATHistogram describes the distribution of resolve and connect
	// times for successful connection attempts using NPP NAT traversal.
	UsingNATHistogram *histogramWrapper
	// UsingQNATHistogram describes the distribution of resolve and connect
	// times for successful connection attempts using NPP NAT traversal over
	// UDP for QUIC.
	UsingQNATHistogram *histogramWrapper
	// UsingRelayHistogram describes the distribution of resolve and connect
	// times for successful connection attempts using Relay server.
	UsingRelayHistogram *histogramWrapper
	// SummaryHistogram describes the distribution of connect times for overall
	// dialing.
	SummaryHistogram *histogramWrapper
	// LastTimeActive shows the time when the last connection attempt was made.
	LastTimeActive *gaugeWrapper
	// LastTimeSuccess shows the time when the last successful connection
	// attempt was made.
	LastTimeSuccess *gaugeWrapper
}

func newDialMetrics() *dialMetrics {
	return &dialMetrics{
		NumAttempts:             newMeterWrapper(),
		NumSuccess:              newMeterWrapper(),
		NumFailed:               newMeterWrapper(),
		UsingTCPDirectHistogram: newHistogramWrapper(),
		UsingNATHistogram:       newHistogramWrapper(),
		UsingQNATHistogram:      newHistogramWrapper(),
		UsingRelayHistogram:     newHistogramWrapper(),
		SummaryHistogram:        newHistogramWrapper(),
		LastTimeActive:          newGaugeWrapper(),
		LastTimeSuccess:         newGaugeWrapper(),
	}
}

func (m *dialMetrics) MetricNames() []string {
	v := reflect.ValueOf(m).Elem()
	ty := v.Type()

	names := make([]string, v.NumField())
	for id := 0; id < v.NumField(); id++ {
		names[id] = ty.Field(id).Name
	}

	return names
}

type NamedMetric struct {
	Name   string
	Metric *prometheusIO.Metric
}

type gaugeWrapper struct {
	prometheus.Gauge
}

func newGaugeWrapper() *gaugeWrapper {
	return &gaugeWrapper{
		Gauge: prometheus.NewGauge(prometheus.GaugeOpts{}),
	}
}

func (m *gaugeWrapper) ToNamedMetrics(prefix string) []*NamedMetric {
	return []*NamedMetric{{Name: prefix + "", Metric: newPrometheusMetric(m)}}
}

type meterWrapper struct {
	metrics.Meter
}

func newMeterWrapper() *meterWrapper {
	return &meterWrapper{
		Meter: metrics.NewMeter(),
	}
}

func (m *meterWrapper) ToNamedMetrics(prefix string) []*NamedMetric {
	return []*NamedMetric{
		{Name: prefix + "", Metric: newPrometheusGaugeMetric(func() float64 { return float64(m.Count()) })},
		{Name: prefix + "Rate01", Metric: newPrometheusGaugeMetric(m.Rate1)},
		{Name: prefix + "Rate05", Metric: newPrometheusGaugeMetric(m.Rate5)},
		{Name: prefix + "Rate15", Metric: newPrometheusGaugeMetric(m.Rate15)},
		{Name: prefix + "RateMean", Metric: newPrometheusGaugeMetric(m.RateMean)},
	}
}

type histogramWrapper struct {
	prometheus.Histogram
}

func newHistogramWrapper() *histogramWrapper {
	return &histogramWrapper{
		Histogram: prometheus.NewHistogram(prometheus.HistogramOpts{}),
	}
}

func (m *histogramWrapper) ToNamedMetrics(prefix string) []*NamedMetric {
	return []*NamedMetric{{Name: prefix + "", Metric: newPrometheusMetric(m)}}
}

func newPrometheusMetric(metric prometheus.Metric) *prometheusIO.Metric {
	value := &prometheusIO.Metric{}

	if err := metric.Write(value); err != nil {
		// Unreachable actually.
		return nil
	}

	return value
}

func newPrometheusGaugeMetric(fn func() float64) *prometheusIO.Metric {
	metric := prometheus.NewGauge(prometheus.GaugeOpts{})
	metric.Set(fn())
	return newPrometheusMetric(metric)
}
