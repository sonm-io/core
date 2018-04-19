package commands

import (
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	masterRootCmd.AddCommand(
		masterListCmd,
		masterConfirmCmd,
		masterRemoveCmd,
	)
}

var masterRootCmd = &cobra.Command{
	Use:   "master",
	Short: "Manage master and workers addresses",
}

var masterListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show known worker's addresses",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		list, err := mm.WorkersList(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get workers list", err)
			os.Exit(1)
		}

		// todo: create printer
		showJSON(cmd, list)
	},
}

var masterConfirmCmd = &cobra.Command{
	Use:   "confirm <worker_eth>",
	Short: "Confirm pending Worker's registration request",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		worker := args[0]
		_, err = mm.WorkerConfirm(ctx, &pb.ID{Id: worker})
		if err != nil {
			showError(cmd, "Cannot approve Worker's request", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

var masterRemoveCmd = &cobra.Command{
	Use:   "remove <worker_eth>",
	Short: "Remove registered worker",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		mm, err := newMasterManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		worker := args[0]
		_, err = mm.WorkerRemove(ctx, &pb.ID{Id: worker})
		if err != nil {
			showError(cmd, "Cannot remove registered worker", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
