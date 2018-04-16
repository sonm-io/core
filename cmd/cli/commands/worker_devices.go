package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Show Worker's hardware",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newWorkerManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		devices, err := hub.Devices(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}
