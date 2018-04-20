package commands

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	dealListFlagFrom   string
	dealListFlagStatus string
)

func init() {
	dealsListCmd.PersistentFlags().StringVar(&dealListFlagFrom, "from", "",
		"Transactions author, using self address if empty")
	dealsListCmd.PersistentFlags().StringVar(&dealListFlagStatus, "status", "ANY",
		"Transaction status (ANY, PENDING, ACCEPTED, CLOSED)")

	nodeDealsRootCmd.AddCommand(
		dealsListCmd,
		dealsStatusCmd,
		dealsFinishCmd,
	)
}

var nodeDealsRootCmd = &cobra.Command{
	Use:   "deals",
	Short: "Manage deals",
}

var dealsListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show my deals",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		status := convertTransactionStatus(dealListFlagStatus)
		from := dealListFlagFrom
		if from == "" {
			from = util.PubKeyToAddr(sessionKey.PublicKey).Hex()
		}

		req := &pb.DealListRequest{
			Owner:  &pb.EthAddress{Address: common.HexToAddress(from).Bytes()},
			Status: status,
		}
		deals, err := dealer.List(ctx, req)
		if err != nil {
			showError(cmd, "Cannot get deals list", err)
			os.Exit(1)
		}

		printDealsList(cmd, deals.GetDeal())
	},
}

var dealsStatusCmd = &cobra.Command{
	Use:    "status <deal_id>",
	Short:  "show deal status",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		id := args[0]
		_, err = util.ParseBigInt(id)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		reply, err := dealer.Status(ctx, &pb.ID{Id: id})
		if err != nil {
			showError(cmd, "Cannot get deal info", err)
			os.Exit(1)
		}

		printDealDetails(cmd, reply)
	},
}

var dealsFinishCmd = &cobra.Command{
	Use:    "finish <deal_id>",
	Short:  "finish deal",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		id := args[0]
		_, err = util.ParseBigInt(id)
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		_, err = dealer.Finish(ctx, &pb.ID{Id: id})
		if err != nil {
			showError(cmd, "Cannot finish deal", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

func convertTransactionStatus(s string) pb.DealStatus {
	s = strings.ToUpper(s)
	// add prefix for protobuf constants
	s = "DEAL_" + s

	id := pb.DealStatus_value[s]
	return pb.DealStatus(id)
}
