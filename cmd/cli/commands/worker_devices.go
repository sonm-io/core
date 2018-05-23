package commands

import (
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerDevicesCmd = &cobra.Command{
	Use:    "devices",
	Short:  "Show Worker's hardware",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := worker.Devices(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}
