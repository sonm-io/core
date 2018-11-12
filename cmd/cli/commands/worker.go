package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"gopkg.in/yaml.v2"
)

var (
	workerAddressFlag string
	worker            sonm.WorkerManagementClient
	workerCtx         context.Context
	workerCancel      context.CancelFunc
)

func workerPreRunE(cmd *cobra.Command, args []string) error {
	if err := loadKeyStoreWrapper(cmd, args); err != nil {
		return err
	}

	workerCtx, workerCancel = newTimeoutContext()
	workerAddr := cfg.WorkerAddr
	if len(workerAddressFlag) != 0 {
		workerAddr = workerAddressFlag
	}
	if len(workerAddr) == 0 {
		key, err := getDefaultKey()
		if err != nil {
			return err
		}
		workerAddr = crypto.PubkeyToAddress(key.PublicKey).Hex()
	}

	md := metadata.MD{
		util.WorkerAddressHeader: []string{workerAddr},
	}
	workerCtx = metadata.NewOutgoingContext(workerCtx, md)

	var err error
	worker, err = newWorkerManagementClient(workerCtx)
	if err != nil {
		return fmt.Errorf("cannot create client connection: %v", err)
	}

	return nil
}

func workerPostRun(_ *cobra.Command, _ []string) {
	if workerCancel != nil {
		workerCancel()
	}
}

func init() {
	workerMgmtCmd.PersistentFlags().StringVar(&workerAddressFlag, "worker-address", "", "Use specified worker address instead of configured value")
	workerMgmtCmd.AddCommand(
		workerStatusCmd,
		askPlansRootCmd,
		benchmarkRootCmd,
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
	PersistentPreRunE: workerPreRunE,
	PersistentPostRun: workerPostRun,
}

var workerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show worker status",
	RunE: func(cmd *cobra.Command, _ []string) error {
		status, err := worker.Status(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get worker status: %v", err)
		}

		printWorkerStatus(cmd, status)
		return nil
	},
}

var workerSwitchCmd = &cobra.Command{
	Use:   "switch <eth_addr>",
	Short: "Switch current worker to specified addr",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		addr, err := auth.NewAddr(args[0])
		if err != nil {
			return fmt.Errorf("invalid address specified: %v", err)
		}

		if _, err := addr.ETH(); err != nil {
			return fmt.Errorf("could not parse eth component of the auth addr - it's malformed or missing")
		}

		cfg.WorkerAddr = addr.String()
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save worker address: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var workerScheduleMaintenanceCmd = &cobra.Command{
	Use:   "maintenance <at or after>",
	Short: "Schedule worker maintanance",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var timePoint time.Time
		timeData := []byte(args[0])
		if err := timePoint.UnmarshalText(timeData); err != nil {
			duration, err := time.ParseDuration(args[0])
			if err != nil {
				return fmt.Errorf("invalid time point or duration specified: %v", err)
			}

			timePoint = time.Now()
			timePoint = timePoint.Add(duration)
		}

		if _, err := worker.ScheduleMaintenance(workerCtx, &sonm.Timestamp{Seconds: timePoint.Unix()}); err != nil {
			return fmt.Errorf("failed to schedule maintenance: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var workerNextMaintenanceCmd = &cobra.Command{
	Use:   "next-maintenance",
	Short: "Print next scheduled maintenance",
	RunE: func(cmd *cobra.Command, args []string) error {
		next, err := worker.NextMaintenance(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to get next maintenance: %v", err)
		}

		// todo: proper printer
		if isSimpleFormat() {
			cmd.Println(next.Unix().String())
		} else {
			showJSON(cmd, next.Unix())
		}

		return nil
	},
}

var workerCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current worker's addr",
	RunE: func(cmd *cobra.Command, args []string) error {
		type Result struct {
			Address     string `json:"address,omitempty"`
			Error       error  `json:"error,omitempty"`
			Description string `json:"description,omitempty"`
		}

		result := func() Result {
			result := Result{}
			if len(cfg.WorkerAddr) == 0 {
				key, err := getDefaultKey()
				if err != nil {
					result.Error = err
					return result
				}

				result.Description = "current worker is not set, using cli's addr"
				result.Address = crypto.PubkeyToAddress(key.PublicKey).Hex()
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
			return fmt.Errorf("%s: %v", result.Description, result.Error)
		}

		if isSimpleFormat() {
			cmd.Printf("%s %s\n", result.Description, result.Address)
		} else {
			showJSON(cmd, result)
		}
		return nil
	},
}

var workerDebugStateCmd = &cobra.Command{
	Use:   "debug-state",
	Short: "Provide some useful worker debugging info",
	RunE: func(cmd *cobra.Command, args []string) error {
		reply, err := worker.DebugState(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to get debug state: %v", err)
		}

		if isSimpleFormat() {
			data, err := yaml.Marshal(reply)
			if err != nil {
				return fmt.Errorf("failed to marshal state: %v", err)
			}

			cmd.Printf("%s\r\n", string(data))
		} else {
			showJSON(cmd, reply)
		}

		return nil
	},
}
