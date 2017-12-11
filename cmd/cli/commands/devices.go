package commands

import (
	"os"

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
	Use:    "list",
	Short:  "Show Hub's aggregated hardware",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		devices, err := hub.DevicesList()
		if err != nil {
			showError(cmd, "Cannot get devices list", err)
			os.Exit(1)
		}

		printDeviceList(cmd, devices)
	},
}

var deviceGetPropsCmd = &cobra.Command{
	Use:    "get <dev_id>",
	Short:  "Get Device properties",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		devID := args[0]
		reply, err := hub.GetDeviceProperties(devID)
		if err != nil {
			showError(cmd, "Cannot get device properties", err)
			os.Exit(1)
		}

		printDevicesProps(cmd, reply.GetProperties())
	},
}

var deviceSetDevPropsCmd = &cobra.Command{
	Use:    "set <dev_id> <props.yaml>",
	Short:  "Set Device properties",
	Args:   cobra.MinimumNArgs(2),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {

		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}
		workerID := args[0]
		propsFile := args[1]

		props, err := loadPropsFile(propsFile)
		if err != nil {
			showError(cmd, errCannotParsePropsFile.Error(), nil)
			os.Exit(1)
		}

		_, err = hub.SetDeviceProperties(workerID, props)
		if err != nil {
			showError(cmd, "Cannot set device properties", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}
