package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
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
		addr, err := auth.NewAddr(args[0])
		if err != nil {
			showError(cmd, "Invalid address specified", err)
			os.Exit(1)
		}
		if _, err := addr.ETH(); err != nil {
			err = errors.New("could not parse eth component of the auth addr - it's malformed or missing")
			showError(cmd, "Invalid address specified", err)
			os.Exit(1)
		}
		cfg.WorkerAddr = addr.String()
		if err := cfg.Save(); err != nil {
			showError(cmd, "Failed to save worker address", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var workerCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current worker's addr",
	Run: func(cmd *cobra.Command, args []string) {
		type Result struct {
			address     string
			err         error
			description string
		}

		result := func() Result {
			result := Result{}
			if len(cfg.WorkerAddr) == 0 {
				result.description = "current worker is not set, using cli's addr"
				result.address = crypto.PubkeyToAddress(getDefaultKeyOrDie().PublicKey).Hex()
				return result
			}
			addr, err := auth.NewAddr(cfg.WorkerAddr)
			if err != nil {
				result.err = err
				result.description = fmt.Sprintf("current worker(%s) is invalid", cfg.WorkerAddr)
				return result
			}
			if _, err := addr.ETH(); err != nil {
				result.err = errors.New("could not parse eth component of the auth addr - it's malformed or missing")
				result.description = fmt.Sprintf("current worker(%s) is invalid", cfg.WorkerAddr)
				return result
			}
			result.description = "current worker is"
			result.address = addr.String()
			return result
		}()

		if result.err != nil {
			showError(cmd, result.description, result.err)
			os.Exit(1)
		}
		if isSimpleFormat() {
			cmd.Printf("%s %s\n", result.description, result.address)
		} else {
			showJSON(cmd, result)
		}

	},
}
