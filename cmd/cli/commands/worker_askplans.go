package commands

import (
	"os"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	askPlansRootCmd.AddCommand(
		askPlanListCmd,
		askPlanCreateCmd,
		askPlanRemoveCmd,
	)
}

var askPlansRootCmd = &cobra.Command{
	Use:   "ask-plan",
	Short: "Operations with ask order plan",
}

var askPlanListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current ask plans",
	Run: func(cmd *cobra.Command, args []string) {
		asks, err := worker.AskPlans(workerCtx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get Ask Orders from Worker", err)
			os.Exit(1)
		}

		printAskList(cmd, asks)
	},
}

var askPlanCreateCmd = &cobra.Command{
	Use:   "create <ask_plan.yaml>",
	Short: "Create new plan",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		planPath := args[0]
		plan := &pb.AskPlan{}

		if err := task_config.LoadFromFile(planPath, plan); err != nil {
			showError(cmd, "Cannot load AskPlan definition", err)
			os.Exit(1)
		}

		id, err := worker.CreateAskPlan(workerCtx, plan)
		if err != nil {
			showError(cmd, "Cannot create new AskOrder", err)
			os.Exit(1)
		}

		printID(cmd, id.GetId())
	},
}

var askPlanRemoveCmd = &cobra.Command{
	Use:   "remove <order_id>",
	Short: "Remove plan by",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ID := args[0]
		_, err := worker.RemoveAskPlan(workerCtx, &pb.ID{Id: ID})
		if err != nil {
			showError(cmd, "Cannot remove AskOrder", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
