package commands

import (
	"fmt"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var workerTasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Show tasks running on Worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		list, err := worker.Tasks(workerCtx, &pb.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get task list: %v", err)
		}

		printNodeTaskStatus(cmd, list.GetInfo())
		return nil
	},
}
