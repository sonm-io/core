package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
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
	Use:   "list",
	Short: "Show task list",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		list, err := hub.TaskList(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get task list", err)
			os.Exit(1)
		}

		printNodeTaskStatus(cmd, list.GetInfo())
	},
}

var hubTaskStatusCmd = &cobra.Command{
	Use:   "status <task_id>",
	Short: "Show task status",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		taskID := args[0]
		status, err := hub.TaskStatus(ctx, &pb.ID{Id: taskID})
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			os.Exit(1)
		}

		printTaskStatus(cmd, taskID, status)
	},
}
