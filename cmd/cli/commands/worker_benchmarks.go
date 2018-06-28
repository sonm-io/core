package commands

import (
	"os"
	"strconv"

	pb "github.com/sonm-io/core/proto"
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
	Run: func(cmd *cobra.Command, args []string) {
		_, err := worker.PurgeBenchmarks(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "failed to purge benchmark cache", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var workerRemoveBenchmarksCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove specified benchmark from cache to rebenchmark on next worker restart",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.ParseUint(args[0], 0, 64)
		if err != nil {
			showError(cmd, "failed to parse benchmark id", err)
			os.Exit(1)
		}
		_, err = worker.RemoveBenchmark(workerCtx, &pb.NumericID{Id: id})
		if err != nil {
			showError(cmd, "failed to get debug state", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}
