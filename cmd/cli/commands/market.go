package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/structs"
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
		marketProcessingCmd,
	)
}

var marketRootCmd = &cobra.Command{
	Use:   "market",
	Short: "Interact with Marketplace",
}

var marketSearchCmd = &cobra.Command{
	Use:    "search <slot.yaml>",
	Short:  "Search for orders on Marketplace",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
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

		orders, err := market.GetOrders(slot, ordType, ordersSearchLimit)
		if err != nil {
			showError(cmd, "Cannot get orders", err)
			os.Exit(1)
		}

		printSearchResults(cmd, orders)
	},
}

var marketShowCmd = &cobra.Command{
	Use:    "show <order_id>",
	Short:  "Show order details",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddressFlag, timeoutFlag)
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

var marketProcessingCmd = &cobra.Command{
	Use:    "processing",
	Short:  "Show processing orders",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		reply, err := market.GetProcessing()
		if err != nil {
			showError(cmd, "Cannot get processing orders", err)
			os.Exit(1)
		}
		printProcessingOrders(cmd, reply)
	},
}

var marketCreteCmd = &cobra.Command{
	Use:    "create <order.yaml> [supplier-eth-addr]",
	Short:  "Place new Bid order on Marketplace",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddressFlag, timeoutFlag)
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

		inner := order.Unwrap()
		if len(args) > 1 {
			addr := common.HexToAddress(args[1])
			inner.SupplierID = addr.Hex()
		}

		created, err := market.CreateOrder(inner)
		if err != nil {
			showError(cmd, "Cannot create order at Marketplace", err)
			os.Exit(1)
		}

		printID(cmd, created.Id)
	},
}

var marketCancelCmd = &cobra.Command{
	Use:    "cancel <order_id>",
	Short:  "Cancel order on Marketplace",
	PreRun: loadKeyStoreWrapper,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		market, err := NewMarketInteractor(nodeAddressFlag, timeoutFlag)
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
