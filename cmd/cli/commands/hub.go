package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	hubRootCmd.AddCommand(
		hubStatusCmd,
		hubWorkerRootCmd,
		hubACLRootCmd,
		hubOrderRootCmd,
		hubTasksRootCmd,
		hubDeviceRootCmd,
	)
}

var hubRootCmd = &cobra.Command{
	Use:   "hub",
	Short: "Hub management",
}

var hubStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show hub status",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		status, err := hub.Status(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get hub status", err)
			os.Exit(1)
		}

		printHubStatus(cmd, status)
	},
}
