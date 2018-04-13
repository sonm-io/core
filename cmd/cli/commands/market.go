package commands

import (
	"context"
	"os"

	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var (
	ordersSearchLimit uint64 = 0
	orderSearchType          = "ANY"
)

func init() {
	marketSearchCmd.PersistentFlags().StringVar(&orderSearchType, "type", "BID",
		"Orders type to search: BID or ASK")
	marketSearchCmd.PersistentFlags().Uint64Var(&ordersSearchLimit, "limit", 10,
		"Orders count to show")

	marketRootCmd.AddCommand(
		marketSearchCmd,
		marketShowCmd,
		marketCreteCmd,
		marketCancelCmd,
	)
}

var marketRootCmd = &cobra.Command{
	Use:   "market",
	Short: "Interact with Marketplace",
}

var marketSearchCmd = &cobra.Command{
	Use:   "search <slot.yaml>",
	Short: "Search for orders on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// todo: need to implement with new market API.
		showError(cmd, "not implemented", nil)
		os.Exit(1)
	},
}

var marketShowCmd = &cobra.Command{
	Use:   "show <order_id>",
	Short: "Show order details",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		orderID := args[0]
		order, err := market.GetOrderByID(ctx, &pb.ID{Id: orderID})
		if err != nil {
			showError(cmd, "Cannot get order by ID", err)
			os.Exit(1)
		}

		printOrderDetails(cmd, order)
	},
}

// Note: here is no processing method at all, we need to move matching code
// into the separated package, and then reinvent processing from scratch.

var marketCreteCmd = &cobra.Command{
	Use:   "create <price> <slot.yaml> [supplier-eth-addr]",
	Short: "Place new Bid order on Marketplace",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// todo: need to implement with new market API.
		// todo: create parser for bid.yaml
		showError(cmd, "not implemented", nil)
		os.Exit(1)
	},
}

var marketCancelCmd = &cobra.Command{
	Use:   "cancel <order_id>",
	Short: "Cancel order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		orderID := args[0]
		_, err = market.CancelOrder(ctx, &pb.ID{Id: orderID})
		if err != nil {
			showError(cmd, "Cannot cancel order on Marketplace", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
