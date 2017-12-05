package commands

import (
	"os"

	"io"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	nodeTaskRootCmd.AddCommand(
		nodeTaskListCmd,
		nodeTaskStartCmd,
		nodeTaskStatusCmd,
		nodeTaskLogsCmd,
		nodeTaskStopCmd,
	)
}

var nodeTaskRootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var nodeTaskListCmd = &cobra.Command{
	Use:    "list [hub_addr]",
	Short:  "Show active tasks",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		var hubAddr string
		if len(args) > 0 {
			hubAddr = args[0]
		}

		list, err := node.List(hubAddr)
		if err != nil {
			showError(cmd, "Cannot get task list", err)
			os.Exit(1)
		}

		showJSON(cmd, list)
	},
}

var nodeTaskStartCmd = &cobra.Command{
	Use:    "start <deal_id> <task.yaml>",
	Short:  "Start task",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		dealID := args[0]
		taskFile := args[1]

		taskDef, err := task_config.LoadConfig(taskFile)
		if err != nil {
			showError(cmd, "Cannot load task definition", err)
			os.Exit(1)
		}

		deal := &pb.Deal{
			Id:      dealID,
			BuyerID: util.PubKeyToAddr(sessionKey.PublicKey),
		}

		var req = &pb.HubStartTaskRequest{
			Deal:          deal,
			Image:         taskDef.GetImageName(),
			Registry:      taskDef.GetRegistryName(),
			Auth:          taskDef.GetRegistryAuth(),
			PublicKeyData: taskDef.GetSSHKey(),
			Env:           taskDef.GetEnvVars(),
		}

		reply, err := node.Start(req)
		if err != nil {
			showError(cmd, "Cannot start task", err)
			os.Exit(1)
		}

		showJSON(cmd, reply)
	},
}

var nodeTaskStatusCmd = &cobra.Command{
	Use:    "status <task_id>",
	Short:  "Show task status",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		taskID := args[0]
		status, err := node.Status(taskID)
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			os.Exit(1)
		}

		showJSON(cmd, status)
	},
}

var nodeTaskLogsCmd = &cobra.Command{
	Use:    "logs <task_id>",
	Short:  "Retrieve task logs",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		taskID := args[0]
		req := &pb.TaskLogsRequest{
			Id:            taskID,
			Since:         since,
			AddTimestamps: addTimestamps,
			Follow:        follow,
			Tail:          tail,
			Details:       details,
		}

		logClient, err := node.Logs(req)
		if err != nil {
			showError(cmd, "Cannot get task logs", err)
			os.Exit(1)
		}

		for {
			chunk, err := logClient.Recv()
			if err == io.EOF {
				return
			}

			if err != nil {
				if err != nil {
					showError(cmd, "Cannot fetch log chunk", err)
					os.Exit(1)
				}
			}

			cmd.Print(string(chunk.Data))
		}
	},
}

var nodeTaskStopCmd = &cobra.Command{
	Use:    "stop <task_id>",
	Short:  "Stop task",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		taskID := args[0]
		status, err := node.Stop(taskID)
		if err != nil {
			showError(cmd, "Cannot stop status", err)
			os.Exit(1)
		}

		showJSON(cmd, status)
	},
}
