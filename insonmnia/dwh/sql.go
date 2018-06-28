package dwh

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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
	setupCommands *sqlSetupCommands
	numBenchmarks uint64
	tablesInfo    *tablesInfo
	builder       func() sq.StatementBuilderType
}

func (m *sqlStorage) Setup(db *sql.DB) error {
	return m.setupCommands.setupTables(db)
}

func (m *sqlStorage) CreateIndices(db *sql.DB) error {
	return m.setupCommands.createIndices(db)
}

func (m *sqlStorage) InsertDeal(conn queryConn, deal *pb.Deal) error {
	ask, err := m.GetOrderByID(conn, deal.AskID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to getOrderDetails (Ask, `%s`): %v", deal.GetAskID().Unwrap().String(), err)
	}

	bid, err := m.GetOrderByID(conn, deal.BidID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to getOrderDetails (Ask, `%s`): %v", deal.GetBidID().Unwrap().String(), err)
	}

	var hasActiveChangeRequests bool
	if _, err := m.GetDealChangeRequestsByDealID(conn, deal.Id.Unwrap()); err == nil {
		hasActiveChangeRequests = true
	}
	values := []interface{}{
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
		ask.GetOrder().GetNetflags().GetFlags(),
		ask.GetOrder().IdentityLevel,
		bid.GetOrder().IdentityLevel,
		ask.CreatorCertificates,
		bid.CreatorCertificates,
		hasActiveChangeRequests,
	}
	benchmarks := deal.GetBenchmarks().GetNValues(m.numBenchmarks)
	for idx, benchmarkValue := range benchmarks {
		if benchmarkValue >= MaxBenchmark {
			return fmt.Errorf("deal benchmark %d is greater than %d", idx, MaxBenchmark)
		}
		values = append(values, benchmarkValue)
	}

	query, args, _ := m.builder().Insert("Deals").
		Columns(m.tablesInfo.DealColumns...).
		Values(values...).
		ToSql()
	_, err = conn.Exec(query, args...)

	return err
}

