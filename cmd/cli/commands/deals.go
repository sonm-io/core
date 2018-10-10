package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	dealsSearchCount  uint64
	blacklistTypeStr  string
	crNewDurationFlag string
	crNewPriceFlag    string
	forceDealFlag     bool
	expandDealFlag    bool
)

func init() {
	dealListCmd.PersistentFlags().Uint64Var(&dealsSearchCount, "limit", 10, "Deals count to show")
	dealCloseCmd.PersistentFlags().StringVar(&blacklistTypeStr, "blacklist", "nobody", "Whom to add to blacklist: `worker`, `master` or `nobody`")

	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewDurationFlag, "new-duration", "", "Propose new duration for a deal")
	changeRequestCreateCmd.PersistentFlags().StringVar(&crNewPriceFlag, "new-price", "", "Propose new price for a deal")

	dealOpenCmd.PersistentFlags().BoolVar(&forceDealFlag, "force", false, "Force deal opening without checking worker availability")
	dealQuickBuyCmd.PersistentFlags().BoolVar(&forceDealFlag, "force", false, "Force deal opening without checking worker availability")

	dealStatusCmd.PersistentFlags().BoolVar(&expandDealFlag, "expand", false, "Print extended orders' info bound to deal")
	dealQuickBuyCmd.PersistentFlags().BoolVar(&expandDealFlag, "expand", false, "Print extended orders' info bound to deal")

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
		dealPurgeCmd,
		changeRequestsRoot,
	)
}

var dealRootCmd = &cobra.Command{
	Use:               "deal",
	Short:             "Manage deals",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var dealListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show your active deals",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dwh, err := newDWHClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		addr, err := keystore.GetDefaultAddress()
		if err != nil {
			return fmt.Errorf("cannot get default address: %v", err)
		}
		req := &sonm.DealsRequest{
			AnyUserID: sonm.NewEthAddress(addr),
			Limit:     dealsSearchCount,
			Status:    sonm.DealStatus_DEAL_ACCEPTED,
			Sortings: []*sonm.SortingOption{{
				Field: "StartTime",
				Order: sonm.SortingOrder_Asc,
			}},
		}

		deals, err := dwh.GetDeals(ctx, req)
		printDealsList(cmd, deals.GetDeals())
		return nil
	},
}

func appendExtendedInfo(ctx context.Context, dealInfo *ExtendedDealInfo) error {
	market, err := newMarketClient(ctx)
	if err != nil {
		return fmt.Errorf("cannot create client connection: %v", err)
	}
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		ask, err := market.GetOrderByID(ctx, &sonm.ID{Id: dealInfo.GetDeal().GetAskID().Unwrap().String()})
		if err != nil {
			return fmt.Errorf("failed to fetch ask order: %v", err)
		}
		dealInfo.Ask = ask
		return nil
	})
	wg.Go(func() error {
		bid, err := market.GetOrderByID(ctx, &sonm.ID{Id: dealInfo.GetDeal().GetBidID().Unwrap().String()})
		if err != nil {
			return fmt.Errorf("failed to fetch bid order: %v", err)
		}
		dealInfo.Bid = bid
		return nil
	})
	return wg.Wait()
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

		id, err := sonm.NewBigIntFromString(args[0])
		if err != nil {
			return err
		}

		reply, err := dealer.Status(ctx, id)
		if err != nil {
			return fmt.Errorf("cannot get deal info: %v", err)
		}
		dealInfo := &ExtendedDealInfo{
			DealInfoReply: reply,
		}
		if expandDealFlag {
			if err := appendExtendedInfo(ctx, dealInfo); err != nil {
				return err
			}
		}

		changeRequests, err := dealer.ChangeRequestsList(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch change request list: %v", err)
		}
		dealInfo.ChangeRequests = changeRequests
		printDealInfo(cmd, dealInfo, printEverything)
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

		deal, err := deals.Open(ctx, &sonm.OpenDealRequest{
			BidID: sonm.NewBigInt(bidID),
			AskID: sonm.NewBigInt(askID),
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

		id, err := sonm.NewBigIntFromString(args[0])
		if err != nil {
			return fmt.Errorf("cannot convert arg to number: %v", err)
		}

		req := &sonm.QuickBuyRequest{
			AskID: id,
			Force: forceDealFlag,
		}

		if len(args) >= 2 {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return fmt.Errorf("cannot parse specified duration: %v", err)
			}
			req.Duration = &sonm.Duration{
				Nanoseconds: duration.Nanoseconds(),
			}
		}

		deal, err := deals.QuickBuy(ctx, req)
		if err != nil {
			return fmt.Errorf("cannot perform quick buy on given order: %v", err)
		}

		info := &ExtendedDealInfo{
			DealInfoReply: deal,
		}
		if expandDealFlag {
			if err := appendExtendedInfo(ctx, info); err != nil {
				return err
			}
		}
		printDealInfo(cmd, info, printEverything)
		return nil
	},
}

func getBlacklistType() (sonm.BlacklistType, error) {
	blacklistTypeStr = "BLACKLIST_" + strings.ToUpper(blacklistTypeStr)
	blacklistType, ok := sonm.BlacklistType_value[blacklistTypeStr]
	if !ok {
		return sonm.BlacklistType_BLACKLIST_NOBODY, fmt.Errorf("cannot parse `blacklist` argumet, allowed values are `nobody`, `worker` and `master`")
	}
	return sonm.BlacklistType(blacklistType), nil
}

var dealCloseCmd = &cobra.Command{
	Use:   "close <deal_id>...",
	Short: "Close given deals",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		blacklistType, err := getBlacklistType()
		if err != nil {
			return err
		}

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		request := &sonm.DealsFinishRequest{
			DealInfo: make([]*sonm.DealFinishRequest, 0, len(args)),
		}
		ids, err := argsToBigInts(args)
		if err != nil {
			return fmt.Errorf("failed to parse parameters to deal ids: %v", err)
		}
		for _, id := range ids {
			request.DealInfo = append(request.DealInfo, &sonm.DealFinishRequest{
				Id:            sonm.NewBigInt(id),
				BlacklistType: blacklistType,
			})
		}

		status, err := dealer.FinishDeals(ctx, request)
		if err != nil {
			return fmt.Errorf("cannot finish deal: %v", err)
		}
		printErrorById(cmd, status)
		return nil
	},
}

var dealPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge all active consumer's deals",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		blacklistType, err := getBlacklistType()
		if err != nil {
			return err
		}

		dealer, err := newDealsClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		status, err := dealer.PurgeDeals(ctx, &sonm.DealsPurgeRequest{BlacklistType: blacklistType})
		if err != nil {
			return fmt.Errorf("cannot purge deals: %v", err)
		}

		printErrorById(cmd, status)
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

		var newPrice = &sonm.Price{}
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

		req := &sonm.DealChangeRequest{
			DealID:   sonm.NewBigInt(id),
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

		if _, err := dealer.ApproveChangeRequest(ctx, sonm.NewBigInt(id)); err != nil {
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

		if _, err := dealer.CancelChangeRequest(ctx, sonm.NewBigInt(id)); err != nil {
			return fmt.Errorf("cannot cancel change request: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
