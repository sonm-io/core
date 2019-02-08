package commands

import (
	"fmt"
	"sort"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show Worker's hardware monitoring",
	RunE: func(cmd *cobra.Command, args []string) error {
		metrics, err := worker.Metrics(workerCtx, &sonm.WorkerMetricsRequest{})
		if err != nil {
			return fmt.Errorf("cannot get metrics: %v", err)
		}

		if isSimpleFormat() {
			var fields []string
			for name, value := range metrics.GetMetrics() {
				fields = append(fields, fmt.Sprintf("%s: %.2f", name, value))
			}

			sort.Strings(fields)
			for _, f := range fields {
				cmd.Println(f)
			}
		} else {
			showJSON(cmd, metrics)
		}

		return nil
	},
}
