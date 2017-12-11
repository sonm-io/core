package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	hubWorkerRootCmd.AddCommand(
		workerListCmd,
		workerStatusCmd,
	)
}

var hubWorkerRootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Operations with connected Workers",
}

var workerListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show connected workers list",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, _ []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
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

var workerStatusCmd = &cobra.Command{
	Use:    "status <worker_id>",
	Short:  "Show worker status",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
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
