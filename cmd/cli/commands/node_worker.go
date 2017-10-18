package commands

import "github.com/spf13/cobra"

func init() {
	nodeWorkerRootCmd.AddCommand(
		nodeWorkerListCmd,
		nodeWorkerStatusCmd,
		nodeWorkerGetPropsCmd,
		nodeWorkerSetPropsCmd,
	)
}

var nodeWorkerRootCmd = &cobra.Command{
	Use:     "worker",
	Short:   "Operations with connected Workers",
	PreRunE: checkNodeAddressIsSet,
}

var nodeWorkerListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show connected workers list",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeWorkerStatusCmd = &cobra.Command{
	Use:     "status <worker_id>",
	Short:   "Show worker status",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeWorkerGetPropsCmd = &cobra.Command{
	Use:     "get-props <worker_id>",
	Short:   "Get resource properties from Worker",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeWorkerSetPropsCmd = &cobra.Command{
	Use:     "set-props <worker_id>",
	Short:   "Set resource properties for Worker",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