func (m *sqlStorage) UpdateDeal(conn queryConn, deal *pb.Deal) error {
	query, args, _ := m.builder().Update("Deals").SetMap(map[string]interface{}{
		"Duration":       deal.Duration,
		"Price":          deal.Price.PaddedString(),
		"StartTime":      deal.StartTime.Seconds,
		"EndTime":        deal.EndTime.Seconds,
		"Status":         uint64(deal.Status),
		"BlockedBalance": deal.BlockedBalance.PaddedString(),
		"TotalPayout":    deal.TotalPayout.PaddedString(),
		"LastBillTS":     deal.LastBillTS.Seconds,
	}).Where("Id = ?", deal.Id.Unwrap().String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealsSupplier(conn queryConn, profile *pb.Profile) error {
	query, args, _ := m.builder().Update("Deals").SetMap(map[string]interface{}{
		"SupplierCertificates": []byte(profile.Certificates),
	}).Where("SupplierID = ?", profile.UserID.Unwrap().Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealsConsumer(conn queryConn, profile *pb.Profile) error {
	query, args, _ := m.builder().Update("Deals").SetMap(map[string]interface{}{
		"ConsumerCertificates": []byte(profile.Certificates),
	}).Where("ConsumerID = ?", profile.UserID.Unwrap().Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealPayout(conn queryConn, dealID, payout *big.Int, billTS uint64) error {
	query, args, _ := m.builder().Update("Deals").SetMap(map[string]interface{}{
		"TotalPayout": util.BigIntToPaddedString(payout),
		"LastBillTS":  billTS,
	}).Where("Id = ?", dealID.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteDeal(conn queryConn, dealID *big.Int) error {
	query, args, _ := m.builder().Delete("Deals").Where("Id = ?", dealID.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetDealByID(conn queryConn, dealID *big.Int) (*pb.DWHDeal, error) {
	query, args, _ := m.builder().Select(m.tablesInfo.DealColumns...).
		From("Deals").
		Where("Id = ?", dealID.String()).
		ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to GetDealByID: %v", err)
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.New("no rows returned")
	}

	return m.decodeDeal(rows)
}

func (m *sqlStorage) GetDeals(conn queryConn, r *pb.DealsRequest) ([]*pb.DWHDeal, uint64, error) {
	builder := m.builder().Select("*").From("Deals")

	if r.Status > 0 {
		builder = builder.Where("Status = ?", r.Status)
	}
	if !r.GetAnyUserID().IsZero() {
		builder = builder.Where(sq.Or{
			sq.Expr("SupplierID = ?", r.GetAnyUserID().Unwrap().Hex()),
			sq.Expr("ConsumerID = ?", r.GetAnyUserID().Unwrap().Hex()),
			sq.Expr("MasterID = ?", r.GetAnyUserID().Unwrap().Hex()),
		})
	} else {
		if !r.SupplierID.IsZero() {
			builder = builder.Where("SupplierID = ?", r.SupplierID.Unwrap().Hex())
		}
		if !r.ConsumerID.IsZero() {
			builder = builder.Where("ConsumerID = ?", r.ConsumerID.Unwrap().Hex())
		}
		if !r.MasterID.IsZero() {
			builder = builder.Where("MasterID = ?", r.MasterID.Unwrap().Hex())
		}
	}
	if !r.AskID.IsZero() {
		builder = builder.Where("AskID = ?", r.AskID)
	}
	if !r.BidID.IsZero() {
		builder = builder.Where("BidID = ?", r.BidID)
	}
	if r.Duration != nil {
		if r.Duration.Max > 0 {
			builder = builder.Where("Duration <= ?", r.Duration.Max)
		}
		builder = builder.Where("Duration >= ?", r.Duration.Min)
	}
	if r.Price != nil {
		if r.Price.Max != nil {
			builder = builder.Where("Price <= ?", r.Price.Max.PaddedString())
		}
		if r.Price.Min != nil {
			builder = builder.Where("Price >= ?", r.Price.Min.PaddedString())
		}
	}
	if r.Netflags != nil && r.Netflags.Value > 0 {
		builder = m.builderWithNetflagsFilter(builder, r.Netflags.Operator, r.Netflags.Value)
	}
	if r.AskIdentityLevel > 0 {
		builder = builder.Where("AskIdentityLevel >= ?", r.AskIdentityLevel)
	}
	if r.BidIdentityLevel > 0 {
		builder = builder.Where("BidIdentityLevel >= ?", r.BidIdentityLevel)
	}
	if r.Benchmarks != nil {
		builder = m.builderWithBenchmarkFilters(builder, r.Benchmarks)
	}
	if r.Offset > 0 {
		builder = builder.Offset(r.Offset)
	}

	builder = m.builderWithSortings(builder, r.Sortings)
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, strings.Join(m.tablesInfo.DealColumns, ", "), r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to runQuery: %v", err)
	}
	defer rows.Close()

	var deals []*pb.DWHDeal
	for rows.Next() {
		deal, err := m.decodeDeal(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeDeal: %v", err)
		}

		deals = append(deals, deal)
	}

	return deals, count, nil
}

func (m *sqlStorage) GetDealConditions(conn queryConn, r *pb.DealConditionsRequest) ([]*pb.DealCondition, uint64, error) {
	builder := m.builder().Select("*").From("DealConditions")
	builder = builder.Where("DealID = ?", r.DealID.Unwrap().String())
	if len(r.Sortings) == 0 {
		builder = m.builderWithSortings(builder, []*pb.SortingOption{{Field: "Id", Order: pb.SortingOrder_Desc}})
	}
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*pb.DealCondition
	for rows.Next() {
		dealCondition, err := m.decodeDealCondition(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeDealCondition: %v", err)
		}
		out = append(out, dealCondition)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, status.Error(codes.Internal, "failed to GetDealConditions")
	}

	return out, count, nil
}

func (m *sqlStorage) InsertOrder(conn queryConn, order *pb.DWHOrder) error {
	values := []interface{}{
		order.GetOrder().Id.Unwrap().String(),
		order.MasterID.Unwrap().String(),
		order.CreatedTS.Seconds,
		order.GetOrder().DealID.Unwrap().String(),
		uint64(order.GetOrder().OrderType),
		uint64(order.GetOrder().OrderStatus),
		order.GetOrder().AuthorID.Unwrap().Hex(),
		order.GetOrder().CounterpartyID.Unwrap().Hex(),
		order.GetOrder().Duration,
		order.GetOrder().Price.PaddedString(),
		order.GetOrder().GetNetflags().GetFlags(),
		uint64(order.GetOrder().IdentityLevel),
		order.GetOrder().Blacklist,
		order.GetOrder().Tag,
		order.GetOrder().FrozenSum.PaddedString(),
		order.CreatorIdentityLevel,
		order.CreatorName,
		order.CreatorCountry,
		[]byte(order.CreatorCertificates),
	}
	benchmarks := order.GetOrder().GetBenchmarks().GetNValues(m.numBenchmarks)
	for idx, benchmarkValue := range benchmarks {
		if benchmarkValue >= MaxBenchmark {
			return fmt.Errorf("order benchmark %d is greater than %d", idx, MaxBenchmark)
		}
		values = append(values, benchmarkValue)
	}
	query, args, _ := m.builder().Insert("Orders").Columns(m.tablesInfo.OrderColumns...).Values(values...).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateOrderStatus(conn queryConn, orderID *big.Int, status pb.OrderStatus) error {
	query, args, _ := m.builder().Update("Orders").Set("Status", status).Where("Id = ?", orderID.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateOrders(conn queryConn, profile *pb.Profile) error {
	query, args, _ := m.builder().Update("Orders").SetMap(map[string]interface{}{
		"CreatorIdentityLevel": profile.IdentityLevel,
		"CreatorName":          profile.Name,
		"CreatorCountry":       profile.Country,
		"CreatorCertificates":  profile.Certificates,
	}).Where("AuthorId = ?", profile.UserID.Unwrap().Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteOrder(conn queryConn, orderID *big.Int) error {
	query, args, _ := m.builder().Delete("Orders").Where("Id = ?", orderID.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetOrderByID(conn queryConn, orderID *big.Int) (*pb.DWHOrder, error) {
	query, args, _ := m.builder().Select(m.tablesInfo.OrderColumns...).
		From("Orders").
		Where("Id = ?", orderID.String()).
		ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to selectOrderByID: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return m.decodeOrder(rows)
}

func (m *sqlStorage) GetOrders(conn queryConn, r *pb.OrdersRequest) ([]*pb.DWHOrder, uint64, error) {
	builder := m.builder().Select("*").From("Orders AS o").
		Where("Status = ?", pb.OrderStatus_ORDER_ACTIVE)
	if !r.DealID.IsZero() {
		builder = builder.Where("DealID = ?", r.DealID.Unwrap().String())
	}
	if r.Type > 0 {
		builder = builder.Where("Type = ?", r.Type)
	}
	if !r.AuthorID.IsZero() {
		builder = builder.Where("AuthorID LIKE ?", r.AuthorID.Unwrap().Hex())
	}
	if !r.MasterID.IsZero() {
		builder = builder.Where("MasterID LIKE ?", r.MasterID.Unwrap().Hex())
	}
	if len(r.CounterpartyID) > 0 {
		var ids []string
		for _, id := range r.CounterpartyID {
			ids = append(ids, id.Unwrap().Hex())
		}
		builder = builder.Where(sq.Eq{"CounterpartyID": ids})
	}
	if r.Duration != nil {
		if r.Duration.Max > 0 {
			builder = builder.Where("Duration <= ?", r.Duration.Max)
		}
		builder = builder.Where("Duration >= ?", r.Duration.Min)
	}
	if r.Price != nil {
		if r.Price.Max != nil {
			builder = builder.Where("Price <= ?", r.Price.Max.PaddedString())
		}
		if r.Price.Min != nil {
			builder = builder.Where("Price >= ?", r.Price.Min.PaddedString())
		}
	}
	if r.Netflags != nil && r.Netflags.Value > 0 {
		builder = m.builderWithNetflagsFilter(builder, r.Netflags.Operator, r.Netflags.Value)
	}
	if len(r.CreatorIdentityLevel) > 0 {
		builder = builder.Where(sq.Eq{"CreatorIdentityLevel": r.CreatorIdentityLevel})
	}
	if r.CreatedTS != nil {
		createdTS := r.CreatedTS
		if createdTS.Max != nil && createdTS.Max.Seconds > 0 {
			builder = builder.Where("CreatedTS <= ?", createdTS.Max.Seconds)
		}
		if createdTS.Min != nil && createdTS.Min.Seconds > 0 {
			builder = builder.Where("CreatedTS >= ?", createdTS.Min.Seconds)
		}
	}
	if r.Benchmarks != nil {
		builder = m.builderWithBenchmarkFilters(builder, r.Benchmarks)
	}

	if len(r.SenderIDs) > 0 {
		var senderIDs []string
		for _, id := range r.SenderIDs {
			senderIDs = append(senderIDs, id.Unwrap().Hex())
		}
		builder = m.builderWithBlacklistFilters(builder, senderIDs, senderIDs)
	}

	builder = m.builderWithSortings(builder, r.Sortings)
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, strings.Join(m.tablesInfo.OrderColumns, ", "), r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := m.decodeOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeOrder: %v", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	return orders, count, nil
}

func (m *sqlStorage) GetMatchingOrders(conn queryConn, r *pb.MatchingOrdersRequest) ([]*pb.DWHOrder, uint64, error) {
	order, err := m.GetOrderByID(conn, r.Id.Unwrap())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to GetOrderByID: %v", err)
	}

	builder := m.builder().Select("*").From("Orders AS o")
	var (
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
	builder = builder.Where("Type = ?", orderType).
		Where("Status = ?", pb.OrderStatus_ORDER_ACTIVE).
		Where(fmt.Sprintf("Price %s ?", priceOp), order.Order.Price.PaddedString())
	builder = builder.Where(fmt.Sprintf("Duration %s ?", durationOp), order.Order.Duration)
	if !order.Order.CounterpartyID.IsZero() {
		builder = builder.Where(sq.Or{
			sq.Eq{"AuthorID": order.Order.CounterpartyID.Unwrap().Hex()},
			sq.Eq{"MasterID": order.Order.CounterpartyID.Unwrap().Hex()},
		})
	}
	builder = builder.Where(sq.Eq{
		"CounterpartyID": []string{
			common.Address{}.Hex(),
			order.Order.AuthorID.Unwrap().Hex(),
			order.MasterID.Unwrap().Hex()},
	})
	if order.Order.OrderType == pb.OrderType_BID {
		builder = m.builderWithNetflagsFilter(builder, pb.CmpOp_GTE, order.Order.GetNetflags().GetFlags())
	} else {
		builder = m.builderWithNetflagsFilter(builder, pb.CmpOp_LTE, order.Order.GetNetflags().GetFlags())
	}
	builder = builder.Where("IdentityLevel >= ?", order.Order.IdentityLevel).
		Where("CreatorIdentityLevel <= ?", order.CreatorIdentityLevel)
	for benchID, benchValue := range order.Order.Benchmarks.Values {
		builder = builder.Where(fmt.Sprintf("%s %s ?", getBenchmarkColumn(uint64(benchID)), benchOp), benchValue)
	}
	builder = m.builderWithSortings(builder, []*pb.SortingOption{{Field: "Price", Order: sortingOrder}})
	var (
		masterID    = order.MasterID.Unwrap().Hex()
		authorID    = order.GetOrder().AuthorID.Unwrap().Hex()
		blacklistID = order.GetOrder().GetBlacklist()
	)
	builder = m.builderWithBlacklistFilters(builder, []string{masterID, authorID},
		[]string{masterID, authorID, blacklistID})

	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, strings.Join(m.tablesInfo.OrderColumns, ", "), r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run Query: %v", err)
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := m.decodeOrder(rows)
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

func (m *sqlStorage) GetProfiles(conn queryConn, r *pb.ProfilesRequest) ([]*pb.Profile, uint64, error) {
	builder := m.builder().Select("*").From("Profiles AS p")
	switch r.Role {
	case pb.ProfileRole_Supplier:
		builder = builder.Where("ActiveAsks >= 1")
	case pb.ProfileRole_Consumer:
		builder = builder.Where("ActiveBids >= 1")
	}
	builder = builder.Where("IdentityLevel >= ?", r.IdentityLevel)
	if len(r.Country) > 0 {
		builder = builder.Where(sq.Eq{"Country": r.Country})
	}
	if len(r.Name) > 0 {
		builder = builder.Where("lower(Name) LIKE lower(?)", r.Name)
	}
	if r.BlacklistQuery != nil && !r.BlacklistQuery.OwnerID.IsZero() {
		ownerBuilder := m.builder().Select("AddeeID").From("Blacklists").
			Where("AdderID = ?", r.BlacklistQuery.OwnerID.Unwrap().Hex()).Where("AddeeID = p.UserID")
		ownerQuery, _, _ := ownerBuilder.ToSql()
		if r.BlacklistQuery != nil && r.BlacklistQuery.OwnerID != nil {
			switch r.BlacklistQuery.Option {
			case pb.BlacklistOption_WithoutMatching:
				builder = builder.Where(fmt.Sprintf("UserID NOT IN (%s)", ownerQuery))
			case pb.BlacklistOption_OnlyMatching:
				builder = builder.Where(fmt.Sprintf("UserID IN (%s)", ownerQuery))
			}
		}
	}
	builder = m.builderWithSortings(builder, r.Sortings)
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()

	if r.BlacklistQuery != nil && !r.BlacklistQuery.OwnerID.IsZero() {
		args = append(args, r.BlacklistQuery.OwnerID.Unwrap().Hex())
	}

	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*pb.Profile
	for rows.Next() {
		if profile, err := m.decodeProfile(rows); err != nil {
			return nil, 0, fmt.Errorf("failed to decodeProfile: %v", err)
		} else {
			out = append(out, profile)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	if r.BlacklistQuery != nil && r.BlacklistQuery.Option == pb.BlacklistOption_IncludeAndMark {
		blacklistReply, err := m.GetBlacklist(conn, &pb.BlacklistRequest{UserID: r.BlacklistQuery.OwnerID})
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get blacklist: %v", err)
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

func (m *sqlStorage) InsertDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error {
	query, args, _ := m.builder().Insert("DealChangeRequests").
		Columns(m.tablesInfo.DealChangeRequestColumns...).
		Values(
			changeRequest.Id.Unwrap().String(),
			changeRequest.CreatedTS.Seconds,
			changeRequest.RequestType,
			changeRequest.Duration,
			changeRequest.Price.PaddedString(),
			changeRequest.Status,
			changeRequest.DealID.Unwrap().String(),
		).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealChangeRequest(conn queryConn, changeRequest *pb.DealChangeRequest) error {
	query, args, _ := m.builder().Update("DealChangeRequests").Set("Status", changeRequest.Status).
		Where("Id = ?", changeRequest.Id.Unwrap().String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteDealChangeRequest(conn queryConn, changeRequestID *big.Int) error {
	query, args, _ := m.builder().Delete("DealChangeRequests").Where("Id = ?", changeRequestID.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetDealChangeRequests(conn queryConn, changeRequest *pb.DealChangeRequest) ([]*pb.DealChangeRequest, error) {
	query, args, _ := m.builder().Select(m.tablesInfo.DealChangeRequestColumns...).
		From("DealChangeRequests").Where("DealID = ?", changeRequest.DealID.Unwrap().String()).
		Where("RequestType = ?", changeRequest.RequestType).
		Where("Status = ?", changeRequest.Status).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select DealChangeRequests: %v", err)
	}
	defer rows.Close()

	var out []*pb.DealChangeRequest
	for rows.Next() {
		changeRequest, err := m.decodeDealChangeRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to decodeDealChangeRequest: %v", err)
		}
		out = append(out, changeRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (m *sqlStorage) GetDealChangeRequestsByDealID(conn queryConn, changeRequestID *big.Int) ([]*pb.DealChangeRequest, error) {
	query, args, _ := m.builder().Select(m.tablesInfo.DealChangeRequestColumns...).
		From("DealChangeRequests").
		Where("DealID = ?", changeRequestID.String()).
		ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select DealChangeRequestsByID: %v", err)
	}
	defer rows.Close()

	var out []*pb.DealChangeRequest
	for rows.Next() {
		changeRequest, err := m.decodeDealChangeRequest(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to decodeDealChangeRequest: %v", err)
		}
		out = append(out, changeRequest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (m *sqlStorage) InsertDealCondition(conn queryConn, condition *pb.DealCondition) error {
	query, args, err := m.builder().Insert("DealConditions").Columns(m.tablesInfo.DealConditionColumns[1:]...).
		Values(
			condition.SupplierID.Unwrap().Hex(),
			condition.ConsumerID.Unwrap().Hex(),
			condition.MasterID.Unwrap().Hex(),
			condition.Duration,
			condition.Price.PaddedString(),
			condition.StartTime.Seconds,
			condition.EndTime.Seconds,
			condition.TotalPayout.PaddedString(),
			condition.DealID.Unwrap().String(),
		).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealConditionPayout(conn queryConn, dealConditionID uint64, payout *big.Int) error {
	query, args, err := m.builder().Update("DealConditions").Set("TotalPayout", util.BigIntToPaddedString(payout)).
		Where("Id = ?", dealConditionID).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealConditionEndTime(conn queryConn, dealConditionID, eventTS uint64) error {
	query, args, err := m.builder().Update("DealConditions").Set("EndTime", eventTS).
		Where("Id = ?", dealConditionID).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) CheckWorkerExists(conn queryConn, masterID, workerID common.Address) (bool, error) {
	query, args, _ := m.builder().Select("MasterID").From("Workers").
		Where("MasterID = ?", masterID.Hex()).Where("WorkerID = ?", workerID.Hex()).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to run CheckWorker query: %v", err)
	}
	defer rows.Close()
	return rows.Next(), nil
}

func (m *sqlStorage) InsertWorker(conn queryConn, masterID, workerID common.Address) error {
	query, args, err := m.builder().Insert("Workers").Values(masterID.Hex(), workerID.Hex(), false).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateWorker(conn queryConn, masterID, workerID common.Address) error {
	query, args, err := m.builder().Update("Workers").Set("Confirmed", true).Where("MasterID = ?", masterID.Hex()).
		Where("WorkerID = ?", workerID.Hex()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteWorker(conn queryConn, masterID, workerID common.Address) error {
	query, args, err := m.builder().Delete("Workers").Where("MasterID = ?", masterID.Hex()).
		Where("WorkerID = ?", workerID.Hex()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetMasterByWorker(conn queryConn, slaveID common.Address) (common.Address, error) {
	query, args, _ := m.builder().Select("MasterID").From("Workers").
		Where("WorkerID = ?", slaveID.Hex()).
		Where("Confirmed = ?", true).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to GetMasterByWorker: %v", err)
	}
	defer rows.Close()
	var masterID string
	if !rows.Next() {
		return common.Address{}, errors.New("no rows returned")
	}
	if err := rows.Scan(&masterID); err != nil {
		return common.Address{}, fmt.Errorf("failed to scan MasterID row: %v", err)
	}
	return util.HexToAddress(masterID)
}

func (m *sqlStorage) InsertBlacklistEntry(conn queryConn, adderID, addeeID common.Address) error {
	query, args, err := m.builder().Insert("Blacklists").Values(adderID.Hex(), addeeID.Hex()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteBlacklistEntry(conn queryConn, removerID, removeeID common.Address) error {
	query, args, err := m.builder().Delete("Blacklists").Where("AdderID = ?", removerID.Hex()).
		Where("AddeeID = ?", removeeID.Hex()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetBlacklist(conn queryConn, r *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	builder := m.builder().Select("*").From("Blacklists")

	if !r.UserID.IsZero() {
		builder = builder.Where("AdderID = ?", r.UserID.Unwrap().Hex())
	}
	builder = m.builderWithSortings(builder, []*pb.SortingOption{})
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var addees []string
	for rows.Next() {
		var (
			adderID string
			addeeID string
		)
		if err := rows.Scan(&adderID, &addeeID); err != nil {
			return nil, fmt.Errorf("failed to scan BlacklistAddress row: %v", err)
		}

		addees = append(addees, addeeID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &pb.BlacklistReply{
		OwnerID:   r.UserID,
		Addresses: addees,
		Count:     count,
	}, nil
}

func (m *sqlStorage) GetBlacklistsContainingUser(conn queryConn, r *pb.BlacklistRequest) (*pb.BlacklistsContainingUserReply, error) {
	if r.UserID.IsZero() {
		return nil, errors.New("UserID must be specified")
	}
	query, args, _ := m.builder().Select("AdderID").From("Blacklists").
		Where("AddeeID = ?", r.UserID.Unwrap().Hex()).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var adders []*pb.EthAddress
	for rows.Next() {
		var adderID string
		if err := rows.Scan(&adderID); err != nil {
			return nil, fmt.Errorf("failed to scan BlacklistAddress row: %v", err)
		}

		ethAddress, err := util.HexToAddress(adderID)
		if err != nil {
			return nil, fmt.Errorf("failed to use `%s` as EthAddress", adderID)
		}
		adders = append(adders, pb.NewEthAddress(ethAddress))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &pb.BlacklistsContainingUserReply{
		Blacklists: adders,
		Count:      count,
	}, nil
}

func (m *sqlStorage) InsertOrUpdateValidator(conn queryConn, validator *pb.Validator) error {
	// Validators are never deleted, so it's O.K. to check in a non-atomic way.
	query, args, _ := m.builder().Select("Id").From("Validators").Where("Id = ?", validator.GetId().Unwrap().Hex()).
		ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to check if Validator exists: %v", err)
	}
	alreadyExists := rows.Next()
	rows.Close()
	if alreadyExists {
		// If this validator exists, it means that it was deactivated; we re-activate it by setting the current
		// identity level.
		return m.UpdateValidator(conn, validator.GetId().Unwrap(), "Level", validator.GetLevel())
	}

	query, args, _ = m.builder().Insert("Validators").Columns("Id", "Level").
		Values(validator.Id.Unwrap().Hex(), validator.Level).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetValidator(conn queryConn, validatorID common.Address) (*pb.DWHValidator, error) {
	query, args, _ := m.builder().Select("*").From("Validators").Where("Id = ?", validatorID.Hex()).ToSql()
	rows, err := conn.Query(query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}
	return m.decodeValidator(rows)
}

func (m *sqlStorage) UpdateValidator(conn queryConn, validatorID common.Address, field string, value interface{}) error {
	if !m.tablesInfo.IsValidatorColumn(field) {
		// Ignore.
		return nil
	}
	if field == "KYC_Price" {
		if bytes, ok := value.([]byte); ok {
			value = pb.NewBigInt(big.NewInt(0).SetBytes(bytes)).PaddedString()
		}
	}
	query, args, _ := m.builder().Update("Validators").Set(field, value).Where("Id = ?", validatorID.Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeactivateValidator(conn queryConn, validatorID common.Address) error {
	// Deactivate validator by setting her identity level to zero.
	query, args, _ := m.builder().Update("Validators").Set("Level", 0).Where("Id = ?", validatorID.Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) InsertCertificate(conn queryConn, certificate *pb.Certificate) error {
	query, args, _ := m.builder().Insert("Certificates").Values(
		certificate.GetId().Unwrap().String(),
		certificate.OwnerID.Unwrap().Hex(),
		certificate.Attribute,
		(certificate.Attribute/uint64(100))%10,
		certificate.Value,
		certificate.ValidatorID.Unwrap().Hex(),
	).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetCertificates(conn queryConn, ownerID common.Address) ([]*pb.Certificate, error) {
	query, args, _ := m.builder().Select("*").From("Certificates").Where("OwnerID = ?", ownerID.Hex()).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to getCertificatesByUseID: %v", err)
	}
	defer rows.Close()

	var (
		certificates     []*pb.Certificate
		maxIdentityLevel uint64
	)
	for rows.Next() {
		if certificate, err := m.decodeCertificate(rows); err != nil {
			return nil, fmt.Errorf("failed to decodeCertificate: %v", err)
		} else {
			certificates = append(certificates, certificate)
			if certificate.IdentityLevel > maxIdentityLevel {
				maxIdentityLevel = certificate.IdentityLevel
			}
		}
	}

	return certificates, nil
}

func (m *sqlStorage) InsertProfileUserID(conn queryConn, profile *pb.Profile) error {
	query, args, _ := m.builder().Select("Id").From("Profiles").Where("UserID = ?", profile.UserID.Unwrap().Hex()).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to check if profile exists: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		// Profile already exists.
		return nil
	}

	query, args, _ = m.builder().Insert("Profiles").Columns(m.tablesInfo.ProfileColumns[1:]...).Values(
		profile.UserID.Unwrap().Hex(),
		0, "", "", false, false,
		profile.Certificates,
		profile.ActiveAsks,
		profile.ActiveBids,
	).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetProfileByID(conn queryConn, userID common.Address) (*pb.Profile, error) {
	query, args, _ := m.builder().Select("*").From("Profiles").Where("UserID = ?", userID.Hex()).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to GettProfileByID: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return m.decodeProfile(rows)
}

func (m *sqlStorage) GetValidators(conn queryConn, r *pb.ValidatorsRequest) ([]*pb.DWHValidator, uint64, error) {
	builder := m.builder().Select("*").From("Validators")
	if r.ValidatorLevel != nil {
		level := r.ValidatorLevel
		builder = builder.Where(fmt.Sprintf("Level %s ?", opsTranslator[level.Operator]), level.Value)
	}
	builder = m.builderWithSortings(builder, r.Sortings)
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*pb.DWHValidator
	for rows.Next() {
		validator, err := m.decodeValidator(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeValidator: %v", err)
		}
		out = append(out, validator)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	return out, count, nil
}

func (m *sqlStorage) GetWorkers(conn queryConn, r *pb.WorkersRequest) ([]*pb.DWHWorker, uint64, error) {
	builder := m.builder().Select("*").From("Workers")
	if !r.MasterID.IsZero() {
		builder = builder.Where("MasterID = ?", r.MasterID.Unwrap().String())
	}
	builder = m.builderWithSortings(builder, []*pb.SortingOption{})
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*pb.DWHWorker
	for rows.Next() {
		worker, err := m.decodeWorker(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeWorker: %v", err)
		}
		out = append(out, worker)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	return out, count, nil
}

func (m *sqlStorage) UpdateProfile(conn queryConn, userID common.Address, field string, value interface{}) error {
	query, args, _ := m.builder().Update("Profiles").Set(field, value).Where("UserID = ?", userID.Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateProfileStats(conn queryConn, userID common.Address, field string, value int) error {
	query, args, _ := m.builder().Update("Profiles").
		Set(field, sq.Expr(fmt.Sprintf("%s + %d", field, value))).
		Where("UserID = ?", userID.Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetLastKnownBlock(conn queryConn) (uint64, error) {
	query, _, _ := m.builder().Select("LastKnownBlock").From("Misc").Where("Id = 1").ToSql()
	rows, err := conn.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to GetLastKnownBlock: %v", err)
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return 0, errors.New("getLastKnownBlock: no entries")
	}

	var lastKnownBlock uint64
	if err := rows.Scan(&lastKnownBlock); err != nil {
		return 0, fmt.Errorf("failed to parse last known block number: %v", err)
	}

	return lastKnownBlock, nil
}

func (m *sqlStorage) InsertLastKnownBlock(conn queryConn, blockNumber int64) error {
	query, args, _ := m.builder().Insert("Misc").Columns("LastKnownBlock").Values(blockNumber).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateLastKnownBlock(conn queryConn, blockNumber int64) error {
	query, args, _ := m.builder().Update("Misc").Set("LastKnownBlock", blockNumber).Where("Id = 1").ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) StoreStaleID(conn queryConn, id *big.Int, entity string) error {
	query, args, _ := m.builder().Insert("StaleIDs").Values(fmt.Sprintf("%s_%s", entity, id.String())).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) RemoveStaleID(conn queryConn, id *big.Int, entity string) error {
	query, args, _ := m.builder().Delete("StaleIDs").Where("Id = ?", fmt.Sprintf("%s_%s", entity, id.String())).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) CheckStaleID(conn queryConn, id *big.Int, entity string) (bool, error) {
	query, args, _ := m.builder().Select("*").From("StaleIDs").
		Where("Id = ?", fmt.Sprintf("%s_%s", entity, id.String())).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	return true, nil
}

func (m *sqlStorage) builderWithBenchmarkFilters(builder sq.SelectBuilder, benches map[uint64]*pb.MaxMinUint64) sq.SelectBuilder {
	for benchID, condition := range benches {
		if condition.Max >= condition.Min {
			builder = builder.Where(sq.And{
				sq.Expr(fmt.Sprintf("%s <= ?", getBenchmarkColumn(benchID)), condition.Max),
				sq.Expr(fmt.Sprintf("%s >= ?", getBenchmarkColumn(benchID)), condition.Min),
			})
		} else {
			builder = builder.Where(fmt.Sprintf("%s >= ?", getBenchmarkColumn(benchID)), condition.Min)
		}
	}

	return builder
}

// builderWithBlacklistFilters filters orders that:
// 	1. have our Master/Author in their Master/Author/Blacklist blacklist,
//	2. whose Master/Author is in our Master/Author/Blacklist blacklist.
func (m *sqlStorage) builderWithBlacklistFilters(builder sq.SelectBuilder, addees, adders []string) sq.SelectBuilder {
	blacklistsQuery := m.builder().Select("*").Prefix("NOT EXISTS (").Suffix(")").From("Blacklists AS b").
		Where(sq.Or{
			sq.And{
				sq.Expr("b.AdderID IN (o.MasterID, o.AuthorID, o.Blacklist)"),
				sq.Eq{"b.AddeeID": addees},
			},
			sq.And{
				sq.Eq{"b.AdderID": adders},
				sq.Expr("b.AddeeID IN (o.MasterID, o.AuthorID)"),
			},
		})
	return builder.Where(blacklistsQuery)
}

func (m *sqlStorage) builderWithOffsetLimit(builder sq.SelectBuilder, limit, offset uint64) sq.SelectBuilder {
	if limit > 0 {
		builder = builder.Limit(limit)
	}
	if offset > 0 {
		builder = builder.Offset(offset)
	}

	return builder
}

func (m *sqlStorage) builderWithSortings(builder sq.SelectBuilder, sortings []*pb.SortingOption) sq.SelectBuilder {
	var sortsFlat []string
	for _, sort := range sortings {
		sortsFlat = append(sortsFlat, fmt.Sprintf("%s %s", sort.Field, pb.SortingOrder_name[int32(sort.Order)]))
	}
	builder = builder.OrderBy(sortsFlat...)

	return builder
}

func (m *sqlStorage) builderWithNetflagsFilter(builder sq.SelectBuilder, operator pb.CmpOp, value uint64) sq.SelectBuilder {
	switch operator {
	case pb.CmpOp_GTE:
		return builder.Where("Netflags | ~ CAST (? as int) = -1", value)
	case pb.CmpOp_LTE:
		return builder.Where("? | ~Netflags = -1", value)
	default:
		return builder.Where("Netflags = ?", value)
	}
}

func (m *sqlStorage) runQuery(conn queryConn, columns string, withCount bool, query string, args ...interface{}) (*sql.Rows, uint64, error) {
	dataQuery := strings.Replace(query, "*", columns, 1)
	var count uint64
	var err error
	if withCount {
		count, err = m.getCount(conn, query, args...)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to getCount: %v", err)
		}
	}

	rows, err := conn.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("data query `%s` failed: %v", dataQuery, err)
	}

	return rows, count, nil
}

func (m *sqlStorage) getCount(conn queryConn, query string, args ...interface{}) (uint64, error) {
	var count uint64
	var countQuery = strings.Replace(query, "*", "count(*)", 1)
	countQuery = strings.Split(countQuery, "ORDER BY")[0]
	countRows, err := conn.Query(countQuery, args...)
	if err != nil {
		return 0, fmt.Errorf("count query `%s` failed: %v", countQuery, err)
	}
	defer countRows.Close()

	for countRows.Next() {
		countRows.Scan(&count)
	}

	return count, nil
}

func (m *sqlStorage) decodeDeal(rows *sql.Rows) (*pb.DWHDeal, error) {
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
	benchmarks := make([]*uint64, m.numBenchmarks)
	for benchID := range benchmarks {
		benchmarks[benchID] = new(uint64)
		allFields = append(allFields, benchmarks[benchID])
	}
	if err := rows.Scan(allFields...); err != nil {
		return nil, fmt.Errorf("failed to scan Deal row: %v", err)
	}

	benchmarksUint64 := make([]uint64, m.numBenchmarks)
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
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}

	bigAskID, err := pb.NewBigIntFromString(*askID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (askID): %v", err)
	}

	bigBidID, err := pb.NewBigIntFromString(*bidID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (bidID): %v", err)
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

func (m *sqlStorage) decodeDealCondition(rows *sql.Rows) (*pb.DealCondition, error) {
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
		return nil, fmt.Errorf("failed to scan DealCondition row: %v", err)
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(totalPayout, 10)
	bigDealID, err := pb.NewBigIntFromString(dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (DealID): %v", err)
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

func (m *sqlStorage) decodeOrder(rows *sql.Rows) (*pb.DWHOrder, error) {
	var (
		id                   = new(string)
		masterID             = new(string)
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
		masterID,
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
	benchmarks := make([]*uint64, m.numBenchmarks)
	for benchID := range benchmarks {
		benchmarks[benchID] = new(uint64)
		allFields = append(allFields, benchmarks[benchID])
	}
	if err := rows.Scan(allFields...); err != nil {
		return nil, fmt.Errorf("failed to scan Order row: %v", err)
	}
	benchmarksUint64 := make([]uint64, m.numBenchmarks)
	for benchID, benchValue := range benchmarks {
		benchmarksUint64[benchID] = *benchValue
	}
	bigPrice, err := pb.NewBigIntFromString(*price)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (Price): %v", err)
	}
	bigFrozenSum, err := pb.NewBigIntFromString(*frozenSum)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (FrozenSum): %v", err)
	}
	bigID, err := pb.NewBigIntFromString(*id)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}
	bigDealID, err := pb.NewBigIntFromString(*dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (DealID): %v", err)
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
			Netflags:       &pb.NetFlags{Flags: *netflags},
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
		MasterID:             pb.NewEthAddress(common.HexToAddress(*masterID)),
	}, nil
}

func (m *sqlStorage) decodeDealChangeRequest(rows *sql.Rows) (*pb.DealChangeRequest, error) {
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
		return nil, fmt.Errorf("failed to scan DealChangeRequest row: %v", err)
	}
	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigDealID, err := pb.NewBigIntFromString(dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}

	bigChangeRequestID, err := pb.NewBigIntFromString(changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ChangeRequestID): %v", err)
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

func (m *sqlStorage) decodeCertificate(rows *sql.Rows) (*pb.Certificate, error) {
	var (
		id            string
		ownerID       string
		attribute     uint64
		identityLevel uint64
		value         []byte
		validatorID   string
	)
	if err := rows.Scan(&id, &ownerID, &attribute, &identityLevel, &value, &validatorID); err != nil {
		return nil, fmt.Errorf("failed to decode Certificate: %v", err)
	} else {
		bigID, err := pb.NewBigIntFromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate id: %v", err)
		}

		return &pb.Certificate{
			Id:            bigID,
			OwnerID:       pb.NewEthAddress(common.HexToAddress(ownerID)),
			Attribute:     attribute,
			IdentityLevel: identityLevel,
			Value:         value,
			ValidatorID:   pb.NewEthAddress(common.HexToAddress(validatorID)),
		}, nil
	}
}

func (m *sqlStorage) decodeProfile(rows *sql.Rows) (*pb.Profile, error) {
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
		return nil, fmt.Errorf("failed to scan Profile row: %v", err)
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

func (m *sqlStorage) decodeValidator(rows *sql.Rows) (*pb.DWHValidator, error) {
	var (
		validatorID string
		level       uint64
		name        string
		logo        string
		url         string
		description string
		price       string
	)
	if err := rows.Scan(&validatorID, &level, &name, &logo, &url, &description, &price); err != nil {
		return nil, fmt.Errorf("failed to scan Validator row: %v", err)
	}

	bigPrice, err := pb.NewBigIntFromString(price)
	if err != nil {
		return nil, fmt.Errorf("failed to use price as big int: %s", price)
	}
	return &pb.DWHValidator{
		Validator: &pb.Validator{
			Id:    pb.NewEthAddress(common.HexToAddress(validatorID)),
			Level: level,
		},
		Name:        name,
		Logo:        logo,
		Url:         url,
		Description: description,
		Price:       bigPrice,
	}, nil
}

func (m *sqlStorage) decodeWorker(rows *sql.Rows) (*pb.DWHWorker, error) {
	var (
		masterID  string
		slaveID   string
		confirmed bool
	)
	if err := rows.Scan(&masterID, &slaveID, &confirmed); err != nil {
		return nil, fmt.Errorf("failed to scan Worker row: %v", err)
	}

	return &pb.DWHWorker{
		MasterID:  pb.NewEthAddress(common.HexToAddress(masterID)),
		SlaveID:   pb.NewEthAddress(common.HexToAddress(slaveID)),
		Confirmed: confirmed,
	}, nil
}

func (m *sqlStorage) filterSortings(sortings []*pb.SortingOption, columns map[string]bool) (out []*pb.SortingOption) {
	for _, sorting := range sortings {
		if columns[sorting.Field] {
			out = append(out, sorting)
		}
	}

	return out
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
	createTableStaleIDs       string
	createIndexCmd            string
	tablesInfo                *tablesInfo
}

func (c *sqlSetupCommands) setupTables(db *sql.DB) error {
	_, err := db.Exec(c.createTableDeals)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableDeals, err)
	}

	_, err = db.Exec(c.createTableDealConditions)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableDealConditions, err)
	}

	_, err = db.Exec(c.createTableDealPayments)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableDealPayments, err)
	}

	_, err = db.Exec(c.createTableChangeRequests)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableChangeRequests, err)
	}

	_, err = db.Exec(c.createTableOrders)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableOrders, err)
	}

	_, err = db.Exec(c.createTableWorkers)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableWorkers, err)
	}

	_, err = db.Exec(c.createTableBlacklists)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableBlacklists, err)
	}

	_, err = db.Exec(c.createTableValidators)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableValidators, err)
	}

	_, err = db.Exec(c.createTableCertificates)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableCertificates, err)
	}

	_, err = db.Exec(c.createTableProfiles)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableProfiles, err)
	}

	_, err = db.Exec(c.createTableStaleIDs)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableStaleIDs, err)
	}

	_, err = db.Exec(c.createTableMisc)
	if err != nil {
		return fmt.Errorf("failed to %s: %v", c.createTableMisc, err)
	}

	return nil
}

func (c *sqlSetupCommands) createIndices(db *sql.DB) error {
	var err error
	for _, column := range c.tablesInfo.DealColumns {
		if err = c.createIndex(db, c.createIndexCmd, "Deals", column); err != nil {
			return err
		}
	}
	for _, column := range []string{"Id", "DealID", "RequestType", "Status"} {
		if err = c.createIndex(db, c.createIndexCmd, "DealChangeRequests", column); err != nil {
			return err
		}
	}
	for _, column := range c.tablesInfo.DealConditionColumns {
		if err = c.createIndex(db, c.createIndexCmd, "DealConditions", column); err != nil {
			return err
		}
	}
	for _, column := range c.tablesInfo.OrderColumns {
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
	for _, column := range c.tablesInfo.ProfileColumns {
		if err = c.createIndex(db, c.createIndexCmd, "Profiles", column); err != nil {
			return err
		}
	}
	if err = c.createIndex(db, c.createIndexCmd, "StaleIDs", "Id"); err != nil {
		return err
	}

	return nil
}

func (c *sqlSetupCommands) createIndex(db *sql.DB, command, table, column string) error {
	cmd := fmt.Sprintf(command, table, column, table, column)
	_, err := db.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to %s (%s): %v", cmd, table, err)
	}

	return nil
}

// tablesInfo is used to get static column names for tables with variable columns set (i.e., with benchmarks).
type tablesInfo struct {
	DealColumns              []string
	NumDealColumns           uint64
	OrderColumns             []string
	NumOrderColumns          uint64
	DealConditionColumns     []string
	DealChangeRequestColumns []string
	ProfileColumns           []string
	ValidatorColumns         []string
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
		"MasterID",
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
	dealChangeRequestColumns := []string{
		"Id",
		"CreatedTS",
		"RequestType",
		"Duration",
		"Price",
		"Status",
		"DealID",
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
		"ActiveAsks",
		"ActiveBids",
	}
	validatorColumns := []string{
		"Id",
		"Level",
		"Name",
		"Logo",
		"KYC_URL",
		"Description",
		"KYC_Price",
	}
	out := &tablesInfo{
		DealColumns:              dealColumns,
		NumDealColumns:           uint64(len(dealColumns)),
		OrderColumns:             orderColumns,
		NumOrderColumns:          uint64(len(orderColumns)),
		DealChangeRequestColumns: dealChangeRequestColumns,
		DealConditionColumns:     dealConditionColumns,
		ProfileColumns:           profileColumns,
		ValidatorColumns:         validatorColumns,
	}
	for benchmarkID := uint64(0); benchmarkID < numBenchmarks; benchmarkID++ {
		out.DealColumns = append(out.DealColumns, getBenchmarkColumn(uint64(benchmarkID)))
		out.OrderColumns = append(out.OrderColumns, getBenchmarkColumn(uint64(benchmarkID)))
	}

	return out
}

func (m *tablesInfo) IsValidatorColumn(column string) bool {
	for _, validatorColumn := range m.ValidatorColumns {
		if validatorColumn == column {
			return true
		}
	}
	return false
}

func makeTableWithBenchmarks(format, benchmarkType string) string {
	benchmarkColumns := make([]string, NumMaxBenchmarks)
	for benchmarkID := uint64(0); benchmarkID < NumMaxBenchmarks; benchmarkID++ {
		benchmarkColumns[benchmarkID] = fmt.Sprintf("%s %s", getBenchmarkColumn(uint64(benchmarkID)), benchmarkType)
	}
	return strings.Join(append([]string{format}, benchmarkColumns...), ",\n") + ")"
}

func newPostgresStorage(numBenchmarks uint64) *sqlStorage {
	tInfo := newTablesInfo(numBenchmarks)
	storage := &sqlStorage{
		setupCommands: &sqlSetupCommands{
			createTableDeals: makeTableWithBenchmarks(`
	CREATE TABLE IF NOT EXISTS Deals (
		Id						TEXT UNIQUE NOT NULL,
		SupplierID				TEXT NOT NULL,
		ConsumerID				TEXT NOT NULL,
		MasterID				TEXT NOT NULL,
		AskID					TEXT NOT NULL,
		BidID					TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		StartTime				INTEGER NOT NULL,
		EndTime					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		BlockedBalance			TEXT NOT NULL,
		TotalPayout				TEXT NOT NULL,
		LastBillTS				INTEGER NOT NULL,
		Netflags				INTEGER NOT NULL,
		AskIdentityLevel		INTEGER NOT NULL,
		BidIdentityLevel		INTEGER NOT NULL,
		SupplierCertificates    BYTEA NOT NULL,
		ConsumerCertificates    BYTEA NOT NULL,
		ActiveChangeRequest     BOOLEAN NOT NULL`, `BIGINT DEFAULT 0`),
			createTableDealConditions: `
	CREATE TABLE IF NOT EXISTS DealConditions (
		Id							BIGSERIAL PRIMARY KEY,
		SupplierID					TEXT NOT NULL,
		ConsumerID					TEXT NOT NULL,
		MasterID					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		StartTime					INTEGER NOT NULL,
		EndTime						INTEGER NOT NULL,
		TotalPayout					TEXT NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			createTableDealPayments: `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE,
		UNIQUE						(BillTS, PaidAmount, DealID)
	)`,
			createTableChangeRequests: `
	CREATE TABLE IF NOT EXISTS DealChangeRequests (
		Id 							TEXT UNIQUE NOT NULL,
		CreatedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			createTableOrders: makeTableWithBenchmarks(`
	CREATE TABLE IF NOT EXISTS Orders (
		Id						TEXT UNIQUE NOT NULL,
		MasterID				TEXT NOT NULL,
		CreatedTS				INTEGER NOT NULL,
		DealID					TEXT NOT NULL,
		Type					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		AuthorID				TEXT NOT NULL,
		CounterpartyID			TEXT NOT NULL,
		Duration 				BIGINT NOT NULL,
		Price					TEXT NOT NULL,
		Netflags				INTEGER NOT NULL,
		IdentityLevel			INTEGER NOT NULL,
		Blacklist				TEXT NOT NULL,
		Tag						BYTEA NOT NULL,
		FrozenSum				TEXT NOT NULL,
		CreatorIdentityLevel	INTEGER NOT NULL,
		CreatorName				TEXT NOT NULL,
		CreatorCountry			TEXT NOT NULL,
		CreatorCertificates		BYTEA NOT NULL`, `BIGINT DEFAULT 0`),
			createTableWorkers: `
	CREATE TABLE IF NOT EXISTS Workers (
		MasterID					TEXT NOT NULL,
		WorkerID					TEXT NOT NULL,
		Confirmed					BOOLEAN NOT NULL,
		UNIQUE						(MasterID, WorkerID)
	)`,
			createTableBlacklists: `
	CREATE TABLE IF NOT EXISTS Blacklists (
		AdderID						TEXT NOT NULL,
		AddeeID						TEXT NOT NULL,
		UNIQUE						(AdderID, AddeeID)
	)`,
			createTableValidators: `
	CREATE TABLE IF NOT EXISTS Validators (
		Id							TEXT UNIQUE NOT NULL,
		Level						INTEGER NOT NULL,
		Name						TEXT NOT NULL DEFAULT '',
		Logo						TEXT NOT NULL DEFAULT '',
		KYC_URL						TEXT NOT NULL DEFAULT '',
		Description					TEXT NOT NULL DEFAULT '',
		KYC_Price					TEXT NOT NULL DEFAULT '0'
	)`,
			createTableCertificates: `
	CREATE TABLE IF NOT EXISTS Certificates (
		Id						    TEXT NOT NULL,
		OwnerID						TEXT NOT NULL,
		Attribute					INTEGER NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						BYTEA NOT NULL,
		ValidatorID					TEXT NOT NULL REFERENCES Validators(Id) ON DELETE CASCADE
	)`,
			createTableProfiles: `
	CREATE TABLE IF NOT EXISTS Profiles (
		Id							BIGSERIAL PRIMARY KEY,
		UserID						TEXT UNIQUE NOT NULL,
		IdentityLevel				INTEGER NOT NULL,
		Name						TEXT NOT NULL,
		Country						TEXT NOT NULL,
		IsCorporation				BOOLEAN NOT NULL,
		IsProfessional				BOOLEAN NOT NULL,
		Certificates				BYTEA NOT NULL,
		ActiveAsks					INTEGER NOT NULL,
		ActiveBids					INTEGER NOT NULL
	)`,
			createTableMisc: `
	CREATE TABLE IF NOT EXISTS Misc (
		Id							BIGSERIAL PRIMARY KEY,
		LastKnownBlock				INTEGER NOT NULL
	)`,
			createTableStaleIDs: `
	CREATE TABLE IF NOT EXISTS StaleIDs (
		Id 							TEXT NOT NULL
	)`,
			createIndexCmd: `CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)`,
			tablesInfo:     tInfo,
		},
		numBenchmarks: numBenchmarks,
		tablesInfo:    tInfo,
		builder: func() sq.StatementBuilderType {
			return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
		},
	}

	return storage
}

func newSQLiteStorage(numBenchmarks uint64) *sqlStorage {
	tInfo := newTablesInfo(numBenchmarks)
	storage := &sqlStorage{
		setupCommands: &sqlSetupCommands{
			// Incomplete, modified during setup.
			createTableDeals: makeTableWithBenchmarks(`
	CREATE TABLE IF NOT EXISTS Deals (
		Id						TEXT UNIQUE NOT NULL,
		SupplierID				TEXT NOT NULL,
		ConsumerID				TEXT NOT NULL,
		MasterID				TEXT NOT NULL,
		AskID					TEXT NOT NULL,
		BidID					TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		StartTime				INTEGER NOT NULL,
		EndTime					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		BlockedBalance			TEXT NOT NULL,
		TotalPayout				TEXT NOT NULL,
		LastBillTS				INTEGER NOT NULL,
		Netflags				INTEGER NOT NULL,
		AskIdentityLevel		INTEGER NOT NULL,
		BidIdentityLevel		INTEGER NOT NULL,
		SupplierCertificates    BLOB NOT NULL,
		ConsumerCertificates    BLOB NOT NULL,
		ActiveChangeRequest     INTEGER NOT NULL`, `INTEGER DEFAULT 0`),
			createTableDealConditions: `
	CREATE TABLE IF NOT EXISTS DealConditions (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		SupplierID					TEXT NOT NULL,
		ConsumerID					TEXT NOT NULL,
		MasterID					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		StartTime					INTEGER NOT NULL,
		EndTime						INTEGER NOT NULL,
		TotalPayout					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			createTableDealPayments: `
	CREATE TABLE IF NOT EXISTS DealPayments (
		BillTS						INTEGER NOT NULL,
		PaidAmount					TEXT NOT NULL,
		DealID						TEXT NOT NULL,
		UNIQUE						(BillTS, PaidAmount, DealID),
		FOREIGN KEY (DealID) 		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			createTableChangeRequests: `
	CREATE TABLE IF NOT EXISTS DealChangeRequests (
		Id 							TEXT UNIQUE NOT NULL,
		CreatedTS					INTEGER NOT NULL,
		RequestType					TEXT NOT NULL,
		Duration 					INTEGER NOT NULL,
		Price						TEXT NOT NULL,
		Status						INTEGER NOT NULL,
		DealID						TEXT NOT NULL,
		FOREIGN KEY (DealID)		REFERENCES Deals(Id) ON DELETE CASCADE
	)`,
			// Incomplete, modified during setup.
			createTableOrders: makeTableWithBenchmarks(`
	CREATE TABLE IF NOT EXISTS Orders (
		Id						TEXT UNIQUE NOT NULL,
		MasterID				TEXT NOT NULL,
		CreatedTS				INTEGER NOT NULL,
		DealID					TEXT NOT NULL,
		Type					INTEGER NOT NULL,
		Status					INTEGER NOT NULL,
		AuthorID				TEXT NOT NULL,
		CounterpartyID			TEXT NOT NULL,
		Duration 				INTEGER NOT NULL,
		Price					TEXT NOT NULL,
		Netflags				INTEGER NOT NULL,
		IdentityLevel			INTEGER NOT NULL,
		Blacklist				TEXT NOT NULL,
		Tag						BLOB NOT NULL,
		FrozenSum				TEXT NOT NULL,
		CreatorIdentityLevel	INTEGER NOT NULL,
		CreatorName				TEXT NOT NULL,
		CreatorCountry			TEXT NOT NULL,
		CreatorCertificates		BLOB NOT NULL`, `INTEGER DEFAULT 0`),
			createTableWorkers: `
	CREATE TABLE IF NOT EXISTS Workers (
		MasterID					TEXT NOT NULL,
		WorkerID					TEXT NOT NULL,
		Confirmed					INTEGER NOT NULL,
		UNIQUE						(MasterID, WorkerID)
	)`,
			createTableBlacklists: `
	CREATE TABLE IF NOT EXISTS Blacklists (
		AdderID						TEXT NOT NULL,
		AddeeID						TEXT NOT NULL,
		UNIQUE						(AdderID, AddeeID)
	)`,
			createTableValidators: `
	CREATE TABLE IF NOT EXISTS Validators (
		Id							TEXT UNIQUE NOT NULL,
		Level						INTEGER NOT NULL,
		Name						TEXT NOT NULL DEFAULT '',
		Logo						TEXT NOT NULL DEFAULT '',
		URL							TEXT NOT NULL DEFAULT '',
		Description					TEXT NOT NULL DEFAULT '',
		Price						TEXT NOT NULL DEFAULT '0'
	)`,
			createTableCertificates: `
	CREATE TABLE IF NOT EXISTS Certificates (
	    Id						    TEXT NOT NULL, 
		OwnerID						TEXT NOT NULL,
		Attribute					INTEGER NOT NULL,
		AttributeLevel				INTEGER NOT NULL,
		Value						BLOB NOT NULL,
		ValidatorID					TEXT NOT NULL,
		FOREIGN KEY (ValidatorID)	REFERENCES Validators(Id) ON DELETE CASCADE
	)`,
			createTableProfiles: `
	CREATE TABLE IF NOT EXISTS Profiles (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		UserID						TEXT UNIQUE NOT NULL,
		IdentityLevel				INTEGER NOT NULL,
		Name						TEXT NOT NULL,
		Country						TEXT NOT NULL,
		IsCorporation				INTEGER NOT NULL,
		IsProfessional				INTEGER NOT NULL,
		Certificates				BLOB NOT NULL,
		ActiveAsks					INTEGER NOT NULL,
		ActiveBids					INTEGER NOT NULL
	)`,
			createTableStaleIDs: `
	CREATE TABLE IF NOT EXISTS StaleIDs (
		Id 							TEXT NOT NULL
	)`,
			createTableMisc: `
	CREATE TABLE IF NOT EXISTS Misc (
		Id							INTEGER PRIMARY KEY AUTOINCREMENT,
		LastKnownBlock				INTEGER NOT NULL
	)`,
			createIndexCmd: `CREATE INDEX IF NOT EXISTS %s_%s ON %s (%s)`,
			tablesInfo:     tInfo,
		},
		numBenchmarks: numBenchmarks,
		tablesInfo:    tInfo,
		builder: func() sq.StatementBuilderType {
			return sq.StatementBuilder.PlaceholderFormat(sq.Question)
		},
	}

	return storage
}
