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
		image := args[1]

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskStartCmdRunner(cmd, miner, image, itr)
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

func taskStartCmdRunner(cmd *cobra.Command, miner, image string, interactor CliInteractor) {
	var registryAuth string
	if registryUser != "" || registryPassword != "" || registryName != "" {
		registryAuth = encodeRegistryAuth(registryUser, registryPassword, registryName)
	}

	if isSimpleFormat() {
		cmd.Printf("Starting \"%s\" on miner %s...\r\n", image, miner)
	}

	var req = &pb.HubStartTaskRequest{
		Miner:    miner,
		Image:    image,
		Registry: registryName,
		Auth:     registryAuth,
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
