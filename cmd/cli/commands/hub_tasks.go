package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	hubTasksRootCmd.AddCommand(hubTaskListCmd, hubTaskStatusCmd)
}

var hubTasksRootCmd = &cobra.Command{
	Use:   "task",
	Short: "Operations with tasks",
}

var hubTaskListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show task list",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		list, err := hub.TaskList()
		if err != nil {
			showError(cmd, "Cannot get task list", err)
			os.Exit(1)
		}

		printNodeTaskStatus(cmd, list.GetInfo())
	},
}

var hubTaskStatusCmd = &cobra.Command{
	Use:    "status <task_id>",
	Short:  "Show task status",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		status, err := hub.TaskStatus(taskID)
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			os.Exit(1)
		}

		printTaskStatus(cmd, taskID, status)
	},
}
