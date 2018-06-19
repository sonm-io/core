package commands

import (
	"os"

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
	Use:   "order",
	Short: "Manage orders",
}

var orderListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show your active orders",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		req := &pb.Count{Count: ordersSearchLimit}
		reply, err := market.GetOrders(ctx, req)
		if err != nil {
			showError(cmd, "Cannot receive orders from marketplace", err)
			os.Exit(1)
		}

		printOrdersList(cmd, reply.Orders)
	},
}

var orderStatusCmd = &cobra.Command{
	Use:    "status <order_id>",
	Short:  "Show order stats",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

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

var orderCreateCmd = &cobra.Command{
	Use:    "create <bid.yaml>",
	Short:  "Place new Bid order on Marketplace",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		path := args[0]
		bid := &pb.BidOrder{}

		if err := task_config.LoadFromFile(path, bid); err != nil {
			showError(cmd, "Cannot load order definition", err)
			os.Exit(1)
		}

		created, err := market.CreateOrder(ctx, bid)
		if err != nil {
			showError(cmd, "Cannot create order on marketplace", err)
			os.Exit(1)
		}

		printID(cmd, created.GetId().Unwrap().String())
	},
}

var orderCancelCmd = &cobra.Command{
	Use:    "cancel <order_id>",
	Short:  "Cancel order on Marketplace",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

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

var orderPurgeCmd = &cobra.Command{
	Use:    "purge",
	Short:  "Remove all your orders from Marketplace",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		if _, err := market.Purge(ctx, &pb.Empty{}); err != nil {
			showError(cmd, "Cannot purge orders", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
