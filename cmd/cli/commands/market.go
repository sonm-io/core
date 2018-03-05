package commands

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
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
		marketProcessingCmd,
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
		ctx := context.Background()
		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		ordType, err := structs.ParseOrderType(orderSearchType)
		slotPath := args[0]
		if err != nil || ordType == pb.OrderType_ANY {
			showError(cmd, "Cannot parse order type", err)
			os.Exit(1)
		}

		slot, err := loadSlotFile(slotPath)
		if err != nil {
			showError(cmd, "Cannot parse slot file", err)
			os.Exit(1)
		}

		req := &pb.GetOrdersRequest{
			Order: &pb.Order{
				OrderType: ordType,
				Slot:      slot.Unwrap(),
			},
			Count: ordersSearchLimit,
		}

		reply, err := market.GetOrders(ctx, req)
		if err != nil {
			showError(cmd, "Cannot get orders", err)
			os.Exit(1)
		}

		printSearchResults(cmd, reply.GetOrders())
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
			showError(cmd, "Cannot get order by Name", err)
			os.Exit(1)
		}

		printOrderDetails(cmd, order)
	},
}

var marketProcessingCmd = &cobra.Command{
	Use:   "processing",
	Short: "Show processing orders",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		reply, err := market.GetProcessing(ctx, &pb.Empty{})
		if err != nil {
			showError(cmd, "Cannot get processing orders", err)
			os.Exit(1)
		}

		printProcessingOrders(cmd, reply)
	},
}

var marketCreteCmd = &cobra.Command{
	Use:   "create <price> <slot.yaml> [supplier-eth-addr]",
	Short: "Place new Bid order on Marketplace",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		market, err := newMarketClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		price := args[0]
		orderPath := args[1]

		bigPrice, err := util.StringToEtherPrice(price)
		if err != nil {
			showError(cmd, "Cannot parse price", err)
			os.Exit(1)
		}

		slot, err := loadSlotFile(orderPath)
		if err != nil {
			showError(cmd, "Cannot load order", err)
			os.Exit(1)
		}

		order := &pb.Order{
			PricePerSecond: pb.NewBigInt(bigPrice),
			Slot:           slot.Unwrap(),
			OrderType:      pb.OrderType_BID,
		}

		if len(args) > 2 {
			order.SupplierID = common.HexToAddress(args[2]).Hex()
		}

		created, err := market.CreateOrder(ctx, order)
		if err != nil {
			showError(cmd, "Cannot create order at Marketplace", err)
			os.Exit(1)
		}

		printID(cmd, created.Id)
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
		_, err = market.CancelOrder(ctx, &pb.Order{Id: orderID})
		if err != nil {
			showError(cmd, "Cannot cancel order on Marketplace", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
