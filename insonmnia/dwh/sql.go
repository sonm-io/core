package dwh

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MaxLimit         = 200
	NumMaxBenchmarks = 128
	gte              = ">="
	lte              = "<="
	eq               = "="
)

var (
	opsTranslator = map[pb.CmpOp]string{
		pb.CmpOp_GTE: gte,
		pb.CmpOp_LTE: lte,
		pb.CmpOp_EQ:  eq,
	}
)

type sqlStorage struct {
	commands      *sqlCommands
	setupCommands *sqlSetupCommands
	numBenchmarks uint64
	queryRunner   queryRunner
	tablesInfo    *tablesInfo
	formatCb      formatArg
}

func (c *sqlStorage) Setup(db *sql.DB) error {
	return c.setupCommands.setupTables(db)
}

func (c *sqlStorage) CreateIndices(db *sql.DB) error {
	return c.setupCommands.createIndices(db)
}

func (c *sqlStorage) InsertDeal(conn queryConn, deal *pb.Deal) error {
	ask, err := c.GetOrderByID(conn, deal.AskID.Unwrap())
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Ask)")
	}

	bid, err := c.GetOrderByID(conn, deal.BidID.Unwrap())
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Bid)")
	}

	var hasActiveChangeRequests bool
	if _, err := c.GetDealChangeRequestsByID(conn, deal.Id.Unwrap()); err == nil {
		hasActiveChangeRequests = true
	}
	allColumns := []interface{}{
		deal.Id.Unwrap().String(),
		deal.SupplierID.Unwrap().Hex(),
		deal.ConsumerID.Unwrap().Hex(),
		deal.MasterID.Unwrap().Hex(),
		deal.AskID.Unwrap().String(),
		deal.BidID.Unwrap().String(),
		deal.Duration,
		deal.Price.PaddedString(),
		deal.StartTime.Seconds,
		deal.EndTime.Seconds,
		uint64(deal.Status),
		deal.BlockedBalance.PaddedString(),
		deal.TotalPayout.PaddedString(),
		deal.LastBillTS.Seconds,
		ask.GetOrder().Netflags,
		ask.GetOrder().IdentityLevel,
		bid.GetOrder().IdentityLevel,
		ask.CreatorCertificates,
		bid.CreatorCertificates,
		hasActiveChangeRequests,
	}
	for benchID := uint64(0); benchID < c.numBenchmarks; benchID++ {
		allColumns = append(allColumns, deal.Benchmarks.Values[benchID])
	}
	_, err = conn.Exec(c.commands.insertDeal, allColumns...)

	return err
}

func (c *sqlStorage) UpdateDeal(conn queryConn, deal *pb.Deal) error {
	_, err := conn.Exec(c.commands.updateDeal,
		deal.Duration,
		deal.Price.PaddedString(),
		deal.StartTime.Seconds,
		deal.EndTime.Seconds,
		uint64(deal.Status),
		deal.BlockedBalance.PaddedString(),
		deal.TotalPayout.PaddedString(),
		deal.LastBillTS.Seconds,
		deal.Id.Unwrap().String())
	return err
}

func (c *sqlStorage) UpdateDealsSupplier(conn queryConn, profile *pb.Profile) error {
	_, err := conn.Exec(c.commands.updateDealsSupplier, []byte(profile.Certificates), profile.UserID.Unwrap().Hex())
	return err
}

func (c *sqlStorage) UpdateDealsConsumer(conn queryConn, profile *pb.Profile) error {
	_, err := conn.Exec(c.commands.updateDealsConsumer, []byte(profile.Certificates), profile.UserID.Unwrap().Hex())
	return err
}

func (c *sqlStorage) UpdateDealPayout(conn queryConn, dealID, payout *big.Int, billTS uint64) error {
	_, err := conn.Exec(c.commands.updateDealPayout, util.BigIntToPaddedString(payout), billTS, dealID.String())
	return err
}

func (c *sqlStorage) DeleteDeal(conn queryConn, dealID *big.Int) error {
	_, err := conn.Exec(c.commands.deleteDeal, dealID.String())
	return err
}

func (c *sqlStorage) GetDealByID(conn queryConn, dealID *big.Int) (*pb.DWHDeal, error) {
	rows, err := conn.Query(c.commands.selectDealByID, dealID.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to GetDealDetails")
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.New("no rows returned")
	}

	return c.decodeDeal(rows)
}

