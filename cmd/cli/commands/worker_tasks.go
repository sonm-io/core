package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerTasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Show tasks running on Worker",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		list, err := hub.Tasks(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get task list", err)
			os.Exit(1)
		}

		printNodeTaskStatus(cmd, list.GetInfo())
	},
}
