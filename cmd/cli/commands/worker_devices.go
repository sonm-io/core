package commands

import (
	"fmt"

	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Show Worker's hardware",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := worker.Devices(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get devices list: %v", err)
		}

		printDeviceList(cmd, devices)
		return nil
	},
}

var workerFreeDevicesCmd = &cobra.Command{
	Use:   "free_devices",
	Short: "Show Worker's hardware with remaining resources available for scheduling",
	RunE: func(cmd *cobra.Command, args []string) error {
		devices, err := worker.FreeDevices(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get devices list: %v", err)
		}

		printDeviceList(cmd, devices)
		return nil
	},
}
