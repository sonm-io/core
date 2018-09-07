package node

import (
	"context"

	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/proto"
)

type monitoringService struct {
	nppDialer *npp.Dialer
}

func newMonitoringAPI(opts *remoteOptions) *monitoringService {
	return &monitoringService{
		nppDialer: opts.nppDialer,
	}
}

func (m *monitoringService) MetricsNPP(context.Context, *sonm.Empty) (*sonm.NPPMetricsReply, error) {
	metrics, err := m.nppDialer.Metrics()
	if err != nil {
		return nil, err
	}

	metricsResponse := map[string]*sonm.NamedMetrics{}

	for addr, metric := range metrics {
		addrMetrics := make([]*sonm.NamedMetric, 0)
		for _, addrMetric := range metric {
			addrMetrics = append(addrMetrics, &sonm.NamedMetric{
				Name:   addrMetric.Name,
				Metric: addrMetric.Metric,
			})
		}

		metricsResponse[addr] = &sonm.NamedMetrics{Metrics: addrMetrics}
	}

	return &sonm.NPPMetricsReply{Metrics: metricsResponse}, nil
}
