package commands

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"encoding/json"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
)

func init() {
	tasksRootCmd.AddCommand(taskListCmd, taskStartCmd, taskStatusCmd, taskStopCmd)
}

func printTaskList(cmd *cobra.Command, minerStatus *pb.StatusMapReply, miner string) {
	if isSimpleFormat() {
		if len(minerStatus.Statuses) == 0 {
			cmd.Printf("There is no tasks on miner \"%s\"\r\n", miner)
			return
		}

		cmd.Printf("There is %d tasks on miner \"%s\":\r\n", len(minerStatus.Statuses), miner)
		for taskID, status := range minerStatus.Statuses {
			cmd.Printf("  %s: %s\r\n", taskID, status.GetStatus())
		}
	} else {
		b, _ := json.Marshal(minerStatus)
		cmd.Println(string(b))
	}
}

func printTaskStart(cmd *cobra.Command, rep *pb.HubStartTaskReply) {
	if isSimpleFormat() {
		cmd.Printf("ID %s, Endpoint %s\r\n", rep.Id, rep.Endpoint)
	} else {
		b, _ := json.Marshal(rep)
		cmd.Println(string(b))
	}
}

func printTaskStatus(cmd *cobra.Command, miner, id string, taskStatus *pb.TaskStatusReply) {
	if isSimpleFormat() {
		cmd.Printf("Task %s (on %s) status is %s\n", id, miner, taskStatus.Status.String())
	} else {
		v := map[string]string{
			"id":     id,
			"miner":  miner,
			"status": taskStatus.Status.String(),
		}
		b, _ := json.Marshal(v)
		cmd.Println(string(b))
	}
}

var tasksRootCmd = &cobra.Command{
	Use:     "task",
	Short:   "Manage tasks",
	PreRunE: checkHubAddressIsSet,
}

var taskListCmd = &cobra.Command{
	Use:     "list <miner_addr>",
	Short:   "Show tasks on given miner",
	PreRunE: tasksRootCmd.PreRunE,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMinerAddressRequired
		}
		miner := args[0]

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError(cmd, "Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		var req = pb.HubStatusMapRequest{Miner: miner}
		minerStatus, err := pb.NewHubClient(cc).MinerStatus(ctx, &req)
		if err != nil {
			showError(cmd, "Cannot get tasks", err)
			return nil
		}

		printTaskList(cmd, minerStatus, miner)
		return nil
	},
}

var taskStartCmd = &cobra.Command{
	Use:     "start <miner_addr> <image>",
	Short:   "Start task on given miner",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMinerAddressRequired
		}
		if len(args) < 2 {
			return errTaskFileRequired
		}

		miner := args[0]
		taskFile := args[1]

		taskDef, err := task_config.LoadConfig(taskFile)
		if err != nil {
			showError(cmd, "Cannot load task definition", err)
			return nil
		}

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError(cmd, "Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		var req = pb.HubStartTaskRequest{
			Miner:    miner,
			Image:    taskDef.GetImageName(),
			Registry: taskDef.GetRegistryName(),
			Auth:     taskDef.GetRegistryAuth(),
		}

		if isSimpleFormat() {
			cmd.Printf("Starting \"%s\" on miner %s...\r\n", taskDef.GetImageName(), miner)
		}

		rep, err := pb.NewHubClient(cc).StartTask(ctx, &req)
		if err != nil {
			showError(cmd, "Cannot start task", err)
			return nil
		}

		printTaskStart(cmd, rep)
		return nil
	},
}

var taskStatusCmd = &cobra.Command{
	Use:     "status <miner_addr> <task_id>",
	Short:   "Show task status",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMinerAddressRequired
		}
		if len(args) < 2 {
			return errTaskIDRequired
		}
		miner := args[0]
		taskID := args[1]

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError(cmd, "Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		var req = pb.TaskStatusRequest{Id: taskID}
		taskStatus, err := pb.NewHubClient(cc).TaskStatus(ctx, &req)
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			return nil
		}

		printTaskStatus(cmd, miner, taskID, taskStatus)
		return nil
	},
}

var taskStopCmd = &cobra.Command{
	Use:     "stop <miner_addr> <task_id>",
	Short:   "Stop task",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMinerAddressRequired
		}
		if len(args) < 2 {
			return errTaskIDRequired
		}
		// miner := args[0]
		taskID := args[1]

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError(cmd, "Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		var req = pb.StopTaskRequest{
			Id: taskID,
		}

		_, err = pb.NewHubClient(cc).StopTask(ctx, &req)
		if err != nil {
			showError(cmd, "Cannot stop task", err)
			return nil
		}

		showOk(cmd)
		return nil
	},
}
