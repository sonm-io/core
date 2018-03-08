package commands

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	hubACLRootCmd.AddCommand(
		nodeACLListCmd,
		nodeACLRegisterCmd,
		nodeACLDeregisterCmd,
	)
}

var hubACLRootCmd = &cobra.Command{
	Use:   "acl",
	Short: "Worker ACL management",
}

var nodeACLListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current ACLs",
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		list, err := hub.GetRegisteredWorkers()
		if err != nil {
			showError(cmd, "Cannot get Workers ACLs: %s", err)
			os.Exit(1)
		}

		printWorkerAclList(cmd, list)
	},
}

var nodeACLRegisterCmd = &cobra.Command{
	Use:   "register <worker_id>",
	Short: "Deregisters a worker credentials",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}
		id := args[0]

		_, err = hub.RegisterWorker(id)
		if err != nil {
			showError(cmd, "Cannot register new Worker", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var nodeACLDeregisterCmd = &cobra.Command{
	Use:   "deregister <worker_id>",
	Short: "Deregisters a worker credentials",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}
		id := args[0]

		_, err = hub.DeregisterWorker(id)
		if err != nil {
			showError(cmd, "Cannot deregister Worker", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}
