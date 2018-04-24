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
	marketSearchCmd.PersistentFlags().Uint64Var(&ordersSearchLimit, "limit", 10, "Orders count to show")

	marketRootCmd.AddCommand(
		marketSearchCmd,
		marketShowCmd,
		marketCreateCmd,
		marketCancelCmd,
	)
}

var marketRootCmd = &cobra.Command{
	Use:   "market",
	Short: "Interact with Marketplace",
}

var marketSearchCmd = &cobra.Command{
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

		printSearchResults(cmd, reply.Orders)
	},
}

var marketShowCmd = &cobra.Command{
	Use:    "show <order_id>",
	Short:  "Show order details",
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

// Note: here is no processing method at all, we need to move matching code
// into the separated package, and then reinvent processing from scratch.

var marketCreateCmd = &cobra.Command{
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

		printID(cmd, created.GetId())
	},
}

var marketCancelCmd = &cobra.Command{
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
