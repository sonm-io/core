package commands

import (
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerTasksCmd = &cobra.Command{
	Use:    "tasks",
	Short:  "Show tasks running on Worker",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		list, err := worker.Tasks(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get task list", err)
			os.Exit(1)
		}

		printNodeTaskStatus(cmd, list.GetInfo())
	},
}
