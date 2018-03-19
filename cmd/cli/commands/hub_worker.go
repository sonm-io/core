package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	hubWorkerRootCmd.AddCommand(
		hubWorkerListCmd,
		hubWorkerStatusCmd,
	)
}

var hubWorkerRootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Operations with connected Workers",
}

var hubWorkerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show connected workers list",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		list, err := hub.WorkersList(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get workers list", err)
			os.Exit(1)
		}

		printWorkerList(cmd, list)
	},
}

var hubWorkerStatusCmd = &cobra.Command{
	Use:   "status <worker_id>",
	Short: "Show worker status",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		workerID := args[0]
		status, err := hub.WorkerStatus(ctx, &pb.ID{Id: workerID})
		if err != nil {
			showError(cmd, "Cannot get workers status", err)
			os.Exit(1)
		}

		printWorkerStatus(cmd, workerID, status)
	},
}
