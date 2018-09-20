package commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	masterRootCmd.AddCommand(
		masterListCmd,
		masterConfirmCmd,
		masterRemoveWorkerCmd,
		masterRemoveMasterCmd,
	)
}

var masterRootCmd = &cobra.Command{
	Use:               "master",
	Short:             "Manage master and workers addresses",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var masterListCmd = &cobra.Command{
	Use:   "list [master_eth]",
	Short: "Show known worker's addresses",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		var master common.Address
		var err error
		if len(args) > 0 {
			master, err = util.HexToAddress(args[0])
			if err != nil {
				return err
			}
		} else {
			key, err := getDefaultKey()
			if err != nil {
				return err
			}

			master = crypto.PubkeyToAddress(key.PublicKey)
		}

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		list, err := mm.WorkersList(ctx, sonm.NewEthAddress(master))
		if err != nil {
			return fmt.Errorf("cannot get workers list: %v", err)
		}

		printWorkersList(cmd, list)
		return nil
	},
}

var masterConfirmCmd = &cobra.Command{
	Use:   "confirm <worker_eth>",
	Short: "Confirm pending Worker's registration request",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		addr, err := util.HexToAddress(args[0])
		if err != nil {
			return fmt.Errorf("invalid address specified: %v", err)
		}
		worker := sonm.NewEthAddress(addr)
		_, err = mm.WorkerConfirm(ctx, worker)
		if err != nil {
			return fmt.Errorf("cannot approve Worker's request: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

func masterRemove(master common.Address, worker common.Address) error {
	ctx, cancel := newTimeoutContext()
	defer cancel()

	mm, err := newMasterManagementClient(ctx)
	if err != nil {
		return fmt.Errorf("cannot create client connection: %v", err)
	}

	_, err = mm.WorkerRemove(ctx, &sonm.WorkerRemoveRequest{
		Master: sonm.NewEthAddress(master),
		Worker: sonm.NewEthAddress(worker),
	})
	if err != nil {
		return fmt.Errorf("cannot drop master-worker relationship: %v", err)
	}

	return nil
}

var masterRemoveWorkerCmd = &cobra.Command{
	Use:   "remove_worker <worker_eth>",
	Short: "Remove registered worker",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		master, err := getDefaultKey()
		if err != nil {
			return err
		}

		worker, err := util.HexToAddress(args[0])
		if err != nil {
			return err
		}

		if err := masterRemove(crypto.PubkeyToAddress(master.PublicKey), worker); err != nil {
			return err
		}

		showOk(cmd)
		return nil
	},
}

var masterRemoveMasterCmd = &cobra.Command{
	Use:   "remove_master <master_eth>",
	Short: "Remove self from specified master",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		worker, err := getDefaultKey()
		if err != nil {
			return err
		}

		master, err := util.HexToAddress(args[0])
		if err != nil {
			return err
		}

		if err := masterRemove(master, crypto.PubkeyToAddress(worker.PublicKey)); err != nil {
			return err
		}

		showOk(cmd)
		return nil
	},
}
