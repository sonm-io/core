package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"

	"github.com/sonm-io/core/cmd/cli/task_config"
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
		hub, err := newHubManagementClient(ctx)
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
	Use:   "create <ask-plan.yaml>",
	Short: "Create new plan",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		planPath := args[0]
		ask, err := loadAskPlanYAML(planPath)
		if err != nil {
			showError(cmd, "Cannot load Ask-plan definition", err)
			os.Exit(1)
		}

		id, err := hub.CreateAskPlan(ctx, ask)
		if err != nil {
			showError(cmd, "Cannot create new Ask-plan", err)
			os.Exit(1)
		}

		printID(cmd, id.GetId())
	},
}

var askPlanRemoveCmd = &cobra.Command{
	Use:   "remove <ask_id>",
	Short: "Remove plan by",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		ID := args[0]
		_, err = hub.RemoveAskPlan(ctx, &pb.ID{Id: ID})
		if err != nil {
			showError(cmd, "Cannot remove Ask-plan", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

func loadAskPlanYAML(path string) (*pb.AskPlan, error) {
	planYAML, err := task_config.LoadAskPlan(path)
	if err != nil {
		return nil, err
	}

	return planYAML.Unwrap()
}
