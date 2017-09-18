package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"time"

	"fmt"

	ds "github.com/c2h5oh/datasize"
	"github.com/docker/go-connections/nat"
	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"io"
	"strings"
)

func init() {
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
		cmd.Printf("ID %s\r\nEndpoint %s\r\n", rep.Id, rep.Endpoint)
	} else {
		b, _ := json.Marshal(rep)
		cmd.Println(string(b))
	}
}

func printTaskStatus(cmd *cobra.Command, id string, taskStatus *pb.TaskStatusReply) {
	if isSimpleFormat() {
		portsParsedOK := false
		ports := nat.PortMap{}
		if len(taskStatus.GetPorts()) > 0 {
			err := json.Unmarshal([]byte(taskStatus.GetPorts()), &ports)
			portsParsedOK = err == nil
		}

		cmd.Printf("Task %s (on %s):\r\n", id, taskStatus.MinerID)
		cmd.Printf("  Image:  %s\r\n", taskStatus.GetImageName())
		cmd.Printf("  Status: %s\r\n", taskStatus.GetStatus().String())
		cmd.Printf("  Uptime: %s\r\n", time.Duration(taskStatus.GetUptime()).String())

		if taskStatus.GetUsage() != nil {
			cmd.Println("  Resources:")
			cmd.Printf("    CPU: %d\r\n", taskStatus.Usage.GetCpu().GetTotal())
			cmd.Printf("    MEM: %s\r\n", ds.ByteSize(taskStatus.Usage.GetMemory().GetMaxUsage()).HR())
			if taskStatus.GetUsage().GetNetwork() != nil {
				cmd.Printf("    NET:\r\n")
				for i, net := range taskStatus.GetUsage().GetNetwork() {
					cmd.Printf("      %s:\r\n", i)
					cmd.Printf("        Tx/Rx bytes: %d/%d\r\n", net.TxBytes, net.RxBytes)
					cmd.Printf("        Tx/Rx packets: %d/%d\r\n", net.TxPackets, net.RxPackets)
					cmd.Printf("        Tx/Rx errors: %d/%d\r\n", net.TxErrors, net.RxErrors)
					cmd.Printf("        Tx/Rx dropped: %d/%d\r\n", net.TxDropped, net.RxDropped)
				}
			}
		}

		if portsParsedOK && len(ports) > 0 {
			cmd.Printf("  Ports:\r\n")
			for containerPort, host := range ports {
				if len(host) > 0 {
					cmd.Printf("    %s: %s:%s\r\n", containerPort, host[0].HostIP, host[0].HostPort)
				} else {
					cmd.Printf("    %s\r\n", containerPort)
				}
			}
		}
	} else {
		v := map[string]interface{}{
			"id":     id,
			"miner":  taskStatus.MinerID,
			"status": taskStatus.Status.String(),
			"image":  taskStatus.GetImageName(),
			"ports":  taskStatus.GetPorts(),
			"uptime": fmt.Sprintf("%d", time.Duration(taskStatus.GetUptime())),
		}
		if taskStatus.GetUsage() != nil {
			v["cpu"] = fmt.Sprintf("%d", taskStatus.GetUsage().GetCpu().GetTotal())
			v["mem"] = fmt.Sprintf("%d", taskStatus.GetUsage().GetMemory().GetMaxUsage())
			v["net"] = taskStatus.GetUsage().GetNetwork()
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

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect ot hub", err)
			return nil
		}

		taskLogCmdRunner(cmd, taskID, itr)
		return nil
	},
}

var taskStartCmd = &cobra.Command{
	Use:     "start <task.yaml>",
	Short:   "Start task on given miner",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errTaskFileRequired
		}

		taskFile := args[0]

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

		taskStartCmdRunner(cmd, taskDef, itr)
		return nil
	},
}

var taskStatusCmd = &cobra.Command{
	Use:     "status <task_id>",
	Short:   "Show task status",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errTaskIDRequired
		}
		taskID := args[0]

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		taskStatusCmdRunner(cmd, taskID, itr)
		return nil
	},
}

var taskStopCmd = &cobra.Command{
	Use:     "stop <task_id>",
	Short:   "Stop task",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errTaskIDRequired
		}
		taskID := args[0]

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

func taskLogCmdRunner(cmd *cobra.Command, taskID string, interactor CliInteractor) {
	req := &pb.TaskLogsRequest{

		Id:            taskID,
		Since:         since,
		AddTimestamps: addTimestamps,
		Follow:        follow,
		Tail:          tail,
		Details:       details,
	}

	logType, ok := pb.TaskLogsRequest_Type_value[strings.ToUpper(logType)]
	if !ok {
		showError(cmd, "Invalid log type", nil)
		return
	}
	req.Type = pb.TaskLogsRequest_Type(logType)

	client, err := interactor.TaskLogs(context.Background(), req)
	if err != nil {
		showError(cmd, "Cannot get task logs", err)
		return
	}

	for {
		buffer, err := client.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			showError(cmd, "IO failure during log fetching", err)
			return
		}
		cmd.Print(string(buffer.Data))
	}
}

func taskStartCmdRunner(cmd *cobra.Command, taskConfig task_config.TaskConfig, interactor CliInteractor) {
	if isSimpleFormat() {
		cmd.Printf("Starting \"%s\" ...\r\n", taskConfig.GetImageName())
	}

	requirements := &pb.TaskRequirements{
		Miners: []string{},
		Resources: &pb.TaskResourceRequirements{
			CPUCores:   taskConfig.GetCPUCount(),
			MaxMemory:  int64(taskConfig.GetRAMCount()),
			GPUSupport: taskConfig.GetGPURequirement(),
		},
	}

	var req = &pb.HubStartTaskRequest{
		Requirements:  requirements,
		Image:         taskConfig.GetImageName(),
		Registry:      taskConfig.GetRegistryName(),
		Auth:          taskConfig.GetRegistryAuth(),
		PublicKeyData: taskConfig.GetSSHKey(),
		Env:           taskConfig.GetEnvVars(),
	}

	rep, err := interactor.TaskStart(context.Background(), req)
	if err != nil {
		showError(cmd, "Cannot start task", err)
		return
	}

	printTaskStart(cmd, rep)
}

func taskStatusCmdRunner(cmd *cobra.Command, taskID string, interactor CliInteractor) {
	taskStatus, err := interactor.TaskStatus(context.Background(), taskID)
	if err != nil {
		showError(cmd, "Cannot get task status", err)
		return
	}

	printTaskStatus(cmd, taskID, taskStatus)
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
