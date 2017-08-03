package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto/hub"
	"github.com/spf13/cobra"
)

const (
	appName        = "cli"
	hubAddressFlag = "addr"
	hubTimeoutFlag = "timeout"

	registryNameFlag     = "registry"
	registryUserFlag     = "user"
	registryPasswordFlag = "password"
)

var (
	gctx = context.Background()

	hubAddress       string
	timeout          = 60 * time.Second
	registryName     string
	registryUser     string
	registryPassword string

	errMinerIDRequired       = errors.New("Miner identifier is required")
	errTaskIDRequired        = errors.New("Task ID is required")
	errContainerNameRequired = errors.New("Container name is required")
)

func checkFlagRequired(cmd *cobra.Command, name string) error {
	if cmd.Flag(name).Value.String() == "" {
		return fmt.Errorf("--%s flag is required", name)
	}
	return nil
}

func main() {
	// --- hub commands
	hubRootCmd := &cobra.Command{
		Use:   "hub",
		Short: "Operations with hub",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkFlagRequired(cmd, hubAddressFlag)
		},
	}

	hubPingCmd := &cobra.Command{
		Use:     "ping",
		Short:   "Ping the hub",
		PreRunE: hubRootCmd.PostRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			_, err = pb.NewHubClient(cc).Ping(ctx, &pb.PingRequest{})
			if err != nil {
				return err
			}

			fmt.Printf("Ping hub %s... OK\r\n", hubAddress)
			return nil
		},
	}

	hubStatusCmd := &cobra.Command{
		Use:     "status",
		Short:   "Show hub status",
		PreRunE: hubRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// todo: implement this on hub
			fmt.Printf("Hub %s status: OK\r\n", hubAddress)
			return nil
		},
	}

	hubRootCmd.AddCommand(hubPingCmd, hubStatusCmd)

	// --- miner commands
	minerRootCmd := &cobra.Command{
		Use:   "miner",
		Short: "Operations with miners",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkFlagRequired(cmd, hubAddressFlag)
		},
	}

	minersListCmd := &cobra.Command{
		Use:     "list",
		Short:   "Show connected miners",
		PreRunE: minerRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			lr, err := pb.NewHubClient(cc).List(ctx, &pb.ListRequest{})
			if err != nil {
				return err
			}

			// todo(sshaman1101): pretty printing
			fmt.Println("Connected Miners")
			buff := new(bytes.Buffer)
			enc := json.NewEncoder(buff)
			enc.SetIndent("", "\t")
			enc.Encode(lr.Info)
			fmt.Printf("%s\n", buff.Bytes())
			return err
		},
	}

	minerStatusCmd := &cobra.Command{
		Use:     "status <miner_addr>",
		Short:   "Miner status",
		PreRunE: minerRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: always empty response
			if len(args) < 1 {
				return errMinerIDRequired
			}
			minerID := args[0]

			conn, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()

			var req = pb.InfoRequest{Miner: minerID}
			metrics, err := pb.NewHubClient(conn).Info(ctx, &req)
			if err != nil {
				return err
			}

			js, err := json.Marshal(metrics)
			if err != nil {
				return err
			}

			// TODO(sshaman1101): pretty printing
			fmt.Printf("%s", js)
			return nil
		},
	}

	minerRootCmd.AddCommand(minersListCmd, minerStatusCmd)

	// -- tasks commands
	tasksRootCmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkFlagRequired(cmd, hubAddressFlag)
		},
	}

	taskListCmd := &cobra.Command{
		Use:     "list <miner_addr>",
		Short:   "Show tasks on given miner",
		PreRunE: tasksRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: always return "null"
			if len(args) < 1 {
				return errMinerIDRequired
			}
			miner := args[0]

			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()

			var req = pb.MinerStatusRequest{Miner: miner}
			minerStatus, err := pb.NewHubClient(cc).MinerStatus(ctx, &req)
			if err != nil {
				fmt.Println(err)
				return err
			}

			buff := new(bytes.Buffer)
			enc := json.NewEncoder(buff)
			enc.SetIndent("", "\t")
			//TODO(sshaman1101): pretty printing
			enc.Encode(minerStatus.Statuses)
			fmt.Printf("%s\n", buff.Bytes())
			return nil
		},
	}

	taskStartCmd := &cobra.Command{
		Use:     "start <miner_addr> <image>",
		Short:   "Start task on given miner",
		PreRunE: tasksRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errMinerIDRequired
			}
			if len(args) < 2 {
				return errContainerNameRequired
			}

			miner := args[0]
			image := args[1]

			var registryAuth string
			if registryUser != "" || registryPassword != "" {
				registryAuth = encodeRegistryAuth(registryUser, registryPassword)
			}

			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
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

			fmt.Printf("Starting %s at %s...\r\n", image, miner)
			rep, err := pb.NewHubClient(cc).StartTask(ctx, &req)
			if err != nil {
				return err
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
		PreRunE: tasksRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: always crash with
			// NotFound desc = no status report for task 302e96de-5327-4bc2-97c0-2d56ce4d29c2
			if len(args) < 1 {
				return errMinerIDRequired
			}
			if len(args) < 2 {
				return errTaskIDRequired
			}
			miner := args[0]
			taskID := args[1]

			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()

			var req = pb.TaskStatusRequest{Id: taskID}
			taskStatus, err := pb.NewHubClient(cc).TaskStatus(ctx, &req)
			if err != nil {
				fmt.Println(err)
				return err
			}

			fmt.Printf("Task %s (on %s) status is %s\n", req.Id, miner, taskStatus.Status.Status.String())
			return nil
		},
	}

	taskStopCmd := &cobra.Command{
		Use:     "stop <miner_addr> <task_id>",
		Short:   "Stop task",
		PreRunE: tasksRootCmd.PreRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			// NOTE: always crash with
			// failed to stop the task 302e96de-5327-4bc2-97c0-2d56ce4d29c2
			if len(args) < 1 {
				return errMinerIDRequired
			}
			if len(args) < 2 {
				return errTaskIDRequired
			}
			miner := args[0]
			taskID := args[1]

			fmt.Sprintf("Stopping task %s at %s...OK\r\n", taskID, miner)
			cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
			if err != nil {
				return err
			}
			defer cc.Close()

			ctx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			var req = pb.StopTaskRequest{
				Id: taskID,
			}

			_, err = pb.NewHubClient(cc).StopTask(ctx, &req)
			if err != nil {
				fmt.Println(err)
				return err
			}

			fmt.Println("OK")
			return nil
		},
	}

	tasksRootCmd.AddCommand(taskListCmd, taskStartCmd, taskStatusCmd, taskStopCmd)

	var rootCmd = &cobra.Command{Use: appName}
	rootCmd.PersistentFlags().StringVar(&hubAddress, hubAddressFlag, "", "hub addr")
	rootCmd.PersistentFlags().DurationVar(&timeout, hubTimeoutFlag, 60*time.Second, "Connection timeout")
	rootCmd.AddCommand(hubRootCmd, minerRootCmd, tasksRootCmd)
	rootCmd.Execute()
}

func encodeRegistryAuth(login, password string) string {
	data := fmt.Sprintf("%s:%s", login, password)
	return b64.StdEncoding.EncodeToString([]byte(data))
}
