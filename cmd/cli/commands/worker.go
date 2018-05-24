package commands

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
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
	workerAddr := cfg.WorkerAddr
	if len(workerAddr) == 0 {
		workerAddr = crypto.PubkeyToAddress(getDefaultKeyOrDie().PublicKey).Hex()
	}
	md := metadata.MD{
		util.WorkerAddressHeader: []string{cfg.WorkerAddr},
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
		workerSwitchCmd,
		workerCurrentCmd,
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

var workerSwitchCmd = &cobra.Command{
	Use:   "switch <eth_addr>",
	Short: "Switch current worker to specified addr",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := util.HexToAddress(args[0])
		if err != nil {
			showError(cmd, "Invalid address specified", err)
			os.Exit(1)
		}
		cfg.WorkerAddr = addr.Hex()
		if err := cfg.Save(); err != nil {
			showError(cmd, "Failed to save worker address", err)
			os.Exit(1)
		}
	},
}

var workerCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current worker's addr",
	Run: func(cmd *cobra.Command, args []string) {
		if len(cfg.WorkerAddr) == 0 {
			addr := crypto.PubkeyToAddress(getDefaultKeyOrDie().PublicKey)
			cmd.Printf("current worker is not set, using cli's addr %s\n", addr.Hex())
		} else if !common.IsHexAddress(cfg.WorkerAddr) {
			cmd.Printf("current worker(%s) is invalid\n", cfg.WorkerAddr)
		} else {
			cmd.Printf("current worker is %s\n", cfg.WorkerAddr)
		}
	},
}
