package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/gosuri/uiprogress"
	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func init() {
	taskLogsCmd.Flags().StringVar(&logType, logTypeFlag, "both", "\"stdout\" or \"stderr\" or \"both\"")
	taskLogsCmd.Flags().StringVar(&since, sinceFlag, "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	taskLogsCmd.Flags().BoolVar(&addTimestamps, addTimestampsFlag, true, "Show timestamp for each log line")
	taskLogsCmd.Flags().BoolVar(&follow, followFlag, false, "Stream logs continuously")
	taskLogsCmd.Flags().StringVar(&tail, tailFlag, "50", "Number of lines to show from the end of the logs")
	taskLogsCmd.Flags().BoolVar(&details, detailsFlag, false, "Show extra details provided to logs")

	taskPullCmd.Flags().StringVar(&taskPullOutput, "output", "", "file to output")

	taskRootCmd.AddCommand(
		taskListCmd,
		taskStartCmd,
		taskStatusCmd,
		taskLogsCmd,
		taskStopCmd,
		taskPullCmd,
		taskPushCmd,
	)
}

var taskPullOutput string

var taskRootCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
}

var taskListCmd = &cobra.Command{
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

		printNodeTaskStatus(cmd, list.Info)
	},
}

var taskStartCmd = &cobra.Command{
	Use:    "start <deal_id> <task.yaml>",
	Short:  "Start task",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
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
			BuyerID: util.PubKeyToAddr(sessionKey.PublicKey).Hex(),
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

		printTaskStart(cmd, reply)
	},
}

var taskStatusCmd = &cobra.Command{
	Use:    "status <hub_addr> <task_id>",
	Short:  "Show task status",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		hubAddr := args[0]
		taskID := args[1]
		status, err := node.Status(taskID, hubAddr)
		if err != nil {
			showError(cmd, "Cannot get task status", err)
			os.Exit(1)
		}

		printTaskStatus(cmd, taskID, status)
	},
}

var taskLogsCmd = &cobra.Command{
	Use:    "logs <hub_addr> <task_id>",
	Short:  "Retrieve task logs",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		hubAddr := args[0]
		taskID := args[1]
		req := &pb.TaskLogsRequest{
			Id:            taskID,
			HubAddr:       hubAddr,
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

var taskStopCmd = &cobra.Command{
	Use:    "stop <hub_addr> <task_id>",
	Short:  "Stop task",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		hubAddr := args[0]
		taskID := args[1]
		_, err = node.Stop(taskID, hubAddr)
		if err != nil {
			showError(cmd, "Cannot stop status", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

var taskPullCmd = &cobra.Command{
	Use:    "pull <deal_id> <task_id>",
	Short:  "Pull committed image from the completed task.",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dealID := args[0]
		taskID := args[1]

		var wr io.Writer
		var err error
		if taskPullOutput == "" {
			wr = os.Stdout
		} else {
			file, err := os.Create(taskPullOutput)
			if err != nil {
				showError(cmd, "Cannot create file", err)
				os.Exit(1)
			}

			defer file.Close()
			wr = file
		}

		w := bufio.NewWriter(wr)

		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		client, err := node.ImagePull(dealID, taskID)
		if err != nil {
			showError(cmd, "Cannot create image pull client", err)
			os.Exit(1)
		}

		var bar *uiprogress.Bar
		var bytesRecv int64

		receivedSize := false
		streaming := true
		for streaming {
			chunk, err := client.Recv()
			if chunk != nil {
				if !receivedSize {

					header, err := client.Header()
					if err != nil {
						showError(cmd, "Cannot get client header", err)
						os.Exit(1)
					}

					size, err := structs.RequireHeaderInt64(header, "size")
					if err != nil {
						showError(cmd, "Cannot convert header value to int64", err)
						os.Exit(1)
					}

					if taskPullOutput != "" {
						uiprogress.Start()
						bar = uiprogress.AddBar(int(size))
						bar.PrependFunc(func(b *uiprogress.Bar) string {
							return fmt.Sprintf("Pushing %d/%d B)", bytesRecv, size)
						})
						bar.AppendCompleted()
					}
					receivedSize = true
				}

				n, err := io.Copy(wr, bytes.NewReader(chunk.Chunk))
				if err != nil {
					showError(cmd, "Cannot write to file", err)
					os.Exit(1)
				}

				bytesRecv += n
				if bar != nil {
					bar.Set(int(bytesRecv))
				}
			}

			if err != nil {
				if err == io.EOF {
					streaming = false
				} else {
					showError(cmd, "Streaming error", err)
					os.Exit(1)
				}
			}
		}

		if err := w.Flush(); err != nil {
			showError(cmd, "Cannot flush writer", err)
			os.Exit(1)
		}
	},
}

var taskPushCmd = &cobra.Command{
	Use:    "push <deal_id> <archive_path>",
	Short:  "Push an image from the filesystem",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dealID := args[0]
		path := args[1]

		file, err := os.Open(path)
		if err != nil {
			showError(cmd, "Cannot open archive path", err)
			os.Exit(1)
		}

		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			showError(cmd, "Cannot stat file", err)
			os.Exit(1)
		}

		node, err := NewTasksInteractor(nodeAddressFlag, timeoutFlag, WithWalletPerRPCCredentials())
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
			"deal": dealID,
			"size": strconv.FormatInt(fileInfo.Size(), 10),
		}))

		client, err := node.ImagePush(ctx)
		if err != nil {
			showError(cmd, "Cannot create push task client", err)
			os.Exit(1)
		}

		readCompleted := false
		bytesRemaining := int64(0)
		bytesCommitted := int64(0)

		uiprogress.Start()
		bar := uiprogress.AddBar(int(fileInfo.Size()))
		bar.PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Pushing %d/%d B)", bytesCommitted, fileInfo.Size())
		})
		bar.AppendCompleted()

		buf := make([]byte, 1*1024*1024)
		for {
			if !readCompleted {
				n, err := file.Read(buf)
				if err != nil {
					if err == io.EOF {
						readCompleted = true

						if err := client.CloseSend(); err != nil {
							showError(cmd, "Cannot close client stream", err)
							os.Exit(1)
						}
					} else {
						showError(cmd, "Cannot read file", err)
						os.Exit(1)
					}
				}

				if n > 0 {
					bytesRemaining = int64(n)
					client.Send(&pb.Chunk{Chunk: buf[:n]})
				}
			}

			for {
				progress, err := client.Recv()
				if err == io.EOF {
					if bytesCommitted == fileInfo.Size() {
						status, ok := client.Trailer()["status"]
						if !ok {
							showError(cmd, "No status returned", nil)
							os.Exit(1)
						}

						showJSON(cmd, map[string]interface{}{"status": status})
						return
					}
				}

				if err != nil {
					showError(cmd, "Cannot read from stream", nil)
					os.Exit(1)
				}

				bytesCommitted += progress.Size
				bytesRemaining -= progress.Size
				bar.Set(int(bytesCommitted))

				if bytesRemaining == 0 {
					break
				}
			}
		}
	},
}
