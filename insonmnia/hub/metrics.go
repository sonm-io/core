package hub

import "github.com/prometheus/client_golang/prometheus"

var (
	tasksGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sonm_tasks_current",
		Help: "Number of currently running tasks",
	})
	dealsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sonm_deals_current",
		Help: "Number of currently running tasks",
	})
)

func init() {
	prometheus.MustRegister(tasksGauge)
	prometheus.MustRegister(dealsGauge)
}
