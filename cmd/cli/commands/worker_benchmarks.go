package commands

import (
	"fmt"
	"strconv"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	benchmarkRootCmd.AddCommand(
		workerRemoveBenchmarksCmd,
		workerPurgeBenchmarksCmd,
	)
}

var benchmarkRootCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Operations with worker benchmarks",
}

var workerPurgeBenchmarksCmd = &cobra.Command{
	Use:   "purge",
	Short: "Remove all benchmarks from cache to rebenchmark on next worker restart",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := worker.PurgeBenchmarks(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to purge benchmark cache: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var workerRemoveBenchmarksCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove specified benchmark from cache to rebenchmark on next worker restart",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseUint(args[0], 0, 64)
		if err != nil {
			return fmt.Errorf("failed to parse benchmark id: %v", err)
		}

		if _, err := worker.RemoveBenchmark(workerCtx, &sonm.NumericID{Id: id}); err != nil {
			return fmt.Errorf("failed to get debug state: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
