package commands

import (
	"fmt"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var workerMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show Worker's hardware monitoring",
	RunE: func(cmd *cobra.Command, args []string) error {
		metrics, err := worker.Metrics(workerCtx, &sonm.WorkerMetricsRequest{})
		if err != nil {
			return fmt.Errorf("cannot get metrics: %v", err)
		}

		data, err := yaml.Marshal(metrics)
		if err != nil {
			return fmt.Errorf("failed to marshal metrics into YAML: %v", err)
		}

		cmd.Println(string(data))
		return nil
	},
}