func (c *sqlStorage) GetDeals(conn queryConn, request *pb.DealsRequest) ([]*pb.DWHDeal, uint64, error) {
	var filters []*filter
	if request.Status > 0 {
		filters = append(filters, newFilter("Status", eq, request.Status, "AND"))
	}
	if !request.SupplierID.IsZero() {
		filters = append(filters, newFilter("SupplierID", eq, request.SupplierID.Unwrap().Hex(), "AND"))
	}
	if !request.ConsumerID.IsZero() {
		filters = append(filters, newFilter("ConsumerID", eq, request.ConsumerID.Unwrap().Hex(), "AND"))
	}
	if !request.MasterID.IsZero() {
		filters = append(filters, newFilter("MasterID", eq, request.MasterID.Unwrap().Hex(), "AND"))
	}
	if !request.AskID.IsZero() {
		filters = append(filters, newFilter("AskID", eq, request.AskID, "AND"))
	}
	if !request.BidID.IsZero() {
		filters = append(filters, newFilter("BidID", eq, request.BidID, "AND"))
	}
	if request.Duration != nil {
		if request.Duration.Max > 0 {
			filters = append(filters, newFilter("Duration", lte, request.Duration.Max, "AND"))
		}
		filters = append(filters, newFilter("Duration", gte, request.Duration.Min, "AND"))
	}
	if request.Price != nil {
		if request.Price.Max != nil {
			filters = append(filters, newFilter("Price", lte, request.Price.Max.PaddedString(), "AND"))
		}
		if request.Price.Min != nil {
			filters = append(filters, newFilter("Price", gte, request.Price.Min.PaddedString(), "AND"))
		}
	}
	if request.Netflags != nil && request.Netflags.Value > 0 {
		filters = append(filters, newNetflagsFilter(request.Netflags.Operator, request.Netflags.Value))
	}
	if request.AskIdentityLevel > 0 {
		filters = append(filters, newFilter("AskIdentityLevel", gte, request.AskIdentityLevel, "AND"))
	}
	if request.BidIdentityLevel > 0 {
		filters = append(filters, newFilter("BidIdentityLevel", gte, request.BidIdentityLevel, "AND"))
	}
	if request.Benchmarks != nil {
		c.addBenchmarksConditions(request.Benchmarks, &filters)
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Deals",
		filters:   filters,
		sortings:  c.filterSortings(request.Sortings, c.tablesInfo.DealColumnsSet),
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var deals []*pb.DWHDeal
	for rows.Next() {
		deal, err := c.decodeDeal(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeDeal")
		}

		deals = append(deals, deal)
	}

	return deals, count, nil
}

func (c *sqlStorage) GetDealConditions(conn queryConn, request *pb.DealConditionsRequest) ([]*pb.DealCondition, uint64, error) {
	var filters []*filter
	if len(request.Sortings) < 1 {
		request.Sortings = []*pb.SortingOption{{Field: "Id", Order: pb.SortingOrder_Desc}}
	}

	filters = append(filters, newFilter("DealID", eq, request.DealID.Unwrap().String(), "AND"))
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "DealConditions",
		filters:   filters,
		sortings:  request.Sortings,
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var out []*pb.DealCondition
	for rows.Next() {
		dealCondition, err := c.decodeDealCondition(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeDealCondition")
		}
		out = append(out, dealCondition)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, status.Error(codes.Internal, "failed to GetDealConditions")
	}

	return out, count, nil
}

func (c *sqlStorage) InsertOrder(conn queryConn, order *pb.DWHOrder) error {
	allColumns := []interface{}{
		order.GetOrder().Id.Unwrap().String(),
		order.CreatedTS.Seconds,
		order.GetOrder().DealID.Unwrap().String(),
		uint64(order.GetOrder().OrderType),
		uint64(order.GetOrder().OrderStatus),
		order.GetOrder().AuthorID.Unwrap().Hex(),
		order.GetOrder().CounterpartyID.Unwrap().Hex(),
		order.GetOrder().Duration,
		order.GetOrder().Price.PaddedString(),
		order.GetOrder().Netflags,
		uint64(order.GetOrder().IdentityLevel),
		order.GetOrder().Blacklist,
		order.GetOrder().Tag,
		order.GetOrder().FrozenSum.PaddedString(),
		order.CreatorIdentityLevel,
		order.CreatorName,
		order.CreatorCountry,
		[]byte(order.CreatorCertificates),
	}
	for benchID := uint64(0); benchID < c.numBenchmarks; benchID++ {
		allColumns = append(allColumns, order.GetOrder().Benchmarks.Values[benchID])
	}

	_, err := conn.Exec(c.commands.insertOrder, allColumns...)
	return err
}

func (c *sqlStorage) UpdateOrderStatus(conn queryConn, orderID *big.Int, status pb.OrderStatus) error {
	_, err := conn.Exec(c.commands.updateOrderStatus, status, orderID.String())
	return err
}

func (c *sqlStorage) UpdateOrders(conn queryConn, profile *pb.Profile) error {
	_, err := conn.Exec(c.commands.updateOrders,
		profile.IdentityLevel,
		profile.Name,
		profile.Country,
		profile.Certificates,
		profile.UserID.Unwrap().Hex())
	return err
}

func (c *sqlStorage) DeleteOrder(conn queryConn, orderID *big.Int) error {
	_, err := conn.Exec(c.commands.deleteOrder, orderID.String())
	return err
}

func (c *sqlStorage) GetOrderByID(conn queryConn, orderID *big.Int) (*pb.DWHOrder, error) {
	rows, err := conn.Query(c.commands.selectOrderByID, orderID.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to selectOrderByID")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return c.decodeOrder(rows)
}

func (c *sqlStorage) GetOrders(conn queryConn, request *pb.OrdersRequest) ([]*pb.DWHOrder, uint64, error) {
	var filters []*filter
	filters = append(filters, newFilter("Status", eq, pb.OrderStatus_ORDER_ACTIVE, "AND"))
	if !request.DealID.IsZero() {
		filters = append(filters, newFilter("DealID", eq, request.DealID.Unwrap().String(), "AND"))
	}
	if request.Type > 0 {
		filters = append(filters, newFilter("Type", eq, request.Type, "AND"))
	}
	if !request.AuthorID.IsZero() {
		filters = append(filters, newFilter("AuthorID", eq, request.AuthorID.Unwrap().Hex(), "AND"))
	}
	if !request.CounterpartyID.IsZero() {
		filters = append(filters, newFilter("CounterpartyID", eq, request.CounterpartyID.Unwrap().Hex(), "AND"))
	}
	if request.Duration != nil {
		if request.Duration.Max > 0 {
			filters = append(filters, newFilter("Duration", lte, request.Duration.Max, "AND"))
		}
		filters = append(filters, newFilter("Duration", gte, request.Duration.Min, "AND"))
	}
	if request.Price != nil {
		if request.Price.Max != nil {
			filters = append(filters, newFilter("Price", lte, request.Price.Max.PaddedString(), "AND"))
		}
		if request.Price.Min != nil {
			filters = append(filters, newFilter("Price", gte, request.Price.Min.PaddedString(), "AND"))
		}
	}
	if request.Netflags != nil && request.Netflags.Value > 0 {
		filters = append(filters, newNetflagsFilter(request.Netflags.Operator, request.Netflags.Value))
	}
	if request.CreatorIdentityLevel > 0 {
		filters = append(filters, newFilter("CreatorIdentityLevel", gte, request.CreatorIdentityLevel, "AND"))
	}
	if request.CreatedTS != nil {
		createdTS := request.CreatedTS
		if createdTS.Max != nil && createdTS.Max.Seconds > 0 {
			filters = append(filters, newFilter("CreatedTS", lte, createdTS.Max.Seconds, "AND"))
		}
		if createdTS.Min != nil && createdTS.Min.Seconds > 0 {
			filters = append(filters, newFilter("CreatedTS", gte, createdTS.Min.Seconds, "AND"))
		}
	}
	if request.Benchmarks != nil {
		c.addBenchmarksConditions(request.Benchmarks, &filters)
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Orders",
		filters:   filters,
		sortings:  c.filterSortings(request.Sortings, c.tablesInfo.OrderColumnsSet),
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := c.decodeOrder(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeOrder")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "rows error")
	}

	return orders, count, nil
}

func (c *sqlStorage) GetMatchingOrders(conn queryConn, request *pb.MatchingOrdersRequest) ([]*pb.DWHOrder, uint64, error) {
	order, err := c.GetOrderByID(conn, request.Id.Unwrap())
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to GetOrderByID")
	}

	var (
		filters      []*filter
		orderType    pb.OrderType
		priceOp      string
		durationOp   string
		benchOp      string
		sortingOrder pb.SortingOrder
	)
	if order.Order.OrderType == pb.OrderType_BID {
		orderType = pb.OrderType_ASK
		priceOp = lte
		durationOp = gte
		benchOp = gte
		sortingOrder = pb.SortingOrder_Asc
	} else {
		orderType = pb.OrderType_BID
		priceOp = gte
		durationOp = lte
		benchOp = lte
		sortingOrder = pb.SortingOrder_Desc
	}
	filters = append(filters, newFilter("Type", eq, orderType, "AND"))
	filters = append(filters, newFilter("Status", eq, pb.OrderStatus_ORDER_ACTIVE, "AND"))
	filters = append(filters, newFilter("Price", priceOp, order.Order.Price.PaddedString(), "AND"))
	if order.Order.Duration > 0 {
		filters = append(filters, newFilter("Duration", durationOp, order.Order.Duration, "AND"))
	} else {
		filters = append(filters, newFilter("Duration", eq, order.Order.Duration, "AND"))
	}
	if !order.Order.CounterpartyID.IsZero() {
		filters = append(filters, newFilter("AuthorID", eq, order.Order.CounterpartyID.Unwrap().Hex(), "AND"))
	}
	counterpartyFilter := newFilter("CounterpartyID", eq, common.Address{}.Hex(), "OR")
	counterpartyFilter.OpenBracket = true
	filters = append(filters, counterpartyFilter)
	counterpartyFilter = newFilter("CounterpartyID", eq, order.Order.AuthorID.Unwrap().Hex(), "AND")
	counterpartyFilter.CloseBracket = true
	filters = append(filters, counterpartyFilter)
	if order.Order.OrderType == pb.OrderType_BID {
		filters = append(filters, newNetflagsFilter(pb.CmpOp_GTE, order.Order.Netflags))
	} else {
		filters = append(filters, newNetflagsFilter(pb.CmpOp_LTE, order.Order.Netflags))
	}
	filters = append(filters, newFilter("IdentityLevel", gte, order.Order.IdentityLevel, "AND"))
	filters = append(filters, newFilter("CreatorIdentityLevel", lte, order.CreatorIdentityLevel, "AND"))
	for benchID, benchValue := range order.Order.Benchmarks.Values {
		filters = append(filters, newFilter(getBenchmarkColumn(uint64(benchID)), benchOp, benchValue, "AND"))
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Orders",
		filters:   filters,
		sortings:  []*pb.SortingOption{{Field: "Price", Order: sortingOrder}},
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run Query")
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := c.decodeOrder(rows)
		if err != nil {
			return nil, 0, status.Error(codes.Internal, "failed to GetMatchingOrders")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, status.Error(codes.Internal, "failed to GetMatchingOrders")
	}

	return orders, count, nil
}

func (c *sqlStorage) GetProfiles(conn queryConn, request *pb.ProfilesRequest) ([]*pb.Profile, uint64, error) {
	var filters []*filter
	switch request.Role {
	case pb.ProfileRole_Supplier:
		filters = append(filters, newFilter("ActiveAsks", gte, 1, "AND"))
	case pb.ProfileRole_Consumer:
		filters = append(filters, newFilter("ActiveBids", gte, 1, "AND"))
	}
	filters = append(filters, newFilter("IdentityLevel", gte, request.IdentityLevel, "AND"))
	if len(request.Country) > 0 {
		filters = append(filters, newFilter("Country", eq, request.Country, "AND"))
	}
	if len(request.Name) > 0 {
		filters = append(filters, newFilter("Name", "LIKE", request.Name, "AND"))
	}

	opts := &queryOpts{
		table:     "Profiles",
		filters:   filters,
		sortings:  c.filterSortings(request.Sortings, c.tablesInfo.ProfileColumnsSet),
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	}
	if request.BlacklistQuery != nil && request.BlacklistQuery.OwnerID != nil {
		opts.selectAs = "AS p"
		switch request.BlacklistQuery.Option {
		case pb.BlacklistOption_WithoutMatching:
			opts.customFilter = &customFilter{
				clause: c.commands.profileNotInBlacklist,
				values: []interface{}{request.BlacklistQuery.OwnerID.Unwrap().Hex()},
			}
		case pb.BlacklistOption_OnlyMatching:
			opts.customFilter = &customFilter{
				clause: c.commands.profileInBlacklist,
				values: []interface{}{request.BlacklistQuery.OwnerID.Unwrap().Hex()},
			}
		}
	}

	rows, count, err := c.queryRunner.Run(conn, opts)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var out []*pb.Profile
	for rows.Next() {
		if profile, err := c.decodeProfile(rows); err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeProfile")
		} else {
			out = append(out, profile)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "rows error")
	}

	if request.BlacklistQuery != nil && request.BlacklistQuery.Option == pb.BlacklistOption_IncludeAndMark {
		blacklistReply, err := c.GetBlacklist(conn, &pb.BlacklistRequest{OwnerID: request.BlacklistQuery.OwnerID})
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to")
		}

		var blacklistedAddrs = map[string]bool{}
		for _, blacklistedAddr := range blacklistReply.Addresses {
			blacklistedAddrs[blacklistedAddr] = true
		}

		for _, profile := range out {
			if blacklistedAddrs[profile.UserID.Unwrap().Hex()] {
				profile.IsBlacklisted = true
			}
		}
	}

	return out, count, nil
}

func (c *sqlStorage) InsertDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error {
	_, err := conn.Exec(c.commands.insertDealChangeRequest,
		changeRequest.Id.Unwrap().String(),
		changeRequest.CreatedTS.Seconds,
		changeRequest.RequestType,
		changeRequest.Duration,
		changeRequest.Price.PaddedString(),
		changeRequest.Status,
		changeRequest.DealID.Unwrap().String())
	return err
}

func (c *sqlStorage) UpdateDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error {
	_, err := conn.Exec(c.commands.updateDealChangeRequest, changeRequest.Status, changeRequest.Id.Unwrap().String())
	return err
}

func (c *sqlStorage) DeleteDealChangeRequest(conn queryConn, changeRequestID *big.Int) error {
	_, err := conn.Exec(c.commands.deleteDealChangeRequest, changeRequestID.String())
	return err
}

func (c *sqlStorage) GetDealChangeRequests(conn queryConn, changeRequest *pb.DealChangeRequest) ([]*pb.DealChangeRequest, error) {
	rows, err := conn.Query(c.commands.selectDealChangeRequests,
		changeRequest.DealID.Unwrap().String(),
		changeRequest.RequestType,
		changeRequest.Status)
	if err != nil {
		return nil, errors.Wrap(err, "failed to selectDealChangeRequests")
	}
	defer rows.Close()

	var out []*pb.DealChangeRequest
	for rows.Next() {
		changeRequest, err := c.decodeDealChangeRequest(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decodeDealChangeRequest")
		}
		out = append(out, changeRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *sqlStorage) GetDealChangeRequestsByID(conn queryConn, changeRequestID *big.Int) ([]*pb.DealChangeRequest, error) {
	rows, err := conn.Query(c.commands.selectDealChangeRequestsByID, changeRequestID.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to selectDealChangeRequests")
	}
	defer rows.Close()

	var out []*pb.DealChangeRequest
	for rows.Next() {
		changeRequest, err := c.decodeDealChangeRequest(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decodeDealChangeRequest")
		}
		out = append(out, changeRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (c *sqlStorage) InsertDealCondition(conn queryConn, condition *pb.DealCondition) error {
	_, err := conn.Exec(c.commands.insertDealCondition,
		condition.SupplierID.Unwrap().Hex(),
		condition.ConsumerID.Unwrap().Hex(),
		condition.MasterID.Unwrap().Hex(),
		condition.Duration,
		condition.Price.PaddedString(),
		condition.StartTime.Seconds,
		condition.EndTime.Seconds,
		condition.TotalPayout.PaddedString(),
		condition.DealID.Unwrap().String())
	return err
}

func (c *sqlStorage) UpdateDealConditionPayout(conn queryConn, dealConditionID uint64, payout *big.Int) error {
	_, err := conn.Exec(c.commands.updateDealConditionPayout, util.BigIntToPaddedString(payout), dealConditionID)
	return err
}

func (c *sqlStorage) UpdateDealConditionEndTime(conn queryConn, dealConditionID, eventTS uint64) error {
	_, err := conn.Exec(c.commands.updateDealConditionEndTime, eventTS, dealConditionID)
	return err
}

func (c *sqlStorage) InsertDealPayment(conn queryConn, payment *pb.DealPayment) error {
	_, err := conn.Exec(c.commands.insertDealPayment, payment.PaymentTS.Seconds, payment.PayedAmount.PaddedString(),
		payment.DealID.Unwrap().String())
	return err
}

func (c *sqlStorage) InsertWorker(conn queryConn, masterID, slaveID string) error {
	_, err := conn.Exec(c.commands.insertWorker, masterID, slaveID, false)
	return err
}

func (c *sqlStorage) UpdateWorker(conn queryConn, masterID, slaveID string) error {
	_, err := conn.Exec(c.commands.updateWorker, true, masterID, slaveID)
	return err
}

func (c *sqlStorage) DeleteWorker(conn queryConn, masterID, slaveID string) error {
	_, err := conn.Exec(c.commands.deleteWorker, masterID, slaveID)
	return err
}

func (c *sqlStorage) InsertBlacklistEntry(conn queryConn, adderID, addeeID string) error {
	_, err := conn.Exec(c.commands.insertBlacklistEntry, adderID, addeeID)
	return err
}

func (c *sqlStorage) DeleteBlacklistEntry(conn queryConn, removerID, removeeID string) error {
	_, err := conn.Exec(c.commands.deleteBlacklistEntry, removerID, removeeID)
	return err
}

func (c *sqlStorage) GetBlacklist(conn queryConn, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	var filters []*filter
	if !request.OwnerID.IsZero() {
		filters = append(filters, newFilter("AdderID", eq, request.OwnerID.Unwrap().Hex(), "AND"))
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Blacklists",
		filters:   filters,
		sortings:  []*pb.SortingOption{},
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var addees []string
	for rows.Next() {
		var (
			adderID string
			addeeID string
		)
		if err := rows.Scan(&adderID, &addeeID); err != nil {
			return nil, errors.Wrap(err, "failed to scan BlacklistAddress row")
		}

		addees = append(addees, addeeID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}

	return &pb.BlacklistReply{
		OwnerID:   request.OwnerID,
		Addresses: addees,
		Count:     count,
	}, nil
}

func (c *sqlStorage) InsertValidator(conn queryConn, validator *pb.Validator) error {
	_, err := conn.Exec(c.commands.insertValidator, validator.Id.Unwrap().Hex(), validator.Level)
	return err
}

func (c *sqlStorage) UpdateValidator(conn queryConn, validator *pb.Validator) error {
	_, err := conn.Exec(c.commands.updateValidator, validator.Level, validator.Id.Unwrap().Hex())
	return err
}

func (c *sqlStorage) InsertCertificate(conn queryConn, certificate *pb.Certificate) error {
	_, err := conn.Exec(c.commands.insertCertificate,
		certificate.OwnerID.Unwrap().Hex(),
		certificate.Attribute,
		(certificate.Attribute/uint64(100))%10,
		certificate.Value,
		certificate.ValidatorID.Unwrap().Hex())
	return err
}

func (c *sqlStorage) GetCertificates(conn queryConn, ownerID common.Address) ([]*pb.Certificate, error) {
	rows, err := conn.Query(c.commands.selectCertificates, ownerID.Hex())
	if err != nil {
		return nil, errors.Wrap(err, "failed to getCertificatesByUseID")
	}

	var (
		certificates     []*pb.Certificate
		maxIdentityLevel uint64
	)
	for rows.Next() {
		if certificate, err := c.decodeCertificate(rows); err != nil {
			return nil, errors.Wrap(err, "failed to decodeCertificate")
		} else {
			certificates = append(certificates, certificate)
			if certificate.IdentityLevel > maxIdentityLevel {
				maxIdentityLevel = certificate.IdentityLevel
			}
		}
	}

	return certificates, nil
}

func (c *sqlStorage) InsertProfileUserID(conn queryConn, profile *pb.Profile) error {
	_, err := conn.Exec(c.commands.insertProfileUserID,
		profile.UserID.Unwrap().Hex(), profile.Certificates, profile.ActiveAsks, profile.ActiveBids)
	return err
}

func (c *sqlStorage) GetProfileByID(conn queryConn, userID common.Address) (*pb.Profile, error) {
	rows, err := conn.Query(c.commands.selectProfileByID, userID.Hex())
	if err != nil {
		return nil, errors.Wrap(err, "failed to selectProfileByID")
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return c.decodeProfile(rows)
}

func (c *sqlStorage) GetValidators(conn queryConn, request *pb.ValidatorsRequest) ([]*pb.Validator, uint64, error) {
	var filters []*filter
	if request.ValidatorLevel != nil {
		level := request.ValidatorLevel
		filters = append(filters, newFilter("Level", opsTranslator[level.Operator], level.Value, "AND"))
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Validators",
		filters:   filters,
		sortings:  request.Sortings,
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var out []*pb.Validator
	for rows.Next() {
		validator, err := c.decodeValidator(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeValidator")
		}

		out = append(out, validator)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "rows error")
	}

	return out, count, nil
}

func (c *sqlStorage) GetWorkers(conn queryConn, request *pb.WorkersRequest) ([]*pb.DWHWorker, uint64, error) {
	var filters []*filter
	if !request.MasterID.IsZero() {
		filters = append(filters, newFilter("Level", eq, request.MasterID, "AND"))
	}
	rows, count, err := c.queryRunner.Run(conn, &queryOpts{
		table:     "Workers",
		filters:   filters,
		sortings:  []*pb.SortingOption{},
		offset:    request.Offset,
		limit:     request.Limit,
		withCount: request.WithCount,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to run query")
	}
	defer rows.Close()

	var out []*pb.DWHWorker
	for rows.Next() {
		worker, err := c.decodeWorker(rows)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to decodeWorker")
		}
		out = append(out, worker)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "rows error")
	}

	return out, count, nil
}

func (c *sqlStorage) UpdateProfile(conn queryConn, userID common.Address, field string, value interface{}) error {
	_, err := conn.Exec(fmt.Sprintf(c.commands.updateProfile, field), value, userID.Hex())
	return err
}

func (c *sqlStorage) GetLastKnownBlock(conn queryConn) (uint64, error) {
	rows, err := conn.Query(c.commands.selectLastKnownBlock)
	if err != nil {
		return 0, errors.Wrap(err, "failed to selectLastKnownBlock")
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return 0, errors.New("selectLastKnownBlock: no entries")
	}

	var lastKnownBlock uint64
	if err := rows.Scan(&lastKnownBlock); err != nil {
		return 0, errors.Wrapf(err, "failed to parse last known block number")
	}

	return lastKnownBlock, nil
}

func (c *sqlStorage) InsertLastKnownBlock(conn queryConn, blockNumber int64) error {
	_, err := conn.Exec(c.commands.insertLastKnownBlock, blockNumber)
	return err
}

func (c *sqlStorage) UpdateLastKnownBlock(conn queryConn, blockNumber int64) error {
	_, err := conn.Exec(c.commands.updateLastKnownBlock, blockNumber)
	return err
}

func (c *sqlStorage) addBenchmarksConditions(benches map[uint64]*pb.MaxMinUint64, filters *[]*filter) {
	for benchID, condition := range benches {
		if condition.Max > 0 {
			*filters = append(*filters, newFilter(getBenchmarkColumn(benchID), lte, condition.Max, "AND"))
		}
		if condition.Min > 0 {
			*filters = append(*filters, newFilter(getBenchmarkColumn(benchID), gte, condition.Max, "AND"))
		}
	}
}

func (c *sqlStorage) decodeDeal(rows *sql.Rows) (*pb.DWHDeal, error) {
	var (
		id                   = new(string)
		supplierID           = new(string)
		consumerID           = new(string)
		masterID             = new(string)
		askID                = new(string)
		bidID                = new(string)
		duration             = new(uint64)
		price                = new(string)
		startTime            = new(int64)
		endTime              = new(int64)
		dealStatus           = new(uint64)
		blockedBalance       = new(string)
		totalPayout          = new(string)
		lastBillTS           = new(int64)
		netflags             = new(uint64)
		askIdentityLevel     = new(uint64)
		bidIdentityLevel     = new(uint64)
		supplierCertificates = &[]byte{}
		consumerCertificates = &[]byte{}
		activeChangeRequest  = new(bool)
	)
	allFields := []interface{}{
		id,
		supplierID,
		consumerID,
		masterID,
		askID,
		bidID,
		duration,
		price,
		startTime,
		endTime,
		dealStatus,
		blockedBalance,
		totalPayout,
		lastBillTS,
		netflags,
		askIdentityLevel,
		bidIdentityLevel,
		supplierCertificates,
		consumerCertificates,
		activeChangeRequest,
	}
	benchmarks := make([]*uint64, c.numBenchmarks)
	for benchID := range benchmarks {
		benchmarks[benchID] = new(uint64)
		allFields = append(allFields, benchmarks[benchID])
	}
	if err := rows.Scan(allFields...); err != nil {
		return nil, errors.Wrap(err, "failed to scan Deal row")
	}

	benchmarksUint64 := make([]uint64, c.numBenchmarks)
	for benchID, benchValue := range benchmarks {
		benchmarksUint64[benchID] = *benchValue
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(*price, 10)
	bigBlockedBalance := new(big.Int)
	bigBlockedBalance.SetString(*blockedBalance, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(*totalPayout, 10)

	bigID, err := pb.NewBigIntFromString(*id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ID)")
	}

	bigAskID, err := pb.NewBigIntFromString(*askID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (askID)")
	}

	bigBidID, err := pb.NewBigIntFromString(*bidID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (bidID)")
	}

	return &pb.DWHDeal{
		Deal: &pb.Deal{
			Id:             bigID,
			SupplierID:     pb.NewEthAddress(common.HexToAddress(*supplierID)),
			ConsumerID:     pb.NewEthAddress(common.HexToAddress(*consumerID)),
			MasterID:       pb.NewEthAddress(common.HexToAddress(*masterID)),
			AskID:          bigAskID,
			BidID:          bigBidID,
			Price:          pb.NewBigInt(bigPrice),
			Duration:       *duration,
			StartTime:      &pb.Timestamp{Seconds: *startTime},
			EndTime:        &pb.Timestamp{Seconds: *endTime},
			Status:         pb.DealStatus(*dealStatus),
			BlockedBalance: pb.NewBigInt(bigBlockedBalance),
			TotalPayout:    pb.NewBigInt(bigTotalPayout),
			LastBillTS:     &pb.Timestamp{Seconds: *lastBillTS},
			Benchmarks:     &pb.Benchmarks{Values: benchmarksUint64},
		},
		Netflags:             *netflags,
		AskIdentityLevel:     *askIdentityLevel,
		BidIdentityLevel:     *bidIdentityLevel,
		SupplierCertificates: *supplierCertificates,
		ConsumerCertificates: *consumerCertificates,
		ActiveChangeRequest:  *activeChangeRequest,
	}, nil
}

func (c *sqlStorage) decodeDealCondition(rows *sql.Rows) (*pb.DealCondition, error) {
	var (
		id          uint64
		supplierID  string
		consumerID  string
		masterID    string
		duration    uint64
		price       string
		startTime   int64
		endTime     int64
		totalPayout string
		dealID      string
	)
	if err := rows.Scan(
		&id,
		&supplierID,
		&consumerID,
		&masterID,
		&duration,
		&price,
		&startTime,
		&endTime,
		&totalPayout,
		&dealID,
	); err != nil {
		return nil, errors.Wrap(err, "failed to scan DealCondition row")
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(totalPayout, 10)
	bigDealID, err := pb.NewBigIntFromString(dealID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (DealID)")
	}

	return &pb.DealCondition{
		Id:          id,
		SupplierID:  pb.NewEthAddress(common.HexToAddress(supplierID)),
		ConsumerID:  pb.NewEthAddress(common.HexToAddress(consumerID)),
		MasterID:    pb.NewEthAddress(common.HexToAddress(masterID)),
		Price:       pb.NewBigInt(bigPrice),
		Duration:    duration,
		StartTime:   &pb.Timestamp{Seconds: startTime},
		EndTime:     &pb.Timestamp{Seconds: endTime},
		TotalPayout: pb.NewBigInt(bigTotalPayout),
		DealID:      bigDealID,
	}, nil
}

func (c *sqlStorage) decodeOrder(rows *sql.Rows) (*pb.DWHOrder, error) {
	var (
		id                   = new(string)
		createdTS            = new(uint64)
		dealID               = new(string)
		orderType            = new(uint64)
		orderStatus          = new(uint64)
		author               = new(string)
		counterAgent         = new(string)
		duration             = new(uint64)
		price                = new(string)
		netflags             = new(uint64)
		identityLevel        = new(uint64)
		blacklist            = new(string)
		tag                  = &[]byte{}
		frozenSum            = new(string)
		creatorIdentityLevel = new(uint64)
		creatorName          = new(string)
		creatorCountry       = new(string)
		creatorCertificates  = &[]byte{}
	)
	allFields := []interface{}{
		id,
		createdTS,
		dealID,
		orderType,
		orderStatus,
		author,
		counterAgent,
		duration,
		price,
		netflags,
		identityLevel,
		blacklist,
		tag,
		frozenSum,
		creatorIdentityLevel,
		creatorName,
		creatorCountry,
		creatorCertificates,
	}
	benchmarks := make([]*uint64, c.numBenchmarks)
	for benchID := range benchmarks {
		benchmarks[benchID] = new(uint64)
		allFields = append(allFields, benchmarks[benchID])
	}
	if err := rows.Scan(allFields...); err != nil {
		return nil, errors.Wrap(err, "failed to scan Order row")
	}
	benchmarksUint64 := make([]uint64, c.numBenchmarks)
	for benchID, benchValue := range benchmarks {
		benchmarksUint64[benchID] = *benchValue
	}
	bigPrice, err := pb.NewBigIntFromString(*price)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (Price)")
	}
	bigFrozenSum, err := pb.NewBigIntFromString(*frozenSum)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (FrozenSum)")
	}
	bigID, err := pb.NewBigIntFromString(*id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ID)")
	}
	bigDealID, err := pb.NewBigIntFromString(*dealID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (DealID)")
	}

	return &pb.DWHOrder{
		Order: &pb.Order{
			Id:             bigID,
			DealID:         bigDealID,
			OrderType:      pb.OrderType(*orderType),
			OrderStatus:    pb.OrderStatus(*orderStatus),
			AuthorID:       pb.NewEthAddress(common.HexToAddress(*author)),
			CounterpartyID: pb.NewEthAddress(common.HexToAddress(*counterAgent)),
			Duration:       *duration,
			Price:          bigPrice,
			Netflags:       *netflags,
			IdentityLevel:  pb.IdentityLevel(*identityLevel),
			Blacklist:      *blacklist,
			Tag:            *tag,
			FrozenSum:      bigFrozenSum,
			Benchmarks:     &pb.Benchmarks{Values: benchmarksUint64},
		},
		CreatedTS:            &pb.Timestamp{Seconds: int64(*createdTS)},
		CreatorIdentityLevel: *creatorIdentityLevel,
		CreatorName:          *creatorName,
		CreatorCountry:       *creatorCountry,
		CreatorCertificates:  *creatorCertificates,
	}, nil
}

func (c *sqlStorage) decodeDealChangeRequest(rows *sql.Rows) (*pb.DealChangeRequest, error) {
	var (
		changeRequestID     string
		createdTS           uint64
		requestType         uint64
		duration            uint64
		price               string
		changeRequestStatus uint64
		dealID              string
	)
	err := rows.Scan(
		&changeRequestID,
		&createdTS,
		&requestType,
		&duration,
		&price,
		&changeRequestStatus,
		&dealID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to scan DealChangeRequest row")
	}
	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigDealID, err := pb.NewBigIntFromString(dealID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ID)")
	}

	bigChangeRequestID, err := pb.NewBigIntFromString(changeRequestID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ChangeRequestID)")
	}

	return &pb.DealChangeRequest{
		Id:          bigChangeRequestID,
		DealID:      bigDealID,
		RequestType: pb.OrderType(requestType),
		Duration:    duration,
		Price:       pb.NewBigInt(bigPrice),
		Status:      pb.ChangeRequestStatus(changeRequestStatus),
	}, nil
}

func (c *sqlStorage) decodeCertificate(rows *sql.Rows) (*pb.Certificate, error) {
	var (
		ownerID       string
		attribute     uint64
		identityLevel uint64
		value         []byte
		validatorID   string
	)
	if err := rows.Scan(&ownerID, &attribute, &identityLevel, &value, &validatorID); err != nil {
		return nil, errors.Wrap(err, "failed to decode Certificate")
	} else {
		return &pb.Certificate{
			OwnerID:       pb.NewEthAddress(common.HexToAddress(ownerID)),
			Attribute:     attribute,
			IdentityLevel: identityLevel,
			Value:         value,
			ValidatorID:   pb.NewEthAddress(common.HexToAddress(validatorID)),
		}, nil
	}
}

func (c *sqlStorage) decodeProfile(rows *sql.Rows) (*pb.Profile, error) {
	var (
		id             uint64
		userID         string
		identityLevel  uint64
		name           string
		country        string
		isCorporation  bool
		isProfessional bool
		certificates   []byte
		activeAsks     uint64
		activeBids     uint64
	)
	if err := rows.Scan(
		&id,
		&userID,
		&identityLevel,
		&name,
		&country,
		&isCorporation,
		&isProfessional,
		&certificates,
		&activeAsks,
		&activeBids,
	); err != nil {
		return nil, errors.Wrap(err, "failed to scan Profile row")
	}

	return &pb.Profile{
		UserID:         pb.NewEthAddress(common.HexToAddress(userID)),
		IdentityLevel:  identityLevel,
		Name:           name,
		Country:        country,
		IsCorporation:  isCorporation,
		IsProfessional: isProfessional,
		Certificates:   string(certificates),
		ActiveAsks:     activeAsks,
		ActiveBids:     activeBids,
	}, nil
}

func (c *sqlStorage) decodeValidator(rows *sql.Rows) (*pb.Validator, error) {
	var (
		validatorID string
		level       uint64
	)
	if err := rows.Scan(&validatorID, &level); err != nil {
		return nil, errors.Wrap(err, "failed to scan Validator row")
	}

	return &pb.Validator{
		Id:    pb.NewEthAddress(common.HexToAddress(validatorID)),
		Level: level,
	}, nil
}

func (c *sqlStorage) decodeWorker(rows *sql.Rows) (*pb.DWHWorker, error) {
	var (
		masterID  string
		slaveID   string
		confirmed bool
	)
	if err := rows.Scan(&masterID, &slaveID, &confirmed); err != nil {
		return nil, errors.Wrap(err, "failed to scan Worker row")
	}

	return &pb.DWHWorker{
		MasterID:  pb.NewEthAddress(common.HexToAddress(masterID)),
		SlaveID:   pb.NewEthAddress(common.HexToAddress(slaveID)),
		Confirmed: confirmed,
	}, nil
}

func (c *sqlStorage) filterSortings(sortings []*pb.SortingOption, columns map[string]bool) (out []*pb.SortingOption) {
	for _, sorting := range sortings {
		if columns[sorting.Field] {
			out = append(out, sorting)
		}
	}

	return out
}

type sqlCommands struct {
	insertDeal                   string
	updateDeal                   string
	updateDealsSupplier          string
	updateDealsConsumer          string
	updateDealPayout             string
	deleteDeal                   string
	selectDealByID               string
	insertOrder                  string
	updateOrderStatus            string
	updateOrders                 string
	deleteOrder                  string
	selectOrderByID              string
	insertDealChangeRequest      string
	updateDealChangeRequest      string
	deleteDealChangeRequest      string
	selectDealChangeRequests     string
	selectDealChangeRequestsByID string
	insertDealCondition          string
	updateDealConditionPayout    string
	updateDealConditionEndTime   string
	insertDealPayment            string
	insertWorker                 string
	updateWorker                 string
	deleteWorker                 string
	insertBlacklistEntry         string
	selectBlacklists             string
	deleteBlacklistEntry         string
	insertValidator              string
	updateValidator              string
	insertCertificate            string
	selectCertificates           string
	insertProfileUserID          string
	selectProfileByID            string
	profileNotInBlacklist        string
	profileInBlacklist           string
	updateProfile                string
	selectLastKnownBlock         string
	insertLastKnownBlock         string
	updateLastKnownBlock         string
}

type sqlSetupCommands struct {
	createTableDeals          string
	createTableDealConditions string
	createTableDealPayments   string
	createTableChangeRequests string
	createTableOrders         string
	createTableWorkers        string
	createTableBlacklists     string
	createTableValidators     string
	createTableCertificates   string
	createTableProfiles       string
	createTableMisc           string
	createIndexCmd            string
	tablesInfo                *tablesInfo
}

func (c *sqlSetupCommands) setupTables(db *sql.DB) error {
	_, err := db.Exec(c.createTableDeals)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDeals)
	}

	_, err = db.Exec(c.createTableDealConditions)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDealConditions)
	}

	_, err = db.Exec(c.createTableDealPayments)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableDealPayments)
	}

	_, err = db.Exec(c.createTableChangeRequests)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableChangeRequests)
	}

	_, err = db.Exec(c.createTableOrders)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableOrders)
	}

	_, err = db.Exec(c.createTableWorkers)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableWorkers)
	}

	_, err = db.Exec(c.createTableBlacklists)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableBlacklists)
	}

	_, err = db.Exec(c.createTableValidators)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableValidators)
	}

	_, err = db.Exec(c.createTableCertificates)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableCertificates)
	}

	_, err = db.Exec(c.createTableProfiles)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableProfiles)
	}

	_, err = db.Exec(c.createTableMisc)
	if err != nil {
		return errors.Wrapf(err, "failed to %s", c.createTableMisc)
	}

	return nil
}

