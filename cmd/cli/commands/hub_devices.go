package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	hubDeviceRootCmd.AddCommand(
		deviceListCmd,
		deviceGetPropsCmd,
		deviceSetDevPropsCmd,
	)
}

var hubDeviceRootCmd = &cobra.Command{
	Use:   "dev",
	Short: "Device properties",
}

var deviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show Hub's aggregated hardware",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		devices, err := hub.DeviceList(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}

var deviceGetPropsCmd = &cobra.Command{
	Use:   "get <dev_id>",
	Short: "Get Device properties",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		devID := args[0]
		reply, err := hub.GetDeviceProperties(ctx, &pb.ID{Id: devID})
		if err != nil {
			showError(cmd, "Cannot get device properties", err)
			os.Exit(1)
		}

		printDevicesProps(cmd, reply.GetProperties())
	},
}

var deviceSetDevPropsCmd = &cobra.Command{
	Use:   "set <dev_id> <props.yaml>",
	Short: "Set Device properties",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		workerID := args[0]
		propsFile := args[1]

		props, err := loadPropsFile(propsFile)
		if err != nil {
			showError(cmd, errCannotParsePropsFile.Error(), nil)
			os.Exit(1)
		}

		req := &pb.SetDevicePropertiesRequest{
			ID:         workerID,
			Properties: props,
		}

		_, err = hub.SetDeviceProperties(ctx, req)
		if err != nil {
			showError(cmd, "Cannot set device properties", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}
