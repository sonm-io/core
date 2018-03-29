package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"

	"github.com/ethereum/go-ethereum/common"
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
	Use:   "create <price> <slot.yaml> [buyer-eth-addr]",
	Short: "Create new plan",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		hub, err := newHubManagementClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		price := args[0]
		planPath := args[1]

		bigPrice, err := util.StringToEtherPrice(price)
		if err != nil {
			showError(cmd, "Cannot parse price", err)
			os.Exit(1)
		}

		slot, err := loadSlotFile(planPath)
		if err != nil {
			showError(cmd, "Cannot load AskOrder definition", err)
			os.Exit(1)
		}

		req := &pb.InsertSlotRequest{
			Slot:           slot.Unwrap(),
			PricePerSecond: pb.NewBigInt(bigPrice),
		}

		if len(args) > 2 {
			addr := common.HexToAddress(args[2])
			req.BuyerID = addr.Hex()
		}

		id, err := hub.CreateAskPlan(ctx, req)
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
		hub, err := newHubManagementClient(ctx)
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
