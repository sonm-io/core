package commands

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gosuri/uiprogress"
	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

func init() {
	taskLogsCmd.Flags().StringVar(&logType, logTypeFlag, "both", "\"stdout\" or \"stderr\" or \"both\"")
	taskLogsCmd.Flags().StringVar(&since, sinceFlag, "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	taskLogsCmd.Flags().BoolVar(&addTimestamps, addTimestampsFlag, true, "Show timestamp for each log line")
	taskLogsCmd.Flags().BoolVar(&follow, followFlag, false, "Stream logs continuously")
	taskLogsCmd.Flags().StringVar(&tail, tailFlag, "50", "Number of lines to show from the end of the logs")
	taskLogsCmd.Flags().BoolVar(&details, detailsFlag, false, "Show extra details provided to logs")
	taskLogsCmd.Flags().BoolVar(&prependStream, prependStreamFlag, false, "Show stream (stderr or stdout) for each line of logs")

	taskPullCmd.Flags().StringVar(&taskPullOutput, "output", "", "file to output")

	taskRootCmd.AddCommand(
		taskListCmd,
		taskStartCmd,
		taskStatusCmd,
		taskLogsCmd,
		taskStopCmd,
		taskPullCmd,
		taskPushCmd,
		taskJoinNetworkCmd,
	)
}

var taskPullOutput string

var taskRootCmd = &cobra.Command{
	Use:               "task",
	Short:             "Tasks management",
	PersistentPreRunE: loadKeyStoreWrapper,
}

func getActiveDealIDs(ctx context.Context) ([]*big.Int, error) {
	dealCli, err := newDealsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create client connection: %v", err)
	}

	deals, err := dealCli.List(ctx, &pb.Count{Count: 0})
	if err != nil {
		return nil, fmt.Errorf("cannot fetch deals list: %s", err)
	}

	dealIDs := make([]*big.Int, 0, len(deals.Deal))
	key, err := getDefaultKey()
	if err != nil {
		return nil, err
	}
	myAddr := crypto.PubkeyToAddress(key.PublicKey).Big()

	for _, deal := range deals.Deal {
		// append active deal id only if current user is supplier
		iamConsumer := deal.GetConsumerID().Unwrap().Big().Cmp(myAddr) == 0
		if iamConsumer {
			dealIDs = append(dealIDs, deal.GetId().Unwrap())
		}
	}
	return dealIDs, nil
}

var taskListCmd = &cobra.Command{
	Use:   "list [deal_id]",
	Short: "Show active tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		var dealIDs []*big.Int
		if len(args) > 0 {
			dealID, err := util.ParseBigInt(args[0])
			if err != nil {
				return nil
			}
			dealIDs = append(dealIDs, dealID)
		} else {
			if !isSimpleFormat() {
				return fmt.Errorf("listing task for all deals is prohibited in JSON mode")
			}

			cmd.Printf("fetching deals ...\n")
			dealIDs, err = getActiveDealIDs(ctx)
			if err != nil {
				return err
			}
		}

		if len(dealIDs) == 0 {
			if isSimpleFormat() {
				cmd.Println("No active deals found.")
			}
			return nil
		}

		for k, dealID := range dealIDs {
			timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
			if isSimpleFormat() {
				cmd.Printf("Deal %s (%d/%d):\n", dealID.String(), k+1, len(dealIDs))
			}

			list, err := node.List(timeoutCtx, &pb.TaskListRequest{DealID: pb.NewBigInt(dealID)})
			if err != nil {
				ShowError(cmd, "cannot get task list for deal", err)
			} else {
				printNodeTaskStatus(cmd, list.GetInfo())
			}
			cancel()
		}

		return nil
	},
}

var taskStartCmd = &cobra.Command{
	Use:   "start <deal_id> <task.yaml>",
	Short: "Start task",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		dealID := args[0]
		taskFile := args[1]
		spec, err := task_config.LoadConfig(taskFile)
		if err != nil {
			return fmt.Errorf("cannot load task definition: %v", err)
		}

		bigDealID, err := pb.NewBigIntFromString(dealID)
		if err != nil {
			return err
		}

		request := &pb.StartTaskRequest{
			DealID: bigDealID,
			Spec:   spec,
		}

		reply, err := node.Start(ctx, request)
		if err != nil {
			return fmt.Errorf("cannot start task: %v", err)
		}

		printTaskStart(cmd, reply)
		return nil
	},
}

var taskStatusCmd = &cobra.Command{
	Use:   "status <deal_id> <task_id>",
	Short: "Show task status",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		dealID, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			return err
		}

		taskID := args[1]
		req := &pb.TaskID{
			Id:     taskID,
			DealID: dealID,
		}

		status, err := node.Status(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot get task status: %v", err)
		}

		printTaskStatus(cmd, taskID, status)
		return nil
	},
}

