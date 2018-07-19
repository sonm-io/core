package commands

import (
	"fmt"
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
	forceDealFlag     bool
)

func init() {
	dealListCmd.PersistentFlags().Uint64Var(&dealsSearchCount, "limit", 10, "Deals count to show")
	dealCloseCmd.PersistentFlags().StringVar(&blacklistTypeStr, "blacklist", "none", "Whom to add to blacklist: `worker`, `master` or `none`")

	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewDurationFlag, "new-duration", "", "Propose new duration for a deal")
	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewPriceFlag, "new-price", "", "Propose new price for a deal")

	dealOpenCmd.PersistentFlags().BoolVar(&forceDealFlag, "force", false, "Force deal opening without checking worker availability")
	dealQuickBuyCmd.PersistentFlags().BoolVar(&forceDealFlag, "force", false, "Force deal opening without checking worker availability")

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
	Use:               "deal",
	Short:             "Manage deals",
	PersistentPreRunE: loadKeyStoreIfRequired,
}

var dealListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show your active deals",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		req := &pb.Count{Count: dealsSearchCount}
		deals, err := dealer.List(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot get deals list: %v", err)
		}

		printDealsList(cmd, deals.GetDeal())
		return nil
	},
}

var dealStatusCmd = &cobra.Command{
	Use:   "status <deal_id>",
	Short: "Show deal status",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			return err
		}

		reply, err := dealer.Status(ctx, id)
		if err != nil {
			return fmt.Errorf("cannot get deal info: %v", err)
		}

		changeRequests, _ := dealer.ChangeRequestsList(ctx, id)
		printDealInfo(cmd, reply, changeRequests, printEverything)
		return nil
	},
}

var dealOpenCmd = &cobra.Command{
	Use:   "open <ask_id> <bid_id>",
	Short: "Open deal with given orders",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		askID, err := util.ParseBigInt(args[0])
		if err != nil {
			return err
		}

		bidID, err := util.ParseBigInt(args[1])
		if err != nil {
			return err
		}

		deals, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		deal, err := deals.Open(ctx, &pb.OpenDealRequest{
			BidID: pb.NewBigInt(bidID),
			AskID: pb.NewBigInt(askID),
			Force: forceDealFlag,
		})

		if err != nil {
			return fmt.Errorf("cannot open deal: %v", err)
		}

		printID(cmd, deal.GetId().Unwrap().String())
		return nil
	},
}

var dealQuickBuyCmd = &cobra.Command{
	Use:   "quick-buy <ask_id> [duration]",
	Short: "Instantly open deal with provided ask order id and optional duration (should be less or equal comparing to ask order)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		deals, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := pb.NewBigIntFromString(args[0])
		if err != nil {
			return fmt.Errorf("cannot convert arg to number: %v", err)
		}

		req := &pb.QuickBuyRequest{
			AskID: id,
			Force: forceDealFlag,
		}

		if len(args) >= 2 {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("cannot parse specified duration: %v", err)
			}
			req.Duration = &pb.Duration{
				Nanoseconds: duration.Nanoseconds(),
			}
		}

		deal, err := deals.QuickBuy(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot perform quick buy on given order: %v", err)
		}

		printDealInfo(cmd, deal, nil, printEverything)
		return nil
	},
}

var dealCloseCmd = &cobra.Command{
	Use:   "close <deal_id>",
	Short: "Close given deal",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("cannot parse `blacklist` argumet, allowed values are `none`, `worker` and `master`")
		}

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			return err
		}

		if _, err = dealer.Finish(ctx, &pb.DealFinishRequest{
			Id:            pb.NewBigInt(id),
			BlacklistType: blacklistType,
		}); err != nil {
			return fmt.Errorf("cannot finish deal: %v", err)
		}

		showOk(cmd)
		return nil
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			return err
		}

		durationRaw := cmd.Flag("new-duration").Value.String()
		priceRaw := cmd.Flag("new-price").Value.String()

		// check that at least one flag is present
		if len(durationRaw) == 0 && len(priceRaw) == 0 {
			return fmt.Errorf("please specify at least one flag: --new-duration or --new-price")
		}

		var newPrice = &pb.Price{}
		var newDuration time.Duration

		if len(durationRaw) > 0 {
			newDuration, err = time.ParseDuration(durationRaw)
			if err != nil {
				return fmt.Errorf("cannot convert flag value to duration: %v", err)
			}
		}

		if len(priceRaw) > 0 {
			if err := newPrice.LoadFromString(priceRaw); err != nil {
				return fmt.Errorf("cannot convert flag value to price: %v", err)
			}
		}

		req := &pb.DealChangeRequest{
			DealID:   pb.NewBigInt(id),
			Duration: uint64(newDuration.Seconds()),
			Price:    newPrice.GetPerSecond(),
		}

		crid, err := dealer.CreateChangeRequest(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot create change request: %v", err)
		}

		cmd.Printf("Change request ID = %v\n", crid.Unwrap().String())
		return nil
	},
}

var changeRequestApproveCmd = &cobra.Command{
	Use:   "approve <req_id>",
	Short: "Agree to change deal conditions with given change request",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			return err
		}

		if _, err := dealer.ApproveChangeRequest(ctx, pb.NewBigInt(id)); err != nil {
			return fmt.Errorf("cannot approve change request: %v", err)
		}

		showOk(cmd)
		return nil
	},
}

var changeRequestCancelCmd = &cobra.Command{
	Use:   "cancel <req_id>",
	Short: "Decline given change request",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		id, err := util.ParseBigInt(args[0])
		if err != nil {
			return fmt.Errorf("cannot convert arg to id: %v", err)
		}

		if _, err := dealer.CancelChangeRequest(ctx, pb.NewBigInt(id)); err != nil {
			return fmt.Errorf("cannot cancel change request: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
