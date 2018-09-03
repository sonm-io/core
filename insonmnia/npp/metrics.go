package npp

import (
	"net"
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	prometheusIO "github.com/prometheus/client_model/go"
	"go.uber.org/atomic"
)

type ListenerMetrics struct {
	RendezvousAddr       net.Addr
	NumConnectionsDirect uint64
	NumConnectionsNAT    uint64
	NumConnectionsRelay  uint64
}

type metrics struct {
	NumConnectionsDirect *atomic.Uint64
	NumConnectionsNAT    *atomic.Uint64
	NumConnectionsRelay  *atomic.Uint64
}

func newMetrics() *metrics {
	return &metrics{
		NumConnectionsDirect: atomic.NewUint64(0),
		NumConnectionsNAT:    atomic.NewUint64(0),
		NumConnectionsRelay:  atomic.NewUint64(0),
	}
}

type dialMetrics struct {
	// NumAttempts describes the total number of attempts to connect
	// to a remote address using NPP dialer.
	NumAttempts prometheus.Counter
	// NumSuccess describes the total number of successful attempts to connect
	// to a remote address using NPP dialer no matter which method was
	// successful.
	NumSuccess prometheus.Counter
	// NumFailed describes the total number of failed attempts to connect to
	// a remove address.
	NumFailed prometheus.Counter
	// UsingTCPDirectHistogram describes the distribution of connect times for
	// successful connection attempts using direct TCP connection.
	UsingTCPDirectHistogram prometheus.Histogram
	// UsingNATHistogram describes the distribution of resolve and connect
	// times for successful connection attempts using NPP NAT traversal.
	UsingNATHistogram prometheus.Histogram
	// UsingRelayHistogram describes the distribution of resolve and connect
	// times for successful connection attempts using Relay server.
	UsingRelayHistogram prometheus.Histogram
	// SummaryHistogram describes the distribution of connect times for overall
	// dialing.
	SummaryHistogram prometheus.Histogram
	// LastTimeActive shows the time when the last connection attempt was made.
	LastTimeActive prometheus.Gauge
	// LastTimeSuccess shows the time when the last successful connection
	// attempt was made.
	LastTimeSuccess prometheus.Gauge
}

func newDialMetrics() *dialMetrics {
	return &dialMetrics{
		NumAttempts:             prometheus.NewCounter(prometheus.CounterOpts{}),
		NumSuccess:              prometheus.NewCounter(prometheus.CounterOpts{}),
		NumFailed:               prometheus.NewCounter(prometheus.CounterOpts{}),
		UsingTCPDirectHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{}),
		UsingNATHistogram:       prometheus.NewHistogram(prometheus.HistogramOpts{}),
		UsingRelayHistogram:     prometheus.NewHistogram(prometheus.HistogramOpts{}),
		SummaryHistogram:        prometheus.NewHistogram(prometheus.HistogramOpts{}),
		LastTimeActive:          prometheus.NewGauge(prometheus.GaugeOpts{}),
		LastTimeSuccess:         prometheus.NewGauge(prometheus.GaugeOpts{}),
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
