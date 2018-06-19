package commands

import (
	"os"
	"time"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	dealsSearchCount  uint64
	blacklistTypeStr  string
	crNewDurationFlag string
	crNewPriceFlag    string
)

func init() {
	dealListCmd.PersistentFlags().Uint64Var(&dealsSearchCount, "limit", 10, "Deals count to show")
	dealCloseCmd.PersistentFlags().StringVar(&blacklistTypeStr, "blacklist", "none", "Whom to add to blacklist (worker, master or neither)")
	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewDurationFlag, "new-duration", "", "Propose new duration for a deal")
	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewPriceFlag, "new-price", "", "Propose new price for a deal")

	changeRequestsRoot.AddCommand(
		changeRequestCreateCmd,
		changeRequestApproveCmd,
		changeRequestCancelCmd,
	)

	dealRootCmd.AddCommand(
		dealListCmd,
		dealStatusCmd,
		dealOpenCmd,
		dealQuickBuyCmd,
		dealCloseCmd,
		changeRequestsRoot,
	)
}

var dealRootCmd = &cobra.Command{
	Use:   "deal",
	Short: "Manage deals",
}

var dealListCmd = &cobra.Command{
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

var dealStatusCmd = &cobra.Command{
	Use:    "status <deal_id>",
	Short:  "Show deal status",
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

		id, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		reply, err := dealer.Status(ctx, id)
		if err != nil {
			showError(cmd, "Cannot get deal info", err)
			os.Exit(1)
		}

		changeRequests, _ := dealer.ChangeRequestsList(ctx, id)
		printDealInfo(cmd, reply, changeRequests)
	},
}

var dealOpenCmd = &cobra.Command{
	Use:    "open <ask_id> <bid_id>",
	Short:  "Open deal with given orders",
	Args:   cobra.MinimumNArgs(2),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
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

		deals, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create blockchain connection", err)
			os.Exit(1)
		}

		deal, err := deals.Open(ctx, &pb.OpenDealRequest{
			BidID: pb.NewBigInt(bidID),
			AskID: pb.NewBigInt(askID),
		})

		if err != nil {
			showError(cmd, "Cannot open deal", err)
			os.Exit(1)
		}

		printID(cmd, deal.GetId().Unwrap().String())
	},
}

var dealQuickBuyCmd = &cobra.Command{
	Use:    "quick-buy <ask_id> [duration]",
	Short:  "Instantly open deal with provided ask order id and optional duration (should be less or equal comparing to ask order)",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		deals, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		id, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		req := &pb.QuickBuyRequest{
			AskId: id,
		}
		if len(args) >= 2 {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				showError(cmd, "Cannot parse specified duration", err)
				os.Exit(1)
			}
			req.Duration = &pb.Duration{
				Nanoseconds: duration.Nanoseconds(),
			}
		}
		deal, err := deals.QuickBuy(ctx, req)
		if err != nil {
			showError(cmd, "Cannot perform quick buy on given order", err)
			os.Exit(1)
		}

		printDealInfo(cmd, &pb.DealInfoReply{Deal: deal}, nil)
	},
}

var dealCloseCmd = &cobra.Command{
	Use:    "close <deal_id>",
	Short:  "Close given deal",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()
		var blacklistType pb.BlacklistType
		switch blacklistTypeStr {
		case "none":
			blacklistType = pb.BlacklistType_BLACKLIST_NOBODY
		case "worker":
			blacklistType = pb.BlacklistType_BLACKLIST_WORKER
		case "master":
			blacklistType = pb.BlacklistType_BLACKLIST_MASTER
		default:
			showError(cmd, "Cannot parse `blacklist` argumet, allowed values are `none`, `worker` and `master`", nil)
			os.Exit(1)
		}

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to number", err)
			os.Exit(1)
		}

		_, err = dealer.Finish(ctx, &pb.DealFinishRequest{
			Id:            pb.NewBigInt(id),
			BlacklistType: blacklistType,
		})
		if err != nil {
			showError(cmd, "Cannot finish deal", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var changeRequestsRoot = &cobra.Command{
	Use:   "change-request",
	Short: "Request changes for deals",
}

var changeRequestCreateCmd = &cobra.Command{
	Use: "create <deal_id>",
	// space is added to align `usage` and `example` output into cobra's help message
	Example: "  sonmcli deal change-request create 123 --new-duration=10h --new-price=0.3USD/h",
	Short:   "Request changes for given deal",
	Args:    cobra.MinimumNArgs(1),
	PreRun:  loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			showError(cmd, "Cannot create client connection", err)
			os.Exit(1)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to id", err)
			os.Exit(1)
		}

		durationRaw := cmd.Flag("new-duration").Value.String()
		priceRaw := cmd.Flag("new-price").Value.String()

		// check that at least one flag is present
		if len(durationRaw) == 0 && len(priceRaw) == 0 {
			showError(cmd, "Please specify at least one flag: --new-duration or --new-price", nil)
			os.Exit(1)
		}

		var newPrice = &pb.Price{}
		var newDuration time.Duration

		if len(durationRaw) > 0 {
			newDuration, err = time.ParseDuration(durationRaw)
			if err != nil {
				showError(cmd, "Cannot convert flag value to duration", err)
				os.Exit(1)
			}
		}

		if len(priceRaw) > 0 {
			if err := newPrice.LoadFromString(priceRaw); err != nil {
				showError(cmd, "Cannot convert flag value to price", err)
				os.Exit(1)
			}
		}

		req := &pb.DealChangeRequest{
			DealID:   pb.NewBigInt(id),
			Duration: uint64(newDuration.Seconds()),
			Price:    newPrice.GetPerSecond(),
		}

		crid, err := dealer.CreateChangeRequest(ctx, req)
		if err != nil {
			showError(cmd, "Cannot create change request", err)
			os.Exit(1)
		}

		cmd.Printf("Change request ID = %v\n", crid.Unwrap().String())
	},
}

var changeRequestApproveCmd = &cobra.Command{
	Use:    "approve <req_id>",
	Short:  "Agree to change deal conditions with given change request",
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

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to id", err)
			os.Exit(1)
		}

		if _, err := dealer.ApproveChangeRequest(ctx, pb.NewBigInt(id)); err != nil {
			showError(cmd, "Cannot approve change request", err)
			os.Exit(1)
		}
		showOk(cmd)
	},
}

var changeRequestCancelCmd = &cobra.Command{
	Use:    "cancel <req_id>",
	Short:  "Decline given change request",
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

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			showError(cmd, "Cannot convert arg to id", err)
			os.Exit(1)
		}

		if _, err := dealer.CancelChangeRequest(ctx, pb.NewBigInt(id)); err != nil {
			showError(cmd, "Cannot cancel change request", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}
