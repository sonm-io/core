package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	nodeTaskRootCmd.AddCommand(nodeTaskListCmd, nodeTaskStatusCmd)
}

var nodeTaskRootCmd = &cobra.Command{
	Use:     "task",
	Short:   "Operations with tasks",
	PreRunE: checkNodeAddressIsSet,
}

func printNodeTaskStatus(cmd *cobra.Command, tasksMap map[string]*pb.TaskListReply_TaskInfo) {
	if isSimpleFormat() {
		for worker, tasks := range tasksMap {
			if len(tasks.GetTasks()) == 0 {
				cmd.Printf("Worker \"%s\" has no tasks\r\n", worker)
				continue
			}

			cmd.Printf("Worker \"%s\":\r\n", worker)
			i := 1
			for ID, status := range tasks.GetTasks() {
				up := time.Duration(status.GetUptime())
				cmd.Printf("  %d) %s \r\n     %s  %s (up: %v)\r\n",
					i, ID, status.Status.String(), status.ImageName, up.String())
				i++
			}
		}
	} else {
		b, _ := json.Marshal(tasksMap)
		fmt.Printf("%s\r\n", string(b))
	}
}

var nodeTaskListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show task list",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddress)
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

var nodeTaskStatusCmd = &cobra.Command{
	Use:     "status <task_id>",
	Short:   "Show task status",
	PreRunE: checkNodeAddressIsSet,
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID := args[0]
		hub, err := NewHubInteractor(nodeAddress)
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
