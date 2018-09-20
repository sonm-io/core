package commands

import (
	"fmt"

	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

func init() {
	askPlansRootCmd.AddCommand(
		askPlanListCmd,
		askPlanCreateCmd,
		askPlanRemoveCmd,
		askPlanPurgeCmd,
	)
}

var askPlansRootCmd = &cobra.Command{
	Use:   "ask-plan",
	Short: "Operations with ask order plan",
}

var askPlanListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current ask plans",
	RunE: func(cmd *cobra.Command, args []string) error {
		asks, err := worker.AskPlans(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot get Ask Orders from Worker: %v", err)
		}

		printAskList(cmd, asks)
		return nil
	},
}

var askPlanCreateCmd = &cobra.Command{
	Use:   "create <ask_plan.yaml>",
	Short: "Create new plan",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		planPath := args[0]
		plan := &sonm.AskPlan{}

		if err := task_config.LoadFromFile(planPath, plan); err != nil {
			return fmt.Errorf("cannot load AskPlan definition: %v", err)
		}

		id, err := worker.CreateAskPlan(workerCtx, plan)
		if err != nil {
			return fmt.Errorf("cannot create new AskOrder: %v", err)
		}

		printID(cmd, id.GetId())
		return nil
	},
}

var askPlanRemoveCmd = &cobra.Command{
	Use:   "remove <order_id>",
	Short: "Remove plan by",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := worker.RemoveAskPlan(workerCtx, &sonm.ID{Id: args[0]})
		if err != nil {
			return fmt.Errorf("cannot remove AskOrder: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var askPlanPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge all exiting ask-plans on worker",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := worker.PurgeAskPlans(workerCtx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("cannot purge ask plans: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
