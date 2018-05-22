package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var (
	worker       pb.WorkerManagementClient
	workerCtx    context.Context
	workerCancel context.CancelFunc
)

func workerPreRun(cmd *cobra.Command, args []string) {
	loadKeyStoreIfRequired(cmd, args)
	workerCtx, workerCancel = newTimeoutContext()
	md := metadata.MD{
		"x_worker_eth_addr": []string{cfg.WorkerEthAddr.Hex()},
		"x_worker_net_addr": []string{cfg.WorkerNetAddr},
	}
	workerCtx = metadata.NewOutgoingContext(workerCtx, md)
	var err error
	worker, err = newWorkerManagementClient(workerCtx)
	if err != nil {
		showError(cmd, "Cannot create client connection", err)
		os.Exit(1)
	}
}

func workerPostRun(cmd *cobra.Command, args []string) {
	if workerCancel != nil {
		workerCancel()
	}
}

func init() {
	workerMgmtCmd.AddCommand(
		workerStatusCmd,
		askPlansRootCmd,
		workerTasksCmd,
		workerDevicesCmd,
	)
}

var workerMgmtCmd = &cobra.Command{
	Use:               "worker",
	Short:             "Worker management",
	PersistentPreRun:  workerPreRun,
	PersistentPostRun: workerPostRun,
}

var workerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show worker status",
	Run: func(cmd *cobra.Command, _ []string) {
		status, err := worker.Status(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get worker status", err)
			os.Exit(1)
		}

		printWorkerStatus(cmd, status)
	},
}
