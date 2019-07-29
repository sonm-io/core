package dwh

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
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
	opsTranslator = map[sonm.CmpOp]string{
		sonm.CmpOp_GTE: gte,
		sonm.CmpOp_LTE: lte,
		sonm.CmpOp_EQ:  eq,
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

func (m *sqlStorage) InsertDeal(conn queryConn, deal *sonm.Deal) error {
	ask, err := m.GetOrderByID(conn, deal.AskID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to getOrderDetails (Ask, `%s`): %v", deal.GetAskID().Unwrap().String(), err)
	}

	bid, err := m.GetOrderByID(conn, deal.BidID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to getOrderDetails (Ask, `%s`): %v", deal.GetBidID().Unwrap().String(), err)
	}

	var hasActiveChangeRequests bool
	if changeRequests, _ := m.GetDealChangeRequestsByDealID(conn, deal.Id.Unwrap(), true); len(changeRequests) > 0 {
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
		ask.GetOrder().GetTag(),
		bid.GetOrder().GetTag(),
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

func (m *sqlStorage) UpdateDeal(conn queryConn, deal *sonm.Deal) error {
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

func (m *sqlStorage) UpdateDealsSupplier(conn queryConn, profile *sonm.Profile) error {
	query, args, _ := m.builder().Update("Deals").SetMap(map[string]interface{}{
		"SupplierCertificates": []byte(profile.Certificates),
	}).Where("SupplierID = ?", profile.UserID.Unwrap().Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateDealsConsumer(conn queryConn, profile *sonm.Profile) error {
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

func (m *sqlStorage) GetDealByID(conn queryConn, dealID *big.Int) (*sonm.DWHDeal, error) {
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

func (m *sqlStorage) GetDeals(conn queryConn, r *sonm.DealsRequest) ([]*sonm.DWHDeal, uint64, error) {
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

	var deals []*sonm.DWHDeal
	for rows.Next() {
		deal, err := m.decodeDeal(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeDeal: %v", err)
		}

		deals = append(deals, deal)
	}

	return deals, count, nil
}

func (m *sqlStorage) GetDealConditions(conn queryConn, r *sonm.DealConditionsRequest) ([]*sonm.DealCondition, uint64, error) {
	builder := m.builder().Select("*").From("DealConditions")
	builder = builder.Where("DealID = ?", r.DealID.Unwrap().String())
	if len(r.Sortings) == 0 {
		builder = m.builderWithSortings(builder, []*sonm.SortingOption{{Field: "Id", Order: sonm.SortingOrder_Desc}})
	}
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*sonm.DealCondition
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

func (m *sqlStorage) InsertOrder(conn queryConn, order *sonm.DWHOrder) error {
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
		order.GetOrder().GetTag(),
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

func (m *sqlStorage) UpdateOrder(conn queryConn, order *sonm.Order) error {
	query, args, _ := m.builder().Update("Orders").SetMap(map[string]interface{}{
		"Status": order.OrderStatus,
		"DealID": order.DealID.String(),
	}).Where("Id = ?", order.Id.String()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateOrders(conn queryConn, profile *sonm.Profile) error {
	query, args, _ := m.builder().Update("Orders").SetMap(map[string]interface{}{
		"CreatorIdentityLevel": profile.IdentityLevel,
		"CreatorName":          profile.Name,
		"CreatorCountry":       profile.Country,
		"CreatorCertificates":  profile.Certificates,
	}).Where("AuthorId = ?", profile.UserID.Unwrap().Hex()).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetOrderByID(conn queryConn, orderID *big.Int) (*sonm.DWHOrder, error) {
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

func (m *sqlStorage) GetOrders(conn queryConn, r *sonm.OrdersRequest) ([]*sonm.DWHOrder, uint64, error) {
	builder := m.builder().Select("*").From("Orders AS o")
	if r.Status > 0 {
		builder = builder.Where("Status = ?", r.Status)
	}
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

	var orders []*sonm.DWHOrder
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

func (m *sqlStorage) GetMatchingOrders(conn queryConn, r *sonm.MatchingOrdersRequest) ([]*sonm.DWHOrder, uint64, error) {
	order, err := m.GetOrderByID(conn, r.Id.Unwrap())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to GetOrderByID: %v", err)
	}

	builder := m.builder().Select("*").From("Orders AS o")
	var (
		orderType    sonm.OrderType
		priceOp      string
		durationOp   string
		benchOp      string
		sortingOrder sonm.SortingOrder
	)
	if order.Order.OrderType == sonm.OrderType_BID {
		orderType = sonm.OrderType_ASK
		priceOp = lte
		durationOp = gte
		benchOp = gte
		sortingOrder = sonm.SortingOrder_Asc
	} else {
		orderType = sonm.OrderType_BID
		priceOp = gte
		durationOp = lte
		benchOp = lte
		sortingOrder = sonm.SortingOrder_Desc
	}
	builder = builder.Where("Type = ?", orderType).
		Where("Status = ?", sonm.OrderStatus_ORDER_ACTIVE).
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
	if order.Order.OrderType == sonm.OrderType_BID {
		builder = m.builderWithNetflagsFilter(builder, sonm.CmpOp_GTE, order.Order.GetNetflags().GetFlags())
	} else {
		builder = m.builderWithNetflagsFilter(builder, sonm.CmpOp_LTE, order.Order.GetNetflags().GetFlags())
	}
	builder = builder.Where("CreatorIdentityLevel >= ?", order.Order.IdentityLevel).
		Where("IdentityLevel <= ?", order.CreatorIdentityLevel)
	for benchID, benchValue := range order.Order.Benchmarks.Values {
		builder = builder.Where(fmt.Sprintf("%s %s ?", getBenchmarkColumn(uint64(benchID)), benchOp), benchValue)
	}
	builder = m.builderWithSortings(builder, []*sonm.SortingOption{{Field: "Price", Order: sortingOrder}})
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

	var orders []*sonm.DWHOrder
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

func (m *sqlStorage) GetProfiles(conn queryConn, r *sonm.ProfilesRequest) ([]*sonm.Profile, uint64, error) {
	builder := m.builder().Select("*").From("Profiles AS p")
	switch r.Role {
	case sonm.ProfileRole_Supplier:
		builder = builder.Where("ActiveAsks >= 1")
	case sonm.ProfileRole_Consumer:
		builder = builder.Where("ActiveBids >= 1")
	}
	builder = builder.Where("p.IdentityLevel >= ?", r.IdentityLevel)
	if len(r.Country) > 0 {
		builder = builder.Where(sq.Eq{"Country": r.Country})
	}
	if len(r.Identifier) > 0 {
		builder = builder.Where(sq.Or{
			sq.Expr("lower(Name) LIKE lower(?)", r.Identifier),
			sq.Expr("lower(UserID) LIKE lower(?)", r.Identifier),
		})
	}

	bQuery := r.BlacklistQuery
	if bQuery != nil && !bQuery.OwnerID.IsZero() {
		ownerBuilder := sq.Select("AddeeID").From("Blacklists").Where("AdderID = ?", bQuery.OwnerID.Unwrap().Hex()).
			Where("AddeeID = p.UserID")
		ownerQuery, _, _ := ownerBuilder.ToSql()
		if bQuery.OwnerID != nil {
			switch r.BlacklistQuery.Option {
			case sonm.BlacklistOption_WithoutMatching:
				builder = builder.Where(fmt.Sprintf("p.UserID NOT IN (%s)", ownerQuery))
			case sonm.BlacklistOption_OnlyMatching:
				builder = builder.Where(fmt.Sprintf("p.UserID IN (%s)", ownerQuery))
			}
		}
	}
	builder = m.builderWithSortings(builder, r.Sortings)
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()

	if bQuery != nil && bQuery.Option != sonm.BlacklistOption_IncludeAndMark && !bQuery.OwnerID.IsZero() {
		args = append(args, r.BlacklistQuery.OwnerID.Unwrap().Hex())
	}

	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var profiles []*sonm.Profile
	for rows.Next() {
		if profile, err := m.decodeProfile(rows); err != nil {
			return nil, 0, fmt.Errorf("failed to decodeProfile: %v", err)
		} else {
			profiles = append(profiles, profile)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	if r.BlacklistQuery != nil && r.BlacklistQuery.Option == sonm.BlacklistOption_IncludeAndMark {
		blacklistReply, err := m.GetBlacklist(conn, &sonm.BlacklistRequest{UserID: r.BlacklistQuery.OwnerID})
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get blacklist: %v", err)
		}

		var blacklistedAddrs = map[string]bool{}
		for _, blacklistedAddr := range blacklistReply.Addresses {
			blacklistedAddrs[blacklistedAddr] = true
		}
		for _, profile := range profiles {
			if blacklistedAddrs[profile.UserID.Unwrap().Hex()] {
				profile.IsBlacklisted = true
			}
		}
	}

	if err := m.addCertificatesToProfiles(conn, profiles); err != nil {
		return nil, 0, fmt.Errorf("failed to addCertificatesToProfiles: %v", err)
	}

	return profiles, count, nil
}

func (m *sqlStorage) addCertificatesToProfiles(conn queryConn, profiles []*sonm.Profile) error {
	var (
		userIDs          []common.Address
		userCertificates = map[common.Address][]*sonm.Certificate{}
	)
	for _, profile := range profiles {
		userIDs = append(userIDs, profile.UserID.Unwrap())
	}
	certs, err := m.GetCertificates(conn, userIDs...)
	if err != nil {
		return fmt.Errorf("failed to GetCertificates: %v", err)
	}

	for _, cert := range certs {
		userCertificates[cert.OwnerID.Unwrap()] = append(userCertificates[cert.OwnerID.Unwrap()], cert)
	}
	for _, profile := range profiles {
		certsEncoded, err := json.Marshal(userCertificates[profile.UserID.Unwrap()])
		if err != nil {
			return fmt.Errorf("failed to marshal %s certificates: %v", profile.UserID.Unwrap().Hex(), err)
		}

		profile.Certificates = string(certsEncoded)
	}

	return nil
}

func (m *sqlStorage) InsertDealChangeRequest(conn queryConn, changeRequest *sonm.DealChangeRequest) error {
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

func (m *sqlStorage) UpdateDealChangeRequest(conn queryConn, changeRequest *sonm.DealChangeRequest) error {
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

func (m *sqlStorage) GetDealChangeRequests(conn queryConn, changeRequest *sonm.DealChangeRequest) ([]*sonm.DealChangeRequest, error) {
	query, args, _ := m.builder().Select(m.tablesInfo.DealChangeRequestColumns...).
		From("DealChangeRequests").Where("DealID = ?", changeRequest.DealID.Unwrap().String()).
		Where("RequestType = ?", changeRequest.RequestType).
		Where("Status = ?", changeRequest.Status).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select DealChangeRequests: %v", err)
	}
	defer rows.Close()

	var out []*sonm.DealChangeRequest
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

func (m *sqlStorage) GetDealChangeRequestsByDealID(conn queryConn, dealID *big.Int, onlyActive bool) ([]*sonm.DealChangeRequest, error) {
	builder := m.builder().Select(m.tablesInfo.DealChangeRequestColumns...).
		From("DealChangeRequests").
		Where("DealID = ?", dealID.String())
	if onlyActive {
		builder = builder.Where("Status = ?", sonm.ChangeRequestStatus_REQUEST_CREATED)
	}
	query, args, _ := builder.OrderBy("CreatedTS").ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to select DealChangeRequestsByID: %v", err)
	}
	defer rows.Close()

	var out []*sonm.DealChangeRequest
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

func (m *sqlStorage) InsertDealCondition(conn queryConn, condition *sonm.DealCondition) error {
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

func (m *sqlStorage) InsertDealPayment(conn queryConn, billTS uint64, amount *big.Int, dealID *big.Int) error {
	query, args, err := m.builder().Insert("DealPayments").Values(
		billTS, util.BigIntToPaddedString(amount), dealID.String()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) InsertWorker(conn queryConn, masterID, workerID common.Address) error {
	query, args, err := m.builder().Insert("Workers").Values(masterID.Hex(), workerID.Hex(), false).
		Suffix("ON CONFLICT (MasterID, WorkerID) DO NOTHING").ToSql()
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
	query, args, err := m.builder().Insert("Blacklists").Values(adderID.Hex(), addeeID.Hex()).
		Suffix("ON CONFLICT (AdderID, AddeeID) DO NOTHING").ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) DeleteBlacklistEntry(conn queryConn, removerID, removeeID common.Address) error {
	query, args, err := m.builder().Delete("Blacklists").Where("AdderID = ?", removerID.Hex()).
		Where("AddeeID = ?", removeeID.Hex()).ToSql()
	_, err = conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetBlacklist(conn queryConn, r *sonm.BlacklistRequest) (*sonm.BlacklistReply, error) {
	builder := m.builder().Select("*").From("Blacklists")
	if !r.UserID.IsZero() {
		builder = builder.Where("AdderID = ?", r.UserID.Unwrap().Hex())
	}
	builder = m.builderWithSortings(builder, []*sonm.SortingOption{})
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

	return &sonm.BlacklistReply{
		OwnerID:   r.UserID,
		Addresses: addees,
		Count:     count,
	}, nil
}

func (m *sqlStorage) GetBlacklistsContainingUser(conn queryConn, r *sonm.BlacklistRequest) (*sonm.BlacklistsContainingUserReply, error) {
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

	var adders []*sonm.EthAddress
	for rows.Next() {
		var adderID string
		if err := rows.Scan(&adderID); err != nil {
			return nil, fmt.Errorf("failed to scan BlacklistAddress row: %v", err)
		}

		ethAddress, err := util.HexToAddress(adderID)
		if err != nil {
			return nil, fmt.Errorf("failed to use `%s` as EthAddress", adderID)
		}
		adders = append(adders, sonm.NewEthAddress(ethAddress))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return &sonm.BlacklistsContainingUserReply{
		Blacklists: adders,
		Count:      count,
	}, nil
}

func (m *sqlStorage) InsertOrUpdateValidator(conn queryConn, validator *sonm.Validator) error {
	query, args, _ := m.builder().Insert("Validators").Columns("Id", "Level").Values(validator.Id.Unwrap().Hex(), validator.Level).
		Suffix("ON CONFLICT (Id) DO UPDATE SET Level = ?", validator.Level).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetValidator(conn queryConn, validatorID common.Address) (*sonm.DWHValidator, error) {
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
			value = sonm.NewBigInt(big.NewInt(0).SetBytes(bytes)).PaddedString()
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

func (m *sqlStorage) InsertCertificate(conn queryConn, certificate *sonm.Certificate) error {
	query, args, _ := m.builder().Insert("Certificates").Values(
		certificate.GetId().Unwrap().String(),
		certificate.OwnerID.Unwrap().Hex(),
		certificate.Attribute,
		certificate.IdentityLevel,
		certificate.Value,
		certificate.ValidatorID.Unwrap().Hex(),
	).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetCertificates(conn queryConn, ownerIDs ...common.Address) ([]*sonm.Certificate, error) {
	var ids []string
	for _, id := range ownerIDs {
		ids = append(ids, id.Hex())
	}
	query, args, _ := m.builder().Select("*").From("Certificates").Where(sq.Eq{"OwnerID": ids}).
		OrderBy("AttributeLevel DESC").ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to getCertificatesByUserID: %v", err)
	}
	defer rows.Close()

	var (
		certificates     []*sonm.Certificate
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

func (m *sqlStorage) GetCertificate(conn queryConn, id *big.Int) (*sonm.Certificate, error) {
	query, args, _ := m.builder().Select("*").From("Certificates").Where("Id = ?", id.String()).ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	if cert, err := m.decodeCertificate(rows); err != nil {
		return nil, fmt.Errorf("failed to decodeCertificate: %v", err)
	} else {
		return cert, nil
	}
}

func (m *sqlStorage) DeleteCertificate(conn queryConn, id *big.Int) error {
	query, args, _ := m.builder().Delete("Certificates").Where("Id = ?", id.String()).ToSql()
	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to run query: %v", err)
	}

	return nil
}

func (m *sqlStorage) InsertProfileUserID(conn queryConn, profile *sonm.Profile) error {
	query, args, _ := m.builder().Insert("Profiles").Columns(m.tablesInfo.ProfileColumns[1:]...).Values(
		profile.UserID.Unwrap().Hex(),
		0, "", "", false, false,
		profile.ActiveAsks,
		profile.ActiveBids,
	).Suffix("ON CONFLICT (UserID) DO NOTHING").ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetProfileByID(conn queryConn, userID common.Address) (*sonm.Profile, error) {
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

func (m *sqlStorage) GetValidators(conn queryConn, r *sonm.ValidatorsRequest) ([]*sonm.DWHValidator, uint64, error) {
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

	var out []*sonm.DWHValidator
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

func (m *sqlStorage) GetWorkers(conn queryConn, r *sonm.WorkersRequest) ([]*sonm.DWHWorker, uint64, error) {
	builder := m.builder().Select("*").From("Workers")
	if !r.MasterID.IsZero() {
		builder = builder.Where("MasterID = ?", r.MasterID.Unwrap().String())
	}
	builder = m.builderWithSortings(builder, []*sonm.SortingOption{
		{Field: "Confirmed", Order: sonm.SortingOrder_Desc},
		{Field: "WorkerID", Order: sonm.SortingOrder_Asc},
	})
	query, args, _ := m.builderWithOffsetLimit(builder, r.Limit, r.Offset).ToSql()
	rows, count, err := m.runQuery(conn, "*", r.WithCount, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var out []*sonm.DWHWorker
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

func (m *sqlStorage) InsertLastEvent(conn queryConn, event *blockchain.Event) error {
	query, args, _ := m.builder().Insert("Misc").Columns("BlockNumber", "TxIndex", "ReceiptIndex").
		Values(event.BlockNumber, event.TxIndex, event.ReceiptIndex).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) UpdateLastEvent(conn queryConn, event *blockchain.Event) error {
	query, args, _ := m.builder().Update("Misc").SetMap(map[string]interface{}{
		"BlockNumber":  event.BlockNumber,
		"TxIndex":      event.TxIndex,
		"ReceiptIndex": event.ReceiptIndex,
	}).ToSql()
	_, err := conn.Exec(query, args...)
	return err
}

func (m *sqlStorage) GetLastEvent(conn queryConn) (*blockchain.Event, error) {
	query, _, _ := m.builder().Select("*").From("Misc").Limit(1).ToSql()
	rows, err := conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to GetLastEvent: %v", err)
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.New("GetLastEvent: no entries")
	}

	var (
		blockNumber  uint64
		txIndex      uint64
		receiptIndex uint64
	)
	if err := rows.Scan(&blockNumber, &txIndex, &receiptIndex); err != nil {
		return nil, fmt.Errorf("failed to parse last event: %v", err)
	}

	return &blockchain.Event{
		BlockNumber:  blockNumber,
		TxIndex:      txIndex,
		ReceiptIndex: receiptIndex,
	}, nil
}

func (m *sqlStorage) getStats(conn queryConn) (*sonm.DWHStatsReply, error) {
	var (
		numCurrDealsQ, argsCurrDeals, _  = sq.Select("count(Id)").From("Deals").Where("Status=?", sonm.DealStatus_DEAL_ACCEPTED).Prefix("(").Suffix(")").ToSql()
		numDealsQ, _, _                  = sq.Select("count(Id)").From("Deals").Prefix("(").Suffix(")").ToSql()
		dealsDurQ, argsDealsDur, _       = sq.Select("case count(id) when 0 then 0 else sum(EndTime - StartTime)/3600 end").From("Deals").Where("Status=?", sonm.DealStatus_DEAL_CLOSED).Prefix("(").Suffix(")").ToSql()
		dealsAvgDurQ, argsDealsAvgDur, _ = sq.Select("case count(id) when 0 then 0 else sum(EndTime - StartTime)/3600/(count(id)+1) end").From("Deals").Where("Status=?", sonm.DealStatus_DEAL_CLOSED).Prefix("(").Suffix(")").ToSql()
		numWorkersQ, _, _                = sq.Select("count(distinct WorkerID)").From("Workers").Prefix("(").Suffix(")").ToSql()
		numMastersQ, _, _                = sq.Select("count(distinct MasterID)").From("Workers").Prefix("(").Suffix(")").ToSql()
		numCustomersQ, argsCustomers, _  = sq.Select("count(distinct ConsumerID)").From("Deals").Where("Status=?", sonm.DealStatus_DEAL_CLOSED).Prefix("(").Suffix(")").ToSql()
	)
	var args []interface{}
	args = append(args, argsCurrDeals...)
	args = append(args, argsDealsDur...)
	args = append(args, argsDealsAvgDur...)
	args = append(args, argsCustomers...)
	query, _, _ := m.builder().
		Select(numCurrDealsQ, numDealsQ, dealsDurQ, dealsAvgDurQ, numWorkersQ, numMastersQ, numCustomersQ).
		ToSql()
	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to getStats: %v", err)
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return nil, errors.New("getStats: no entries")
	}

	var (
		numCurrDeals uint64
		numDeals     uint64
		dealsDur     uint64
		dealsAvgDur  uint64
		numWorkers   uint64
		numMasters   uint64
		numCustomers uint64
	)
	if err := rows.Scan(
		&numCurrDeals,
		&numDeals,
		&dealsDur,
		&dealsAvgDur,
		&numWorkers,
		&numMasters,
		&numCustomers,
	); err != nil {
		return nil, fmt.Errorf("failed to parse stats: %v", err)
	}

	return &sonm.DWHStatsReply{
		CurrentDeals:        numCurrDeals,
		TotalDeals:          numDeals,
		TotalDealsDuration:  dealsDur,
		AverageDealDuration: dealsAvgDur,
		Workers:             numWorkers,
		Masters:             numMasters,
		Customers:           numCustomers,
	}, nil
}

func (m *sqlStorage) builderWithBenchmarkFilters(builder sq.SelectBuilder, benches map[uint64]*sonm.MaxMinUint64) sq.SelectBuilder {
	for benchID, condition := range benches {
		if condition.Max == condition.Min {
			builder = builder.Where(fmt.Sprintf("%s = ?", getBenchmarkColumn(benchID)), condition.Max)
			continue
		}
		if condition.Max > 0 {
			builder = builder.Where(fmt.Sprintf("%s <= ?", getBenchmarkColumn(benchID)), condition.Max)
		}
		if condition.Min > 0 {
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

func (m *sqlStorage) builderWithSortings(builder sq.SelectBuilder, sortings []*sonm.SortingOption) sq.SelectBuilder {
	var sortsFlat []string
	for _, sort := range sortings {
		sortsFlat = append(sortsFlat, fmt.Sprintf("%s %s", sort.Field, sonm.SortingOrder_name[int32(sort.Order)]))
	}
	builder = builder.OrderBy(sortsFlat...)

	return builder
}

func (m *sqlStorage) builderWithNetflagsFilter(builder sq.SelectBuilder, operator sonm.CmpOp, value uint64) sq.SelectBuilder {
	switch operator {
	case sonm.CmpOp_GTE:
		return builder.Where("Netflags | ~ CAST (? as int) = -1", value)
	case sonm.CmpOp_LTE:
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

func (m *sqlStorage) decodeDeal(rows *sql.Rows) (*sonm.DWHDeal, error) {
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
		askTag               = &[]byte{}
		bidTag               = &[]byte{}
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
		askTag,
		bidTag,
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

	bigID, err := sonm.NewBigIntFromString(*id)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}

	bigAskID, err := sonm.NewBigIntFromString(*askID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (askID): %v", err)
	}

	bigBidID, err := sonm.NewBigIntFromString(*bidID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (bidID): %v", err)
	}

	return &sonm.DWHDeal{
		Deal: &sonm.Deal{
			Id:             bigID,
			SupplierID:     sonm.NewEthAddress(common.HexToAddress(*supplierID)),
			ConsumerID:     sonm.NewEthAddress(common.HexToAddress(*consumerID)),
			MasterID:       sonm.NewEthAddress(common.HexToAddress(*masterID)),
			AskID:          bigAskID,
			BidID:          bigBidID,
			Price:          sonm.NewBigInt(bigPrice),
			Duration:       *duration,
			StartTime:      &sonm.Timestamp{Seconds: *startTime},
			EndTime:        &sonm.Timestamp{Seconds: *endTime},
			Status:         sonm.DealStatus(*dealStatus),
			BlockedBalance: sonm.NewBigInt(bigBlockedBalance),
			TotalPayout:    sonm.NewBigInt(bigTotalPayout),
			LastBillTS:     &sonm.Timestamp{Seconds: *lastBillTS},
			Benchmarks:     &sonm.Benchmarks{Values: benchmarksUint64},
		},
		Netflags:             *netflags,
		AskIdentityLevel:     *askIdentityLevel,
		BidIdentityLevel:     *bidIdentityLevel,
		SupplierCertificates: *supplierCertificates,
		ConsumerCertificates: *consumerCertificates,
		ActiveChangeRequest:  *activeChangeRequest,
		AskTag:               *askTag,
		BidTag:               *bidTag,
	}, nil
}

func (m *sqlStorage) decodeDealCondition(rows *sql.Rows) (*sonm.DealCondition, error) {
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
	bigDealID, err := sonm.NewBigIntFromString(dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (DealID): %v", err)
	}

	return &sonm.DealCondition{
		Id:          id,
		SupplierID:  sonm.NewEthAddress(common.HexToAddress(supplierID)),
		ConsumerID:  sonm.NewEthAddress(common.HexToAddress(consumerID)),
		MasterID:    sonm.NewEthAddress(common.HexToAddress(masterID)),
		Price:       sonm.NewBigInt(bigPrice),
		Duration:    duration,
		StartTime:   &sonm.Timestamp{Seconds: startTime},
		EndTime:     &sonm.Timestamp{Seconds: endTime},
		TotalPayout: sonm.NewBigInt(bigTotalPayout),
		DealID:      bigDealID,
	}, nil
}

func (m *sqlStorage) decodeOrder(rows *sql.Rows) (*sonm.DWHOrder, error) {
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
	bigPrice, err := sonm.NewBigIntFromString(*price)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (Price): %v", err)
	}
	bigFrozenSum, err := sonm.NewBigIntFromString(*frozenSum)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (FrozenSum): %v", err)
	}
	bigID, err := sonm.NewBigIntFromString(*id)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}
	bigDealID, err := sonm.NewBigIntFromString(*dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (DealID): %v", err)
	}

	return &sonm.DWHOrder{
		Order: &sonm.Order{
			Id:             bigID,
			DealID:         bigDealID,
			OrderType:      sonm.OrderType(*orderType),
			OrderStatus:    sonm.OrderStatus(*orderStatus),
			AuthorID:       sonm.NewEthAddress(common.HexToAddress(*author)),
			CounterpartyID: sonm.NewEthAddress(common.HexToAddress(*counterAgent)),
			Duration:       *duration,
			Price:          bigPrice,
			Netflags:       &sonm.NetFlags{Flags: *netflags},
			IdentityLevel:  sonm.IdentityLevel(*identityLevel),
			Blacklist:      *blacklist,
			Tag:            *tag,
			FrozenSum:      bigFrozenSum,
			Benchmarks:     &sonm.Benchmarks{Values: benchmarksUint64},
		},
		CreatedTS:            &sonm.Timestamp{Seconds: int64(*createdTS)},
		CreatorIdentityLevel: *creatorIdentityLevel,
		CreatorName:          *creatorName,
		CreatorCountry:       *creatorCountry,
		CreatorCertificates:  *creatorCertificates,
		MasterID:             sonm.NewEthAddress(common.HexToAddress(*masterID)),
	}, nil
}

func (m *sqlStorage) decodeDealChangeRequest(rows *sql.Rows) (*sonm.DealChangeRequest, error) {
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
	bigDealID, err := sonm.NewBigIntFromString(dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ID): %v", err)
	}

	bigChangeRequestID, err := sonm.NewBigIntFromString(changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("failed to NewBigIntFromString (ChangeRequestID): %v", err)
	}

	return &sonm.DealChangeRequest{
		Id:          bigChangeRequestID,
		DealID:      bigDealID,
		RequestType: sonm.OrderType(requestType),
		Duration:    duration,
		Price:       sonm.NewBigInt(bigPrice),
		Status:      sonm.ChangeRequestStatus(changeRequestStatus),
	}, nil
}

func (m *sqlStorage) decodeCertificate(rows *sql.Rows) (*sonm.Certificate, error) {
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
		bigID, err := sonm.NewBigIntFromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate id: %v", err)
		}

		return &sonm.Certificate{
			Id:            bigID,
			OwnerID:       sonm.NewEthAddress(common.HexToAddress(ownerID)),
			Attribute:     attribute,
			IdentityLevel: identityLevel,
			Value:         value,
			ValidatorID:   sonm.NewEthAddress(common.HexToAddress(validatorID)),
		}, nil
	}
}

func (m *sqlStorage) decodeProfile(rows *sql.Rows) (*sonm.Profile, error) {
	var (
		id             uint64
		userID         string
		identityLevel  uint64
		name           string
		country        string
		isCorporation  bool
		isProfessional bool
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
		&activeAsks,
		&activeBids,
	); err != nil {
		return nil, fmt.Errorf("failed to scan Profile row: %v", err)
	}

	return &sonm.Profile{
		UserID:         sonm.NewEthAddress(common.HexToAddress(userID)),
		IdentityLevel:  identityLevel,
		Name:           name,
		Country:        country,
		IsCorporation:  isCorporation,
		IsProfessional: isProfessional,
		ActiveAsks:     activeAsks,
		ActiveBids:     activeBids,
	}, nil
}

func (m *sqlStorage) decodeValidator(rows *sql.Rows) (*sonm.DWHValidator, error) {
	var (
		validatorID string
		level       uint64
		name        string
		kycIcon     string
		kycURL      string
		description string
		kycPrice    string
	)
	if err := rows.Scan(&validatorID, &level, &name, &kycIcon, &kycURL, &description, &kycPrice); err != nil {
		return nil, fmt.Errorf("failed to scan Validator row: %v", err)
	}

	bigPrice, err := sonm.NewBigIntFromString(kycPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to use price as big int: %s", kycPrice)
	}
	return &sonm.DWHValidator{
		Validator: &sonm.Validator{
			Id:    sonm.NewEthAddress(common.HexToAddress(validatorID)),
			Level: level,
		},
		Name:        name,
		Icon:        kycIcon,
		Url:         kycURL,
		Description: description,
		Price:       bigPrice,
	}, nil
}

func (m *sqlStorage) decodeWorker(rows *sql.Rows) (*sonm.DWHWorker, error) {
	var (
		masterID  string
		slaveID   string
		confirmed bool
	)
	if err := rows.Scan(&masterID, &slaveID, &confirmed); err != nil {
		return nil, fmt.Errorf("failed to scan Worker row: %v", err)
	}

	return &sonm.DWHWorker{
		MasterID:  sonm.NewEthAddress(common.HexToAddress(masterID)),
		SlaveID:   sonm.NewEthAddress(common.HexToAddress(slaveID)),
		Confirmed: confirmed,
	}, nil
}

func (m *sqlStorage) filterSortings(sortings []*sonm.SortingOption, columns map[string]bool) (out []*sonm.SortingOption) {
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
		"AskTag",
		"BidTag",
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
		"ActiveAsks",
		"ActiveBids",
	}
	validatorColumns := []string{
		"Id",
		"Level",
		"Name",
		"KYC_icon",
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
		ActiveChangeRequest     BOOLEAN NOT NULL,
		AskTag					BYTEA,
		BidTag					BYTEA`, `BIGINT DEFAULT 0`),
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
		KYC_icon					TEXT NOT NULL DEFAULT '',
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
		ActiveAsks					INTEGER NOT NULL,
		ActiveBids					INTEGER NOT NULL
	)`,
			createTableMisc: `
	CREATE TABLE IF NOT EXISTS Misc (
		BlockNumber 				INTEGER NOT NULL,
		TxIndex						INTEGER NOT NULL,
		ReceiptIndex				INTEGER NOT NULL
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

func setupDB(ctx context.Context, db *sql.DB, blockchain blockchain.API) (*sqlStorage, error) {
	numBenchmarks, err := blockchain.Market().GetNumBenchmarks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to GetNumBenchmarks: %v", err)
	}
	if numBenchmarks >= NumMaxBenchmarks {
		return nil, errors.New("market number of benchmarks is greater than NumMaxBenchmarks")
	}

	var storage = newPostgresStorage(numBenchmarks)
	if err := storage.Setup(db); err != nil {
		return nil, fmt.Errorf("failed to setup storage: %v", err)
	}

	return storage, nil
}

func (m *sqlStorage) GetOrdersByIDs(conn queryConn, r *sonm.OrdersByIDsRequest) ([]*sonm.DWHOrder, uint64, error) {
	var ids = make([]string, len(r.Ids))
	for idx, id := range r.Ids {
		ids[idx] = id.Unwrap().String()
	}

	if len(ids) < 1 {
		return nil, 0, errors.New("no IDs provided")
	}

	builder := m.builder().Select("*").From("Orders AS o").Where(sq.Eq{"ID": ids})

	query, args, _ := builder.ToSql()
	rows, count, err := m.runQuery(conn, strings.Join(m.tablesInfo.OrderColumns, ", "), true, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to run query: %v", err)
	}
	defer rows.Close()

	var orders = make([]*sonm.DWHOrder, len(r.Ids))
	for idx := 0; rows.Next(); idx++ {
		order, err := m.decodeOrder(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to decodeOrder: %v", err)
		}
		orders[idx] = order
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %v", err)
	}

	return orders, count, nil
}
