package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"

	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/util"
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
		ctx := context.Background()
		hub, err := newWorkerManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		asks, err := hub.AskPlans(ctx, &pb.Empty{})
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
		ctx := context.Background()
		hub, err := newWorkerManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		planPath := args[0]

		plan, err := task_config.LoadAskPlan(planPath)
		if err != nil {
			showError(cmd, "Cannot load AskPlan definition", err)
			os.Exit(1)
		}

		id, err := hub.CreateAskPlan(ctx, plan)
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
		ctx := context.Background()
		hub, err := newWorkerManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		ID := args[0]
		_, err = hub.RemoveAskPlan(ctx, &pb.ID{Id: ID})
		if err != nil {
			showError(cmd, "Cannot remove AskOrder", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

func loadSlotFile(path string) (*structs.Slot, error) {
	cfg := task_config.SlotConfig{}
	err := util.LoadYamlFile(path, &cfg)
	if err != nil {
		return nil, err
	}

	slot, err := cfg.IntoSlot()
	if err != nil {
		return nil, err
	}

	return slot, nil
}
