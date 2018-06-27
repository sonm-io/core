package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v2"
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
		workerFreeDevicesCmd,
		workerSwitchCmd,
		workerCurrentCmd,
		workerScheduleMaintenanceCmd,
		workerNextMaintenanceCmd,
		workerDebugStateCmd,
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

var workerScheduleMaintenanceCmd = &cobra.Command{
	Use:   "maintenance <at or after>",
	Short: "Schedule worker maintanance",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var timePoint time.Time
		timeData := []byte(args[0])
		if err := timePoint.UnmarshalText(timeData); err != nil {
			duration, err := time.ParseDuration(args[0])
			if err != nil {
				showError(cmd, "Invalid time point or duration specified", err)
				os.Exit(1)
			}
			timePoint = time.Now()
			timePoint = timePoint.Add(duration)
		}

		if _, err := worker.ScheduleMaintenance(workerCtx, &pb.Timestamp{Seconds: timePoint.Unix()}); err != nil {
			showError(cmd, "failed to schedule maintenance", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var workerNextMaintenanceCmd = &cobra.Command{
	Use:   "next-maintenance",
	Short: "Print next scheduled maintenance",
	Run: func(cmd *cobra.Command, args []string) {
		next, err := worker.NextMaintenance(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "failed to get next maintenance", err)
			os.Exit(1)
		}
		if isSimpleFormat() {
			cmd.Println(next.Unix().String())
		} else {
			showJSON(cmd, next.Unix())
		}
	},
}

var workerCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current worker's addr",
	Run: func(cmd *cobra.Command, args []string) {
		type Result struct {
			Address     string `json:"address,omitempty"`
			Error       error  `json:"error,omitempty"`
			Description string `json:"description,omitempty"`
		}

		result := func() Result {
			result := Result{}
			if len(cfg.WorkerAddr) == 0 {
				result.Description = "current worker is not set, using cli's addr"
				result.Address = crypto.PubkeyToAddress(getDefaultKeyOrDie().PublicKey).Hex()
				return result
			}
			addr, err := auth.NewAddr(cfg.WorkerAddr)
			if err != nil {
				result.Error = err
				result.Description = fmt.Sprintf("current worker(%s) is invalid", cfg.WorkerAddr)
				return result
			}
			if _, err := addr.ETH(); err != nil {
				result.Error = errors.New("could not parse eth component of the auth addr - it's malformed or missing")
				result.Description = fmt.Sprintf("current worker(%s) is invalid", cfg.WorkerAddr)
				return result
			}
			result.Description = "current worker is"
			result.Address = addr.String()
			return result
		}()

		if result.Error != nil {
			showError(cmd, result.Description, result.Error)
			os.Exit(1)
		}
		if isSimpleFormat() {
			cmd.Printf("%s %s\n", result.Description, result.Address)
		} else {
			showJSON(cmd, result)
		}

	},
}

var workerDebugStateCmd = &cobra.Command{
	Use:   "debug-state",
	Short: "Provide some useful worker debugging info",
	Run: func(cmd *cobra.Command, args []string) {
		reply, err := worker.DebugState(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "failed to get debug state", err)
			os.Exit(1)
		}
		if isSimpleFormat() {
			data, err := yaml.Marshal(reply)
			if err != nil {
				showError(cmd, "failed to marshal state", err)
				os.Exit(1)
			}
			cmd.Printf("%s\r\n", string(data))
		} else {
			showJSON(cmd, reply)
		}

	},
}
