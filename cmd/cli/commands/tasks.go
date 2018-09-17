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

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gosuri/uiprogress"
	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
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
		taskPurgeCmd,
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

func newDealContext(ctx context.Context, dealID string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"deal": dealID,
	}))
}

func getActiveDealIDs(ctx context.Context) ([]*big.Int, error) {
	dealCli, err := newDealsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot create client connection: %v", err)
	}

	deals, err := dealCli.List(ctx, &sonm.Count{Count: 0})
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

		var printer Printer = cmd
		if isSimpleFormat() {
			printer = &IndentPrinter{
				Subprinter: printer,
				IdentCount: 2,
				Ident:      ' ',
			}
		}
		for k, dealID := range dealIDs {
			timeoutCtx, cancel := context.WithTimeout(ctx, timeoutFlag)
			if isSimpleFormat() {
				cmd.Printf("Deal %s (%d/%d):\n", dealID.String(), k+1, len(dealIDs))
			}

			ctx := newDealContext(timeoutCtx, dealID.String())
			list, err := node.GetDealInfo(ctx, &sonm.ID{Id: dealID.String()})
			if err != nil {
				ShowError(printer, "cannot get task list for deal", err)
			} else {
				printNodeTaskStatus(printer, list)
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

		bigDealID, err := sonm.NewBigIntFromString(dealID)
		if err != nil {
			return err
		}

		request := &sonm.StartTaskRequest{
			DealID: bigDealID,
			Spec:   spec,
		}

		ctx = newDealContext(ctx, dealID)
		reply, err := node.StartTask(ctx, request)
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

		dealIDStr := args[0]
		// Check if passed value is valid bigint
		_, err := sonm.NewBigIntFromString(dealIDStr)
		if err != nil {
			return err
		}
		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		ctx = newDealContext(ctx, dealIDStr)

		taskID := args[1]
		req := &sonm.ID{
			Id: taskID,
		}

		status, err := node.TaskStatus(ctx, req)
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

		dealIDStr := args[0]
		// Check if passed dealID value is valid bigint
		_, err := sonm.NewBigIntFromString(dealIDStr)
		if err != nil {
			return err
		}

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		taskID := args[1]
		netID := args[2]
		ctx = newDealContext(ctx, dealIDStr)
		spec, err := node.JoinNetwork(ctx, &sonm.WorkerJoinNetworkRequest{
			TaskID:    taskID,
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

func parseType(logType string) (sonm.TaskLogsRequest_Type, error) {
	if len(logType) == 0 {
		return sonm.TaskLogsRequest_BOTH, nil
	}
	key := strings.ToUpper(logType)
	t, ok := sonm.TaskLogsRequest_Type_value[key]
	if !ok {
		return sonm.TaskLogsRequest_Type(0), fmt.Errorf("invalid log type %s", logType)
	}
	return sonm.TaskLogsRequest_Type(t), nil
}

var taskLogsCmd = &cobra.Command{
	Use:   "logs <deal_id> <task_id>",
	Short: "Retrieve task logs",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dealIDStr := args[0]
		dealID, err := util.ParseBigInt(dealIDStr)
		if err != nil {
			return err
		}
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

		logType, err := parseType(logType)
		if err != nil {
			return fmt.Errorf("failed to parse log type: %v", err)
		}

		req := &sonm.TaskLogsRequest{
			Type:          logType,
			Id:            args[1],
			DealID:        sonm.NewBigInt(dealID),
			Since:         since,
			AddTimestamps: addTimestamps,
			Follow:        follow,
			Tail:          tail,
			Details:       details,
		}

		ctx = newDealContext(ctx, dealIDStr)
		logClient, err := node.TaskLogs(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot get task logs: %v", err)
		}

		reader := sonm.NewLogReader(logClient)
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

		dealIDStr := args[0]
		_, err := sonm.NewBigIntFromString(dealIDStr)
		if err != nil {
			return err
		}

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		req := &sonm.ID{
			Id: args[1],
		}
		ctx = newDealContext(ctx, dealIDStr)

		if _, err := node.StopTask(ctx, req); err != nil {
			return fmt.Errorf("cannot stop task: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var taskPurgeCmd = &cobra.Command{
	Use:   "purge <deal_id>",
	Short: "Purge all tasks running on given deal",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		strID := args[0]
		id, err := sonm.NewBigIntFromString(strID)
		if err != nil {
			return err
		}

		node, err := newTaskClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		taskErr, err := node.PurgeTasks(newDealContext(ctx, strID), &sonm.PurgeTasksRequest{DealID: id})
		if err != nil {
			return fmt.Errorf("cannot purge tasks: %v", err)
		}

		printErrorByID(cmd, newTupleFromString(taskErr))
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

		req := &sonm.PullTaskRequest{
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
					client.Send(&sonm.Chunk{Chunk: buf[:n]})
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
