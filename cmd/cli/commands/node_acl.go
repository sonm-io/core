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
		nodeACLDeregisterCmd,
	)
}

var nodeACLRootCmd = &cobra.Command{
	Use:   "acl",
	Short: "Worker ACL management",
}

func printWorkerAclList(cmd *cobra.Command, list *pb.GetRegisteredWorkersReply) {
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
	Use:    "list",
	Short:  "Show current ACLs",
	PreRun: loadKeyStoreWrapper,
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
	Use:    "register <worker_id>",
	Short:  "Deregisters a worker credentials",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
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
	Use:    "deregister <worker_id>",
	Short:  "Deregisters a worker credentials",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
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
