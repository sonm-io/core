package commands

import "github.com/spf13/cobra"

func init() {
	nodeTaskRootCmd.AddCommand(nodeTaskListCmd, nodeTaskStatusCmd)
}

var nodeTaskRootCmd = &cobra.Command{
	Use:     "task",
	Short:   "Operations with tasks",
	PreRunE: checkNodeAddressIsSet,
}

var nodeTaskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show task list",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeTaskStatusCmd = &cobra.Command{
	Use:     "status <task_id>",
	Short:   "Show task status",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
