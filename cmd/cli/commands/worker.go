package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	workerMgmtCmd.AddCommand(
		workerStatusCmd,
		masterRootCmd,
		askPlansRootCmd,
		workerTasksCmd,
		workerDevicesCmd,
	)
}

var workerMgmtCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker management",
}

var workerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show worker status",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := context.Background()
		hub, err := newWorkerManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		status, err := hub.Status(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get worker status", err)
			os.Exit(1)
		}

		printHubStatus(cmd, status)
	},
}
