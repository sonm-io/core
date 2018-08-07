package commands

import (
	"fmt"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var (
	ordersSearchLimit uint64 = 0
)

func init() {
	orderListCmd.PersistentFlags().Uint64Var(&ordersSearchLimit, "limit", 10, "Orders count to show")

	orderRootCmd.AddCommand(
		orderListCmd,
		orderStatusCmd,
		orderCreateCmd,
		orderCancelCmd,
		orderPurgeCmd,
	)
}

var orderRootCmd = &cobra.Command{
	Use:               "order",
	Short:             "Manage orders",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var orderListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show your active orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			return fmt.Errorf("сannot create client connection: %v", err)
		}

		req := &pb.Count{Count: ordersSearchLimit}
		reply, err := market.GetOrders(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot receive orders from marketplace: %v", err)
		}

		printOrdersList(cmd, reply.Orders)
		return nil
	},
}

var orderStatusCmd = &cobra.Command{
	Use:   "status <order_id>",
	Short: "Show order stats",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		orderID := args[0]
		order, err := market.GetOrderByID(ctx, &pb.ID{Id: orderID})
		if err != nil {
			return fmt.Errorf("cannot get order by ID: %v", err)
		}

		printOrderDetails(cmd, order)
		return nil
	},
}

var orderCreateCmd = &cobra.Command{
	Use:   "create <bid.yaml>",
	Short: "Place new Bid order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		path := args[0]
		bid := &pb.BidOrder{}
		if err := task_config.LoadFromFile(path, bid); err != nil {
			return fmt.Errorf("cannot load order definition: %v", err)
		}

		created, err := market.CreateOrder(ctx, bid)
		if err != nil {
			return fmt.Errorf("cannot create order on marketplace: %v", err)
		}

		printID(cmd, created.GetId().Unwrap().String())
		return nil
	},
}

var orderCancelCmd = &cobra.Command{
	Use:   "cancel <order_id>",
	Short: "Cancel order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		orderID := args[0]
		_, err = market.CancelOrder(ctx, &pb.ID{Id: orderID})
		if err != nil {
			return fmt.Errorf("cannot cancel order on Marketplace: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var orderPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Remove all your orders from Marketplace",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		if _, err := market.Purge(ctx, &pb.Empty{}); err != nil {
			return fmt.Errorf("cannot purge orders: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
