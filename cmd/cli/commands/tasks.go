package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/sonm-io/core/proto/hub"
)

func init() {
	taskStartCmd.Flags().StringVar(&registryName, registryNameFlag, "", "Registry to pull image")
	taskStartCmd.Flags().StringVar(&registryUser, registryUserFlag, "", "Registry username")
	taskStartCmd.Flags().StringVar(&registryPassword, registryPasswordFlag, "", "Registry password")

	tasksRootCmd.AddCommand(taskListCmd, taskStartCmd, taskStatusCmd, taskStopCmd)
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
			showError("Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		var req = pb.MinerStatusRequest{Miner: miner}
		minerStatus, err := pb.NewHubClient(cc).MinerStatus(ctx, &req)
		if err != nil {
			showError("Cannot get tasks", err)
			return nil
		}

		if len(minerStatus.Statuses) == 0 {
			fmt.Printf("There is no tasks on miner \"%s\"\r\n", miner)
			return nil
		}

		fmt.Printf("There is %d tasks on miner \"%s\":\r\n", len(minerStatus.Statuses), miner)
		for taskID, status := range minerStatus.Statuses {
			fmt.Printf("  %s: %s\r\n", taskID, getMinerStatusByID(status))
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
			return errImageNameRequired
		}

		miner := args[0]
		image := args[1]

		var registryAuth string
		if registryUser != "" || registryPassword != "" {
			registryAuth = encodeRegistryAuth(registryUser, registryPassword)
		}

		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError("Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		var req = pb.StartTaskRequest{
			Miner:    miner,
			Image:    image,
			Registry: registryName,
			Auth:     registryAuth,
		}

		fmt.Printf("Starting \"%s\" on miner %s...\r\n", image, miner)
		rep, err := pb.NewHubClient(cc).StartTask(ctx, &req)
		if err != nil {
			showError("Cannot start task", err)
			return nil
		}

		fmt.Printf("ID %s, Endpoint %s\r\n", rep.Id, rep.Endpoint)
		return nil
	},
}

var taskStatusCmd = &cobra.Command{
	Use:     "status <miner_addr> <task_id>",
	Short:   "Show task status",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		// NOTE: always crash with
		// NotFound desc = no status report for task 302e96de-5327-4bc2-97c0-2d56ce4d29c2
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
			showError("Cannot create connection", err)
			return nil
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		var req = pb.TaskStatusRequest{Id: taskID}
		taskStatus, err := pb.NewHubClient(cc).TaskStatus(ctx, &req)
		if err != nil {
			showError("Cannot get task status", err)
			return nil
		}

		fmt.Printf("Task %s (on %s) status is %s\n", req.Id, miner, taskStatus.Status.Status.String())
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
			showError("Cannot create connection", err)
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
			showError("Cannot stop task", err)
			return nil
		}

		showOk()
		return nil
	},
}
