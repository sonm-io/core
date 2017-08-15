package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
)

func init() {
	tasksRootCmd.AddCommand(taskListCmd, taskStartCmd, taskStatusCmd, taskStopCmd)

	taskLogsCmd.Flags().StringVar(&logType, logTypeFlag, "both", "\"stdout\" or \"stderr\" or \"both\"")
	taskLogsCmd.Flags().StringVar(&since, sinceFlag, "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	taskLogsCmd.Flags().BoolVar(&addTimestamps, addTimestampsFlag, true, "Show timestamp for each log line")
	taskLogsCmd.Flags().BoolVar(&follow, followFlag, false, "Stream logs continuously")
	taskLogsCmd.Flags().StringVar(&tail, tailFlag, "50", "Number of lines to show from the end of the logs")
	taskLogsCmd.Flags().BoolVar(&details, detailsFlag, false, "Show extra details provided to logs")

	tasksRootCmd.AddCommand(taskListCmd, taskLogsCmd, taskStartCmd, taskStatusCmd, taskStopCmd)
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
		minerID := args[0]

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskListCmdRunner(cmd, minerID, itr)
		return nil
	},
}

var taskLogsCmd = &cobra.Command{
	Use:     "logs <task_id>",
	Short:   "Show task status",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errTaskIDRequired
		}
		taskID := args[0]
		req := pb.TaskLogsRequest{}
		if logType == "stderr" {
			req.Type = pb.TaskLogsRequest_STDERR
		} else if logType == "stdout" {
			req.Type = pb.TaskLogsRequest_STDOUT
		} else if logType == "both" {
			req.Type = pb.TaskLogsRequest_BOTH
		} else {
			showError(cmd, "Invalid log type", nil)
			return nil
		}
		req.Id = taskID
		req.Since = since
		req.AddTimestamps = addTimestamps
		req.Follow = follow
		req.Tail = tail
		req.Details = details

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError(cmd, "Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		client, err := pb.NewHubClient(cc).TaskLogs(ctx, &req)
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			return nil
		}
		for {
			buffer, err := client.Recv()
			if err != nil {
				return nil
			}
			cmd.Print(buffer.Data)
		}
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

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskStartCmdRunner(cmd, miner, taskDef, itr)
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
		minerID := args[0]
		taskID := args[1]

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskStatusCmdRunner(cmd, minerID, taskID, itr)
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

		taskID := args[1]

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskStopCmdRunner(cmd, taskID, itr)
		return nil
	},
}

func taskListCmdRunner(cmd *cobra.Command, minerID string, interactor CliInteractor) {
	minerStatus, err := interactor.TaskList(context.Background(), minerID)
	if err != nil {
		showError(cmd, "Cannot get tasks", err)
		return
	}

	printTaskList(cmd, minerStatus, minerID)
}

func taskStartCmdRunner(cmd *cobra.Command, miner string, taskConfig task_config.TaskConfig, interactor CliInteractor) {
	if isSimpleFormat() {
		cmd.Printf("Starting \"%s\" on miner %s...\r\n", taskConfig.GetImageName(), miner)
	}

	var req = &pb.HubStartTaskRequest{
		Miner:         miner,
		Image:         taskConfig.GetImageName(),
		Registry:      taskConfig.GetRegistryName(),
		Auth:          taskConfig.GetRegistryAuth(),
		PublicKeyData: taskConfig.GetSSHKey(),
	}

	rep, err := interactor.TaskStart(context.Background(), req)
	if err != nil {
		showError(cmd, "Cannot start task", err)
		return
	}

	printTaskStart(cmd, rep)
}

func taskStatusCmdRunner(cmd *cobra.Command, minerID, taskID string, interactor CliInteractor) {
	taskStatus, err := interactor.TaskStatus(context.Background(), taskID)
	if err != nil {
		showError(cmd, "Cannot get task status", err)
		return
	}

	printTaskStatus(cmd, minerID, taskID, taskStatus)
	return
}

func taskStopCmdRunner(cmd *cobra.Command, taskID string, interactor CliInteractor) {
	_, err := interactor.TaskStop(context.Background(), taskID)
	if err != nil {
		showError(cmd, "Cannot stop task", err)
		return
	}

	showOk(cmd)
}
