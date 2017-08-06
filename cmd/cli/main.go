package main

import (
	b64 "encoding/base64"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

const (
	appName        = "sonm"
	hubAddressFlag = "addr"
	hubTimeoutFlag = "timeout"

	registryNameFlag     = "registry"
	registryUserFlag     = "user"
	registryPasswordFlag = "password"
)

var (
	gctx = context.Background()

	version          string
	hubAddress       string
	timeout          = 60 * time.Second
	registryName     string
	registryUser     string
	registryPassword string

	errMinerAddressRequired = errors.New("Miner address is required")
	errTaskIDRequired       = errors.New("Task ID is required")
	errImageNameRequired    = errors.New("Image name is required")
)

func checkHubAddressIsSet(cmd *cobra.Command, args []string) error {
	if cmd.Flag(hubAddressFlag).Value.String() == "" {
		return fmt.Errorf("--%s flag is required", hubAddressFlag)
	}
	return nil
}

func main() {
	// --- hub commands
	hubRootCmd := &cobra.Command{
		Use:     "hub",
		Short:   "Operations with hub",
		PreRunE: checkHubAddressIsSet,
	}

	hubPingCmd := &cobra.Command{
		Use:     "ping",
		Short:   "Ping the hub",
		PreRunE: checkHubAddressIsSet,
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				showError("Cannot create connection", err)
				return
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			_, err = pb.NewHubClient(cc).Ping(ctx, &pb.PingRequest{})
			if err != nil {
				showError("Ping failed", err)
				return
			}

			fmt.Printf("Ping hub %s... OK\r\n", hubAddress)
		},
	}

	hubStatusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Show hub status",
		PreRunE: checkHubAddressIsSet,
		Run: func(cmd *cobra.Command, args []string) {
			// todo: implement this on hub
			fmt.Printf("Hub %s status: OK\r\n", hubAddress)
		},
	}

	hubRootCmd.AddCommand(hubPingCmd, hubStatusCmd)

	// --- miner commands
	minerRootCmd := &cobra.Command{
		Use:     "miner",
		Short:   "Operations with miners",
		PreRunE: checkHubAddressIsSet,
	}

	minersListCmd := &cobra.Command{
		Use:     "list",
		Short:   "Show connected miners",
		PreRunE: minerRootCmd.PreRunE,
		Run: func(cmd *cobra.Command, args []string) {
			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				// fmt.Printf("Cannot create connection: %s\r\n", err)
				showError("Cannot create connection", err)
				return
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			lr, err := pb.NewHubClient(cc).List(ctx, &pb.ListRequest{})
			if err != nil {
				showError("Cannot get miners list", err)
				return
			}

			for addr, meta := range lr.Info {
				fmt.Printf("Miner: %s\r\n", addr)

				if len(meta.Values) == 0 {
					fmt.Println("Miner is idle")
				} else {
					fmt.Println("Tasks:")
					for i, task := range meta.Values {
						fmt.Printf("  %d) %s\r\n", i+1, task)
					}
				}
			}
		},
	}

	minerStatusCmd := &cobra.Command{
		Use:     "status <miner_addr>",
		Short:   "Miner status",
		PreRunE: checkHubAddressIsSet,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errMinerAddressRequired
			}
			minerID := args[0]

			conn, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				showError("Cannot create connection", err)
				return nil
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()

			var req = pb.HubInfoRequest{Miner: minerID}
			metrics, err := pb.NewHubClient(conn).Info(ctx, &req)
			if err != nil {
				showError("Cannot get miner status", err)
				return nil
			}

			if len(metrics.Stats) == 0 {
				fmt.Println("Miner is idle")
			} else {
				fmt.Println("Miner tasks:")
				for task, stat := range metrics.Stats {
					// fixme: what the hell with this ID?
					fmt.Printf("  ID: %s\r\n", task)
					fmt.Printf("      CPU: %d\r\n", stat.CPU.TotalUsage)
					fmt.Printf("      RAM: %s\r\n", humanize.Bytes(stat.Memory.MaxUsage))
				}
			}

			return nil
		},
	}

	minerRootCmd.AddCommand(minersListCmd, minerStatusCmd)

	// -- tasks commands
	tasksRootCmd := &cobra.Command{
		Use:     "task",
		Short:   "Manage tasks",
		PreRunE: checkHubAddressIsSet,
	}

	taskListCmd := &cobra.Command{
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

			var req = pb.HubStatusMapRequest{Miner: miner}
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
				fmt.Printf("  %s: %s\r\n", taskID, status.Status.String())
			}
			return nil
		},
	}

	taskStartCmd := &cobra.Command{
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
			var req = pb.HubStartTaskRequest{
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
	taskStartCmd.Flags().StringVar(&registryName, registryNameFlag, "", "Registry to pull image")
	taskStartCmd.Flags().StringVar(&registryUser, registryUserFlag, "", "Registry username")
	taskStartCmd.Flags().StringVar(&registryPassword, registryPasswordFlag, "", "Registry password")

	taskStatusCmd := &cobra.Command{
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

			fmt.Printf("Task %s (on %s) status is %s\n", req.Id, miner, taskStatus.Status.String())
			return nil
		},
	}

	taskStopCmd := &cobra.Command{
		Use:     "stop <miner_addr> <task_id>",
		Short:   "Stop task",
		PreRunE: checkHubAddressIsSet,
		RunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: always crash with
			// failed to stop the task 302e96de-5327-4bc2-97c0-2d56ce4d29c2
			if len(args) < 1 {
				return errMinerAddressRequired
			}
			if len(args) < 2 {
				return errTaskIDRequired
			}
			miner := args[0]
			taskID := args[1]

			fmt.Sprintf("Stopping task %s at %s...OK\r\n", taskID, miner)
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

			fmt.Println("OK")
			return nil
		},
	}

	tasksRootCmd.AddCommand(taskListCmd, taskStartCmd, taskStatusCmd, taskStopCmd)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\r\n", version)
		},
	}

	var rootCmd = &cobra.Command{Use: appName}
	rootCmd.PersistentFlags().StringVar(&hubAddress, hubAddressFlag, "", "hub addr")
	rootCmd.PersistentFlags().DurationVar(&timeout, hubTimeoutFlag, 60*time.Second, "Connection timeout")
	rootCmd.AddCommand(hubRootCmd, minerRootCmd, tasksRootCmd, versionCmd)
	rootCmd.Execute()
}

func encodeRegistryAuth(login, password string) string {
	data := fmt.Sprintf("%s:%s", login, password)
	return b64.StdEncoding.EncodeToString([]byte(data))
}

func showError(message string, err error) {
	if err != nil {
		fmt.Printf("[ERR] %s: %s\r\n", message, err.Error())
	} else {
		fmt.Printf("[ERR] %s\r\n", message)
	}
}
