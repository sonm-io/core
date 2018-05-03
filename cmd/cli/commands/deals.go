package commands

import (
	"os"

	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	dealsSearchCount uint64
)

func init() {
	dealsListCmd.PersistentFlags().Uint64Var(&dealsSearchCount, "limit", 10, "Deals count to show")

	nodeDealsRootCmd.AddCommand(
		dealsListCmd,
		dealsStatusCmd,
		dealsOpenCmd,
		dealsFinishCmd,
	)
}

var nodeDealsRootCmd = &cobra.Command{
	Use:   "deals",
	Short: "Manage deals",
}

var dealsListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show your active deals",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		req := &pb.Count{Count: dealsSearchCount}
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

		printDealInfo(cmd, reply.GetDeal())
	},
}

var dealsOpenCmd = &cobra.Command{
	Use:    "open <ask_id> <bid_id>",
	Short:  "open deal with given orders",
	Args:   cobra.MinimumNArgs(2),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(`WARN: this command exists only for debugging purposes and may be removed in future.`)
		ctx, cancel := newTimeoutContext()
		defer cancel()

		askID, err := util.ParseBigInt(args[0])
		if err != nil {
			// do not wraps error with human-readable text, the error text is self-explainable.
			showError(cmd, err.Error(), nil)
			os.Exit(1)
		}

		bidID, err := util.ParseBigInt(args[1])
		if err != nil {
			showError(cmd, err.Error(), nil)
			os.Exit(1)
		}

		eth, err := blockchain.NewAPI()
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		dealOrErr := <-eth.Market().OpenDeal(ctx, sessionKey, askID, bidID)
		if dealOrErr.Err != nil {
			showError(cmd, "Cannot open deal", dealOrErr.Err)
			os.Exit(1)
		}

		printID(cmd, dealOrErr.Deal.GetId().Unwrap().String())
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
