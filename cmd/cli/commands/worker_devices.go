package commands

import (
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Show Worker's hardware",
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := worker.Devices(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}

var workerFreeDevicesCmd = &cobra.Command{
	Use:   "free_devices",
	Short: "Show Worker's hardware with remaining resources available for scheduling",
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := worker.FreeDevices(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}