func (c *sqlSetupCommands) createIndices(db *sql.DB) error {
	var err error
	for column := range c.tablesInfo.DealColumnsSet {
		if err = c.createIndex(db, c.createIndexCmd, "Deals", column); err != nil {
			return err
		}
	}
	for _, column := range []string{"Id", "DealID", "RequestType", "Status"} {
		if err = c.createIndex(db, c.createIndexCmd, "DealChangeRequests", column); err != nil {
			return err
		}
	}
	for column := range c.tablesInfo.DealConditionColumnsSet {
		if err = c.createIndex(db, c.createIndexCmd, "DealConditions", column); err != nil {
			return err
		}
	}
	for column := range c.tablesInfo.OrderColumnsSet {
		if err = c.createIndex(db, c.createIndexCmd, "Orders", column); err != nil {
			return err
		}
	}
	for _, column := range []string{"MasterID", "WorkerID"} {
		if err = c.createIndex(db, c.createIndexCmd, "Workers", column); err != nil {
			return err
		}
	}
	for _, column := range []string{"AdderID", "AddeeID"} {
		if err = c.createIndex(db, c.createIndexCmd, "Blacklists", column); err != nil {
			return err
		}
	}
	if err = c.createIndex(db, c.createIndexCmd, "Validators", "Id"); err != nil {
		return err
	}
	if err = c.createIndex(db, c.createIndexCmd, "Certificates", "OwnerID"); err != nil {
		return err
	}
	for column := range c.tablesInfo.ProfileColumnsSet {
		if err = c.createIndex(db, c.createIndexCmd, "Profiles", column); err != nil {
			return err
		}
	}

	return nil
}