var taskJoinNetworkCmd = &cobra.Command{
	Use:   "join <deal_id> <task_id> <network_id>",
	Short: "Provide network specs for joining to specified task's specific network",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		dealID, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			return err
		}

		taskID := args[1]
		netID := args[2]
		spec, err := node.JoinNetwork(ctx, &pb.JoinNetworkRequest{
			TaskID: &pb.TaskID{
				Id:     taskID,
				DealID: dealID,
			},
			NetworkID: netID,
		})
		if err != nil {
			return fmt.Errorf("cannot get task status: %v", err)
		}

		printNetworkSpec(cmd, spec)
		return nil
	},
}

type logWriter struct {
	writer io.Writer
	prefix string
}

func (m *logWriter) Write(p []byte) (n int, err error) {
	if len(m.prefix) > 0 {
		if _, err := m.writer.Write([]byte(m.prefix)); err != nil {
			return 0, err
		}
	}
	return m.writer.Write(p)
}

func parseType(logType string) (pb.TaskLogsRequest_Type, error) {
	if len(logType) == 0 {
		return pb.TaskLogsRequest_BOTH, nil
	}
	key := strings.ToUpper(logType)
	t, ok := pb.TaskLogsRequest_Type_value[key]
	if !ok {
		return pb.TaskLogsRequest_Type(0), fmt.Errorf("invalid log type %s", logType)
	}
	return pb.TaskLogsRequest_Type(t), nil
}

var taskLogsCmd = &cobra.Command{
	Use:   "logs <deal_id> <task_id>",
	Short: "Retrieve task logs",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var ctx context.Context
		var cancel context.CancelFunc
		if follow {
			ctx, cancel = context.WithCancel(context.Background())
		} else {
			ctx, cancel = newTimeoutContext()
		}
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		dealID, err := util.ParseBigInt(args[0])
		if err != nil {
			return err
		}

		logType, err := parseType(logType)
		if err != nil {
			return fmt.Errorf("failed to parse log type: %v", err)
		}

		req := &pb.TaskLogsRequest{
			Type:          logType,
			Id:            args[1],
			DealID:        pb.NewBigInt(dealID),
			Since:         since,
			AddTimestamps: addTimestamps,
			Follow:        follow,
			Tail:          tail,
			Details:       details,
		}

		logClient, err := node.Logs(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot get task logs: %v", err)
		}

		reader := pb.NewLogReader(logClient)
		stdout := cmd.OutOrStdout()
		stderr := cmd.OutOrStderr()
		if prependStream {
			stdout = &logWriter{stdout, "[STDOUT] "}
			stderr = &logWriter{stderr, "[STDERR] "}
		}

		if _, err := stdcopy.StdCopy(stdout, stderr, reader); err != nil {
			return fmt.Errorf("failed to read logs: %v", err)
		}

		return nil
	},
}

var taskStopCmd = &cobra.Command{
	Use:   "stop <deal_id> <task_id>",
	Short: "Stop task",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		dealID, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			return nil
		}

		req := &pb.TaskID{
			Id:     args[1],
			DealID: dealID,
		}

		if _, err := node.Stop(ctx, req); err != nil {
			return fmt.Errorf("cannot stop status: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var taskPullCmd = &cobra.Command{
	Use:   "pull <deal_id> <task_id>",
	Short: "Pull committed image from the completed task.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dealID := args[0]
		taskID := args[1]

		var wr io.Writer
		var err error
		if taskPullOutput == "" {
			wr = os.Stdout
		} else {
			file, err := os.Create(taskPullOutput)
			if err != nil {
				return fmt.Errorf("cannot create file: %v", err)
			}

			defer file.Close()
			wr = file
		}

		w := bufio.NewWriter(wr)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		req := &pb.PullTaskRequest{
			DealId: dealID,
			TaskId: taskID,
		}

		client, err := node.PullTask(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot create image pull client: %v", err)
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
						return fmt.Errorf("cannot get client header: %v", err)
					}

					size, err := structs.RequireHeaderInt64(header, "size")
					if err != nil {
						return fmt.Errorf("cannot convert header value to int64: %v", err)
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
					return fmt.Errorf("cannot write to file: %v", err)
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
					return fmt.Errorf("streaming error: %v", err)
				}
			}
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("cannot flush writer: %v", err)
		}

		return nil
	},
}

var taskPushCmd = &cobra.Command{
	Use:   "push <deal_id> <archive_path>",
	Short: "Push an image from the filesystem",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dealID := args[0]
		path := args[1]

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("сannot open archive path: %v", err)
		}

		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("cannot stat file: %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"deal": dealID,
			"size": strconv.FormatInt(fileInfo.Size(), 10),
		}))

		client, err := node.PushTask(ctx)
		if err != nil {
			return fmt.Errorf("cannot create push task client: %v", err)
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
							return fmt.Errorf("cannot close client stream: %v", err)
						}
					} else {
						return fmt.Errorf("cannot read file: %v", err)
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
						id, ok := client.Trailer()["id"]
						if !ok || len(id) == 0 {
							return fmt.Errorf("no status returned: %v", nil)
						}

						printID(cmd, id[0])
						return nil
					}
				}

				if err != nil {
					return fmt.Errorf("cannot read from stream: %v", err)
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
