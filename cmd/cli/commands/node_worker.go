package commands

import (
	"encoding/json"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	nodeWorkerRootCmd.AddCommand(
		nodeWorkerListCmd,
		nodeWorkerStatusCmd,
		nodeWorkerGetPropsCmd,
		nodeWorkerSetPropsCmd,
	)
}

func printWorkerProps(cmd *cobra.Command, props map[string]string) {
	if isSimpleFormat() {
		for k, v := range props {
			cmd.Printf("%s = %s\r\n", k, v)
		}
	} else {
		b, _ := json.Marshal(props)
		cmd.Println(string(b))
	}
}

var nodeWorkerRootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Operations with connected Workers",
}

var nodeWorkerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show connected workers list",
	Run: func(cmd *cobra.Command, _ []string) {
		hub, err := NewHubInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		list, err := hub.WorkersList()
		if err != nil {
			showError(cmd, "Cannot get workers list", err)
			os.Exit(1)
		}

		printWorkerList(cmd, list)
	},
}

var nodeWorkerStatusCmd = &cobra.Command{
	Use:   "status <worker_id>",
	Short: "Show worker status",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		workerID := args[0]
		status, err := hub.WorkerStatus(workerID)
		if err != nil {
			showError(cmd, "Cannot get workers status", err)
			os.Exit(1)
		}

		printWorkerStatus(cmd, workerID, status)
	},
}

var nodeWorkerGetPropsCmd = &cobra.Command{
	Use:   "get-props <worker_id>",
	Short: "Get resource properties from Worker",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		workerID := args[0]
		props, err := hub.GetWorkerProperties(workerID)
		if err != nil {
			showError(cmd, "Cannot get workers status", err)
			os.Exit(1)
		}

		printWorkerProps(cmd, props.Properties)
	},
}

var nodeWorkerSetPropsCmd = &cobra.Command{
	Use:   "set-props <worker_id> <props.yaml>",
	Short: "Set resource properties for Worker",
	Args:  cobra.MinimumNArgs(2),
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

		req := &pb.SetMinerPropertiesRequest{
			ID:         workerID,
			Properties: props,
		}

		_, err = hub.SetWorkerProperties(req)
		if err != nil {
			showError(cmd, "Cannot get workers status", err)
			os.Exit(1)
		}
	},
}