func (c *sqlSetupCommands) createIndex(db *sql.DB, command, table, column string) error {
	cmd := fmt.Sprintf(command, table, column, table, column)
	_, err := db.Exec(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to %s (%s)", cmd)
	}

	return nil
}

type customFilter struct {
	clause string
	values []interface{}
}

type filter struct {
	Field        string
	CmpOperator  string
	BoolOperator string
	OpenBracket  bool
	CloseBracket bool
	Value        interface{}
}

func newFilter(field string, cmpOperator string, value interface{}, boolOperator string) *filter {
	return &filter{
		Field:        field,
		CmpOperator:  cmpOperator,
		BoolOperator: boolOperator,
		Value:        value,
	}
}

func newNetflagsFilter(operator pb.CmpOp, value uint64) *filter {
	switch operator {
	case pb.CmpOp_GTE:
		return newFilter("Netflags", fmt.Sprintf(" | ~%d = ", value), -1, "AND")
	case pb.CmpOp_LTE:
		return newFilter("", fmt.Sprintf("%d | ~Netflags = ", value), -1, "AND")
	default:
		return newFilter("Netflags", eq, value, "AND")
	}
}

// queryRunner implements DB-specific querying using queryOpts.
type queryRunner interface {
	// Run must return the obtained rows, a count of all rows that can be obtained (if queryOpts.withCount
	// is set to `true`) and an error. If queryOpts.withCount is `false`, 0 count is returned.
	Run(tx queryConn, opts *queryOpts) (*sql.Rows, uint64, error)
}

