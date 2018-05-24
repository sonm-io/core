package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pb "github.com/sonm-io/core/proto"
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
	Use:   "master",
	Short: "Manage master and workers addresses",
}

var masterListCmd = &cobra.Command{
	Use:    "list [master_eth]",
	Short:  "Show known worker's addresses",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		var master common.Address
		var err error
		if len(args) > 0 {
			master, err = util.HexToAddress(args[0])
			if err != nil {
				showError(cmd, "invalid address specified", err)
				os.Exit(1)
			}
		} else {
			key := getDefaultKeyOrDie()
			master = crypto.PubkeyToAddress(key.PublicKey)
		}

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		list, err := mm.WorkersList(ctx, pb.NewEthAddress(master))
		if err != nil {
			showError(cmd, "Cannot get workers list", err)
			os.Exit(1)
		}

		// todo: create printer
		showJSON(cmd, list)
	},
}

var masterConfirmCmd = &cobra.Command{
	Use:    "confirm <worker_eth>",
	Short:  "Confirm pending Worker's registration request",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		addr, err := util.HexToAddress(args[0])
		if err != nil {
			showError(cmd, "invalid address specified", err)
			os.Exit(1)
		}
		worker := pb.NewEthAddress(addr)
		_, err = mm.WorkerConfirm(ctx, worker)
		if err != nil {
			showError(cmd, "Cannot approve Worker's request", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

func masterRemove(cmd *cobra.Command, master common.Address, worker common.Address) {
	ctx, cancel := newTimeoutContext()
	defer cancel()

	mm, err := newMasterManagementClient(ctx)
	if err != nil {
		showError(cmd, "Cannot create client connection", err)
		os.Exit(1)
	}

	_, err = mm.WorkerRemove(ctx, &pb.WorkerRemoveRequest{
		Master: pb.NewEthAddress(master),
		Worker: pb.NewEthAddress(worker),
	})
	if err != nil {
		showError(cmd, "Cannot drop master - worker relationship", err)
		os.Exit(1)
	}

	showOk(cmd)
}

var masterRemoveWorkerCmd = &cobra.Command{
	Use:    "remove_worker <worker_eth>",
	Short:  "Remove registered worker",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		master := getDefaultKeyOrDie()

		worker, err := util.HexToAddress(args[0])
		if err != nil {
			showError(cmd, "invalid address specified", err)
			os.Exit(1)
		}
		masterRemove(cmd, crypto.PubkeyToAddress(master.PublicKey), worker)
	},
}

var masterRemoveMasterCmd = &cobra.Command{
	Use:    "remove_master <master_eth>",
	Short:  "Remove self from specified master",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		worker := getDefaultKeyOrDie()

		master, err := util.HexToAddress(args[0])
		if err != nil {
			showError(cmd, "invalid address specified", err)
			os.Exit(1)
		}
		masterRemove(cmd, master, crypto.PubkeyToAddress(worker.PublicKey))
	},
}
