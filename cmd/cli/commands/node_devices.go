package commands

import (
	"encoding/json"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	nodeDeviceRootCmd.AddCommand(
		nodeDeviceListCmd,
		nodeGetDevPropsCmd,
		nodeSetDevPropsCmd,
	)
}

func printDeviceList(cmd *cobra.Command, devices *pb.DevicesReply) {
	if isSimpleFormat() {
		CPUs := devices.GetCPUs()
		GPUs := devices.GetGPUs()

		if len(CPUs) == 0 && len(GPUs) == 0 {
			cmd.Printf("No devices detected.\r\n")
			return
		}

		if len(CPUs) > 0 {
			cmd.Printf("CPUs:\r\n")
			for id, cpu := range CPUs {
				cmd.Printf(" %s: %s\r\n", id, cpu.Device.ModelName)
			}
		} else {
			cmd.Printf("No CPUs detected.\r\n")
		}

		if len(GPUs) > 0 {
			cmd.Printf("GPUs:\r\n")
			for id, gpu := range GPUs {
				cmd.Printf(" %s: %s\r\n", id, gpu.Device.Name)
			}
		} else {
			cmd.Printf("No GPUs detected.\r\n")
		}
	} else {
		b, _ := json.Marshal(devices)
		cmd.Println(string(b))
	}
}

func printDevicesProps(cmd *cobra.Command, props map[string]float64) {
	if isSimpleFormat() {
		for k, v := range props {
			cmd.Printf("%s = %f\r\n", k, v)
		}
	} else {
		b, _ := json.Marshal(props)
		cmd.Println(string(b))
	}
}

var nodeDeviceRootCmd = &cobra.Command{
	Use:   "dev",
	Short: "Device properties",
}

var nodeDeviceListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show Hub's aggregated hardware",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress, timeout)
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

var nodeGetDevPropsCmd = &cobra.Command{
	Use:    "get <dev_id>",
	Short:  "Get Device properties",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress, timeout)
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

var nodeSetDevPropsCmd = &cobra.Command{
	Use:    "set <dev_id> <props.yaml>",
	Short:  "Set Device properties",
	Args:   cobra.MinimumNArgs(2),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {

		hub, err := NewHubInteractor(nodeAddress, timeout)
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