type queryOpts struct {
	table        string
	filters      []*filter
	sortings     []*pb.SortingOption
	offset       uint64
	limit        uint64
	customFilter *customFilter
	selectAs     string
	withCount    bool
}

// formatArg is a callback that inserts an SQL placeholder into query (e.g., ? for SQLIte of $1, $2, etc.
// for Postgres).
type formatArg func(argID uint64, lastArg bool) string

// tablesInfo is used to get static column names for tables with variable columns set (i.e., with benchmarks).
type tablesInfo struct {
	DealColumns             []string
	DealColumnsSet          map[string]bool
	NumDealColumns          uint64
	OrderColumns            []string
	OrderColumnsSet         map[string]bool
	NumOrderColumns         uint64
	ProfileColumnsSet       map[string]bool
	DealConditionColumnsSet map[string]bool
}

func newTablesInfo(numBenchmarks uint64) *tablesInfo {
	dealColumns := []string{
		"Id",
		"SupplierID",
		"ConsumerID",
		"MasterID",
		"AskID",
		"BidID",
		"Duration",
		"Price",
		"StartTime",
		"EndTime",
		"Status",
		"BlockedBalance",
		"TotalPayout",
		"LastBillTS",
		"Netflags",
		"AskIdentityLevel",
		"BidIdentityLevel",
		"SupplierCertificates",
		"ConsumerCertificates",
		"ActiveChangeRequest",
	}
	orderColumns := []string{
		"Id",
		"CreatedTS",
		"DealID",
		"Type",
		"Status",
		"AuthorID",
		"CounterpartyID",
		"Duration",
		"Price",
		"Netflags",
		"IdentityLevel",
		"Blacklist",
		"Tag",
		"FrozenSum",
		"CreatorIdentityLevel",
		"CreatorName",
		"CreatorCountry",
		"CreatorCertificates",
	}
	dealConditionColumns := []string{
		"Id",
		"SupplierID",
		"ConsumerID",
		"MasterID",
		"Duration",
		"Price",
		"StartTime",
		"EndTime",
		"TotalPayout",
		"DealID",
	}
	profileColumns := []string{
		"Id",
		"UserID",
		"IdentityLevel",
		"Name",
		"Country",
		"IsCorporation",
		"IsProfessional",
		"Certificates",
	}
	out := &tablesInfo{
		DealColumns:             dealColumns,
		DealColumnsSet:          stringSliceToSet(dealColumns),
		NumDealColumns:          uint64(len(dealColumns)),
		OrderColumns:            orderColumns,
		OrderColumnsSet:         stringSliceToSet(orderColumns),
		NumOrderColumns:         uint64(len(orderColumns)),
		DealConditionColumnsSet: stringSliceToSet(dealConditionColumns),
		ProfileColumnsSet:       stringSliceToSet(profileColumns),
	}
	for benchmarkID := uint64(0); benchmarkID < numBenchmarks; benchmarkID++ {
		out.DealColumns = append(out.DealColumns, getBenchmarkColumn(uint64(benchmarkID)))
		out.DealColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
		out.OrderColumns = append(out.OrderColumns, getBenchmarkColumn(uint64(benchmarkID)))
		out.OrderColumnsSet[getBenchmarkColumn(uint64(benchmarkID))] = true
	}

	return out
}

