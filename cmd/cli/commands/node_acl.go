package commands

import (
	"os"

	"encoding/json"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	nodeACLRootCmd.AddCommand(
		nodeACLListCmd,
		nodeACLRegisterCmd,
		nodeACLUnregisterCmd,
	)
}

var nodeACLRootCmd = &cobra.Command{
	Use:     "acl",
	Short:   "Operations with Access Control Lists",
	PreRunE: checkNodeAddressIsSet,
}

func printWorkerAclList(cmd *cobra.Command, list *pb.GetRegistredWorkersReply) {
	if isSimpleFormat() {
		for i, id := range list.GetIds() {
			cmd.Printf("%d) %s\r\n", i+1, id.GetId())
		}

	} else {
		b, _ := json.Marshal(list)
		cmd.Printf("%s\r\n", string(b))
	}
}

var nodeACLListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show current ACLs",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		list, err := hub.GetRegistredWorkers()
		if err != nil {
			showError(cmd, "Cannot get Workers ACLs: %s", err)
			os.Exit(1)
		}

		printWorkerAclList(cmd, list)
	},
}

var nodeACLRegisterCmd = &cobra.Command{
	Use:     "register <worker_id>",
	Short:   "Register new Worker",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress)
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

var nodeACLUnregisterCmd = &cobra.Command{
	Use:     "unregister <worker_id>",
	Short:   "Unregister known worker",
	Args:    cobra.MinimumNArgs(1),
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}
		id := args[0]

		_, err = hub.UnregisterWorker(id)
		if err != nil {
			showError(cmd, "Cannot unregister Worker", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}
