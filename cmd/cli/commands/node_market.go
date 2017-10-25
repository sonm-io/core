package commands

import (
	"os"

	ds "github.com/c2h5oh/datasize"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
)

var (
	ordersSearchLimit uint64 = 0
	orderSearchType          = "ANY"
)

func init() {
	nodeMarketSearchCmd.PersistentFlags().StringVar(&orderSearchType, "type", "ANY",
		"Orders type to search: ANY, BID or ASK")
	nodeMarketSearchCmd.PersistentFlags().Uint64Var(&ordersSearchLimit, "limit", 10,
		"Orders count to show")

	nodeMarketRootCmd.AddCommand(
		nodeMarketSearchCmd,
		nodeMarketShowCmd,
		nodeMarketCreteCmd,
		nodeMarketCancelCmd,
	)
}

var nodeMarketRootCmd = &cobra.Command{
	Use:   "market",
	Short: "Interact with Marketplace",
}

func printSearchResults(cmd *cobra.Command, orders []*pb.Order) {
	if len(orders) == 0 {
		cmd.Printf("No matching orders found")
		return
	}

	for i, order := range orders {
		cmd.Printf("%d) %s %s | price = %d\r\n", i+1, order.OrderType.String(), order.Id, order.Price)
	}
}

func printOrderDetails(cmd *cobra.Command, order *pb.Order) {
	cmd.Printf("ID:             %s\r\n", order.Id)
	cmd.Printf("Price:          %d\r\n", order.Price)

	cmd.Printf("SupplierID:     %s\r\n", order.SupplierID)
	cmd.Printf("SupplierRating: %d\r\n", order.Slot.SupplierRating)
	cmd.Printf("BuyerID:        %s\r\n", order.ByuerID)
	cmd.Printf("BuyerRating:    %d\r\n", order.Slot.BuyerRating)

	rs := order.Slot.Resources
	cmd.Printf("Resources:\r\n")
	cmd.Printf("  CPU:     %d\r\n", rs.CpuCores)
	cmd.Printf("  GPU:     %d\r\n", rs.GpuCount)
	cmd.Printf("  RAM:     %s\r\n", ds.ByteSize(rs.RamBytes).HR())
	cmd.Printf("  Storage: %s\r\n", ds.ByteSize(rs.Storage).HR())
	cmd.Printf("  Network: %s\r\n", rs.NetworkType.String())
	cmd.Printf("    In:   %s\r\n", ds.ByteSize(rs.NetTrafficIn).HR())
	cmd.Printf("    Out:  %s\r\n", ds.ByteSize(rs.NetTrafficOut).HR())
}

func printOrderCreated(cmd *cobra.Command, order *pb.Order) {
	cmd.Println("Order created!")
	cmd.Printf("ID = %s\r\n", order.Id)
}

var nodeMarketSearchCmd = &cobra.Command{
	Use:   "search <slot.yaml>",
	Short: "Place new Bid order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		ordType, err := structs.ParseOrderType(orderSearchType)
		slotPath := args[0]
		if err != nil {
			showError(cmd, "Cannot parse order type", err)
			os.Exit(1)
		}

		slot, err := loadSlotFile(slotPath)
		if err != nil {
			showError(cmd, "Cannot parse slot file", err)
			os.Exit(1)
		}

		orders, err := market.GetOrders(slot, ordType, ordersSearchLimit)
		if err != nil {
			showError(cmd, "Cannot get orders", err)
			os.Exit(1)
		}

		printSearchResults(cmd, orders)

	},
}

var nodeMarketShowCmd = &cobra.Command{
	Use:   "show <order_id>",
	Short: "Show order details",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		orderID := args[0]
		order, err := market.GetOrderByID(orderID)
		if err != nil {
			showError(cmd, "Cannot get order by ID", err)
			os.Exit(1)
		}

		printOrderDetails(cmd, order)
	},
}

var nodeMarketCreteCmd = &cobra.Command{
	Use:   "create <order.yaml>",
	Short: "Place new Bid order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		orderPath := args[0]
		order, err := loadOrderFile(orderPath)
		if err != nil {
			showError(cmd, "Cannot load order", err)
			os.Exit(1)
		}

		created, err := market.CreateOrder(order.Unwrap())
		if err != nil {
			showError(cmd, "Cannot create order at Marketplace", err)
			os.Exit(1)
		}

		printOrderCreated(cmd, created)
	},
}

var nodeMarketCancelCmd = &cobra.Command{
	Use:   "cancel <order_id>",
	Short: "Cancel order on Marketplace",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		orderID := args[0]

		err = market.CancelOrder(orderID)
		if err != nil {
			showError(cmd, "Cannot cancel order on Marketplace", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