func makeInsertDealQuery(format string, formatCb formatArg, numBenchmarks uint64, tInfo *tablesInfo) string {
	dealPlaceholders := ""
	for i := uint64(0); i < tInfo.NumDealColumns; i++ {
		dealPlaceholders += formatCb(i, false)
	}
	for i := tInfo.NumDealColumns; i < tInfo.NumDealColumns+numBenchmarks; i++ {
		if i == numBenchmarks+tInfo.NumDealColumns-1 {
			dealPlaceholders += formatCb(i, true)
		} else {
			dealPlaceholders += formatCb(i, false)
		}
	}
	return fmt.Sprintf(format, strings.Join(tInfo.DealColumns, ", "), dealPlaceholders)
}

func makeSelectDealByIDQuery(format string, tInfo *tablesInfo) string {
	return fmt.Sprintf(format, strings.Join(tInfo.DealColumns, ", "))
}

func makeInsertOrderQuery(format string, formatCb formatArg, numBenchmarks uint64, tInfo *tablesInfo) string {
	orderPlaceholders := ""
	for i := uint64(0); i < tInfo.NumOrderColumns; i++ {
		orderPlaceholders += formatCb(i, false)
	}
	for i := tInfo.NumOrderColumns; i < tInfo.NumOrderColumns+numBenchmarks; i++ {
		if i == numBenchmarks+tInfo.NumOrderColumns-1 {
			orderPlaceholders += formatCb(i, true)
		} else {
			orderPlaceholders += formatCb(i, false)
		}
	}
	return fmt.Sprintf(format, strings.Join(tInfo.OrderColumns, ", "), orderPlaceholders)
}

func makeSelectOrderByIDQuery(format string, tInfo *tablesInfo) string {
	return fmt.Sprintf(format, strings.Join(tInfo.OrderColumns, ", "))
}

func makeTableWithBenchmarks(format, benchmarkType string) string {
	benchmarkColumns := make([]string, NumMaxBenchmarks)
	for benchmarkID := uint64(0); benchmarkID < NumMaxBenchmarks; benchmarkID++ {
		benchmarkColumns[benchmarkID] = fmt.Sprintf("%s %s", getBenchmarkColumn(uint64(benchmarkID)), benchmarkType)
	}
	return strings.Join(append([]string{format}, benchmarkColumns...), ",\n") + ")"
}
