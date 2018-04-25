package dwh

import (
	"crypto/ecdsa"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type DWH struct {
	mu          sync.RWMutex
	ctx         context.Context
	cfg         *Config
	cancel      context.CancelFunc
	grpc        *grpc.Server
	http        *rest.Server
	logger      *zap.Logger
	db          *sql.DB
	creds       credentials.TransportCredentials
	certRotator util.HitlessCertRotator
	blockchain  blockchain.API
	commands    map[string]string
}

func NewDWH(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (w *DWH, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	w = &DWH{
		ctx:    ctx,
		cfg:    cfg,
		logger: log.GetLogger(ctx),
	}

	setupDB, ok := setupDBCallbacks[cfg.Storage.Backend]
	if !ok {
		return nil, errors.Errorf("unsupported backend: %s", cfg.Storage.Backend)
	}

	if err = setupDB(w); err != nil {
		return nil, errors.Wrap(err, "failed to setupDB")
	}

	var TLSConfig *tls.Config
	w.certRotator, TLSConfig, err = util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	w.creds = util.NewTLS(TLSConfig)
	server := xgrpc.NewServer(w.logger, xgrpc.Credentials(w.creds), xgrpc.DefaultTraceInterceptor())
	w.grpc = server
	pb.RegisterDWHServer(w.grpc, w)
	grpc_prometheus.Register(w.grpc)

	return
}

func (w *DWH) Serve() error {
	if w.cfg.Blockchain != nil {
		bch, err := blockchain.NewAPI(blockchain.WithEthEndpoint(w.cfg.Blockchain.EthEndpoint))
		if err != nil {
			return errors.Wrap(err, "failed to create NewAPI")
		}
		w.blockchain = bch
		go w.monitorBlockchain()
	} else {
		w.logger.Info("monitoring disabled")
	}

	wg := errgroup.Group{}
	wg.Go(w.serveGRPC)
	wg.Go(w.serveHTTP)

	return wg.Wait()
}

func (w *DWH) Close() {
	w.grpc.Stop()
	w.http.Close()
}

func (w *DWH) serveGRPC() error {
	lis, err := net.Listen("tcp", w.cfg.GRPCListenAddr)
	if err != nil {
		return errors.Wrapf(err, "failed to listen on %s", w.cfg.GRPCListenAddr)
	}

	return w.grpc.Serve(lis)
}

func (w *DWH) serveHTTP() error {
	options := []rest.Option{rest.WithContext(w.ctx)}

	lis, err := net.Listen("tcp", w.cfg.HTTPListenAddr)
	if err != nil {
		log.S(w.ctx).Info("failed to create http listener")
		return err
	}

	options = append(options, rest.WithListener(lis))

	srv, err := rest.NewServer(options...)
	if err != nil {
		return errors.Wrap(err, "failed to create rest server")
	}

	err = srv.RegisterService((*pb.DWHServer)(nil), w)
	if err != nil {
		return errors.Wrap(err, "failed to RegisterService")
	}

	w.http = srv

	return srv.Serve()
}

func (w *DWH) GetDeals(ctx context.Context, request *pb.DealsRequest) (*pb.DWHDealsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDeals(ctx, request)
}

func (w *DWH) getDeals(ctx context.Context, request *pb.DealsRequest) (*pb.DWHDealsReply, error) {
	var filters []*filter
	if request.Status > 0 {
		filters = append(filters, newFilter("Status", eq, request.Status, "AND"))
	}
	if request.SupplierID > "0" {
		filters = append(filters, newFilter("SupplierID", eq, request.SupplierID, "AND"))
	}
	if request.ConsumerID > "0" {
		filters = append(filters, newFilter("ConsumerID", eq, request.ConsumerID, "AND"))
	}
	if request.MasterID > "0" {
		filters = append(filters, newFilter("MasterID", eq, request.MasterID, "AND"))
	}
	if request.AskID > "0" {
		filters = append(filters, newFilter("AskID", eq, request.AskID, "AND"))
	}
	if request.BidID > "0" {
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
		w.addBenchmarksConditions(request.Benchmarks, &filters)
	}

	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Deals",
		filters:  filters,
		sortings: filterSortings(request.Sortings, DealsColumns),
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDeals")
	}
	defer rows.Close()

	var deals []*pb.DWHDeal
	for rows.Next() {
		deal, err := w.decodeDeal(rows)
		if err != nil {
			w.logger.Error("failed to decodeDeal", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetDeals")
		}

		deals = append(deals, deal)
	}

	return &pb.DWHDealsReply{Deals: deals}, nil
}

func (w *DWH) GetDealDetails(ctx context.Context, request *pb.ID) (*pb.DWHDeal, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealDetails(ctx, request)
}

func (w *DWH) getDealDetails(ctx context.Context, request *pb.ID) (*pb.DWHDeal, error) {
	rows, err := w.db.Query(w.commands["selectDealByID"], request.Id)
	if err != nil {
		w.logger.Error("failed to selectDealByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDealDetails")
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		w.logger.Error("deal not found", zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDealDetails")
	}

	return w.decodeDeal(rows)
}

func (w *DWH) GetDealConditions(ctx context.Context, request *pb.DealConditionsRequest) (*pb.DealConditionsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealConditions(ctx, request)
}

func (w *DWH) getDealConditions(ctx context.Context, request *pb.DealConditionsRequest) (*pb.DealConditionsReply, error) {
	var filters []*filter
	if len(request.Sortings) < 1 {
		request.Sortings = []*pb.SortingOption{{Field: "rowid", Order: pb.SortingOrder_Desc}}
	}

	filters = append(filters, newFilter("DealID", eq, request.DealID, "AND"))
	rows, _, err := runQuery(w.db, &queryOpts{
		table:     "DealConditions",
		filters:   filters,
		sortings:  request.Sortings,
		offset:    request.Offset,
		limit:     request.Limit,
		withRowid: true,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDealConditions")
	}
	defer rows.Close()

	var out []*pb.DealCondition
	for rows.Next() {
		dealCondition, err := w.decodeDealCondition(rows)
		if err != nil {
			w.logger.Error("failed to decodeDealCondition", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetDealConditions")
		}
		out = append(out, dealCondition)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetDealConditions")
	}

	return &pb.DealConditionsReply{Conditions: out}, nil
}

func (w *DWH) GetOrders(ctx context.Context, request *pb.OrdersRequest) (*pb.DWHOrdersReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getOrders(ctx, request)
}

func (w *DWH) getOrders(ctx context.Context, request *pb.OrdersRequest) (*pb.DWHOrdersReply, error) {
	var filters []*filter
	filters = append(filters, newFilter("Status", eq, pb.OrderStatus_ORDER_ACTIVE, "AND"))
	if request.DealID > "0" {
		filters = append(filters, newFilter("DealID", eq, request.DealID, "AND"))
	}
	if request.Type > 0 {
		filters = append(filters, newFilter("Type", eq, request.Type, "AND"))
	}
	if request.AuthorID > "0" {
		filters = append(filters, newFilter("AuthorID", eq, request.AuthorID, "AND"))
	}
	if request.CounterpartyID > "0" {
		filters = append(filters, newFilter("CounterpartyID", eq, request.CounterpartyID, "AND"))
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
	if request.Benchmarks != nil {
		w.addBenchmarksConditions(request.Benchmarks, &filters)
	}
	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Orders",
		filters:  filters,
		sortings: filterSortings(request.Sortings, OrdersColumns),
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetOrders")
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := w.decodeOrder(rows)
		if err != nil {
			w.logger.Error("failed to decodeOrder", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetOrders")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders}, nil
}

func (w *DWH) GetMatchingOrders(ctx context.Context, request *pb.MatchingOrdersRequest) (*pb.DWHOrdersReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getMatchingOrders(ctx, request)
}

func (w *DWH) getMatchingOrders(ctx context.Context, request *pb.MatchingOrdersRequest) (*pb.DWHOrdersReply, error) {
	order, err := w.getOrderDetails(ctx, request.Id)
	if err != nil {
		w.logger.Error("failed to getOrderDetails", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetMatchingOrders (no matching order)")
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
	if order.Order.CounterpartyID > "0" {
		filters = append(filters, newFilter("AuthorID", eq, order.Order.CounterpartyID, "AND"))
	}
	counterpartyFilter := newFilter("CounterpartyID", eq, "", "OR")
	counterpartyFilter.OpenBracket = true
	filters = append(filters, counterpartyFilter)
	counterpartyFilter = newFilter("CounterpartyID", eq, order.Order.AuthorID, "AND")
	counterpartyFilter.CloseBracket = true
	filters = append(filters, counterpartyFilter)
	if order.Order.OrderType == pb.OrderType_BID {
		filters = append(filters, newNetflagsFilter(pb.CmpOp_GTE, order.Order.Netflags))
	} else {
		filters = append(filters, newNetflagsFilter(pb.CmpOp_LTE, order.Order.Netflags))
	}
	filters = append(filters, newFilter("IdentityLevel", gte, order.Order.IdentityLevel, "AND"))
	filters = append(filters, newFilter("CreatorIdentityLevel", lte, order.CreatorIdentityLevel, "AND"))
	filters = append(filters, newFilter("CPUSysbenchMulti", benchOp, order.Order.Benchmarks.CPUSysbenchMulti(), "AND"))
	filters = append(filters, newFilter("CPUSysbenchOne", benchOp, order.Order.Benchmarks.CPUSysbenchOne(), "AND"))
	filters = append(filters, newFilter("CPUCores", benchOp, order.Order.Benchmarks.CPUCores(), "AND"))
	filters = append(filters, newFilter("RAMSize", benchOp, order.Order.Benchmarks.RAMSize(), "AND"))
	filters = append(filters, newFilter("StorageSize", benchOp, order.Order.Benchmarks.StorageSize(), "AND"))
	filters = append(filters, newFilter("NetTrafficIn", benchOp, order.Order.Benchmarks.NetTrafficIn(), "AND"))
	filters = append(filters, newFilter("NetTrafficOut", benchOp, order.Order.Benchmarks.NetTrafficOut(), "AND"))
	filters = append(filters, newFilter("GPUCount", benchOp, order.Order.Benchmarks.GPUCount(), "AND"))
	filters = append(filters, newFilter("GPUMem", benchOp, order.Order.Benchmarks.GPUMem(), "AND"))
	filters = append(filters, newFilter("GPUEthHashrate", benchOp, order.Order.Benchmarks.GPUEthHashrate(), "AND"))
	filters = append(filters, newFilter("GPUCashHashrate", benchOp, order.Order.Benchmarks.GPUCashHashrate(), "AND"))
	filters = append(filters, newFilter("GPURedshift", benchOp, order.Order.Benchmarks.GPURedshift(), "AND"))

	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Orders",
		filters:  filters,
		sortings: []*pb.SortingOption{{Field: "Price", Order: sortingOrder}},
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetMatchingOrders")
	}
	defer rows.Close()

	var orders []*pb.DWHOrder
	for rows.Next() {
		order, err := w.decodeOrder(rows)
		if err != nil {
			w.logger.Error("failed to decodeOrder", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetMatchingOrders")
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetMatchingOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders}, nil
}

func (w *DWH) GetOrderDetails(ctx context.Context, request *pb.ID) (*pb.DWHOrder, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getOrderDetails(ctx, request)
}

func (w *DWH) getOrderDetails(ctx context.Context, request *pb.ID) (*pb.DWHOrder, error) {
	rows, err := w.db.Query(w.commands["selectOrderByID"], request.Id)
	if err != nil {
		w.logger.Error("failed to selectOrderByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetOrderDetails")
	}
	defer rows.Close()

	if !rows.Next() {
		w.logger.Error("order not found", zap.Error(rows.Err()), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetOrderDetails")
	}

	return w.decodeOrder(rows)
}

func (w *DWH) GetProfiles(ctx context.Context, request *pb.ProfilesRequest) (*pb.ProfilesReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getProfiles(ctx, request)
}

func (w *DWH) getProfiles(ctx context.Context, request *pb.ProfilesRequest) (*pb.ProfilesReply, error) {
	var filters []*filter
	switch request.Role {
	case pb.ProfileRole_Supplier:
		filters = append(filters, newFilter("ActiveAsks", gte, 0, "AND"))
	case pb.ProfileRole_Consumer:
		filters = append(filters, newFilter("ActiveBids", gte, 0, "AND"))
	}
	filters = append(filters, newFilter("IdentityLevel", gte, request.IdentityLevel, "AND"))
	if len(request.Country) > 0 {
		filters = append(filters, newFilter("Country", eq, request.Country, "AND"))
	}
	if len(request.Name) > 0 {
		filters = append(filters, newFilter("Name", eq, request.Name, "AND"))
	}

	opts := &queryOpts{
		table:    "Profiles",
		filters:  filters,
		sortings: filterSortings(request.Sortings, ProfilesColumns),
		offset:   request.Offset,
		limit:    request.Limit,
	}
	if request.BlacklistQuery != nil {
		opts.selectAs = "AS p"
		switch request.BlacklistQuery.Option {
		case pb.BlacklistOption_WithoutMatching:
			opts.customFilter = &customFilter{
				clause: w.commands["profileNotInBlacklist"],
				values: []interface{}{request.BlacklistQuery.OwnerID},
			}
		case pb.BlacklistOption_OnlyMatching:
			opts.customFilter = &customFilter{
				clause: w.commands["profileInBlacklist"],
				values: []interface{}{request.BlacklistQuery.OwnerID},
			}
		}
	}

	rows, _, err := runQuery(w.db, opts)
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfiles")
	}

	var out []*pb.Profile
	for rows.Next() {
		if profile, err := w.decodeProfile(rows); err != nil {
			w.logger.Error("failed to decodeProfile", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetProfiles")
		} else {
			out = append(out, profile)
		}
	}

	if request.BlacklistQuery != nil && request.BlacklistQuery.Option == pb.BlacklistOption_IncludeAndMark {
		blacklistReply, err := w.getBlacklist(w.ctx, &pb.BlacklistRequest{OwnerID: request.BlacklistQuery.OwnerID})
		if err != nil {
			w.logger.Error("failed to GetBlacklist", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetProfiles")
		}

		var blacklistedAddrs = map[string]bool{}
		for _, blacklistedAddr := range blacklistReply.Addresses {
			blacklistedAddrs[blacklistedAddr] = true
		}

		for _, profile := range out {
			if blacklistedAddrs[profile.UserID] {
				profile.IsBlacklisted = true
			}
		}
	}

	return &pb.ProfilesReply{Profiles: out}, nil
}

func (w *DWH) GetProfileInfo(ctx context.Context, request *pb.ID) (*pb.Profile, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getProfileInfo(ctx, request, true)
}

func (w *DWH) getProfileInfo(ctx context.Context, request *pb.ID, logErrors bool) (*pb.Profile, error) {
	rows, err := w.db.Query(w.commands["selectProfileByID"], request.Id)
	if err != nil {
		w.logger.Error("failed to selectProfileByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfileInfo")
	}
	defer rows.Close()

	if !rows.Next() {
		if logErrors {
			w.logger.Error("profile not found", zap.Error(rows.Err()), zap.Any("request", request))
		}
		return nil, status.Error(codes.Internal, "failed to GetProfileInfo")
	}

	return w.decodeProfile(rows)
}

func (w *DWH) getProfileInfoTx(tx *sql.Tx, request *pb.ID) (*pb.Profile, error) {
	rows, err := tx.Query(w.commands["selectProfileByID"], request.Id)
	if err != nil {
		w.logger.Error("failed to selectProfileByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfileInfo")
	}
	defer rows.Close()

	if !rows.Next() {
		w.logger.Error("profile not found", zap.Error(rows.Err()), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfileInfo")
	}

	return w.decodeProfile(rows)
}

func (w *DWH) GetBlacklist(ctx context.Context, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getBlacklist(ctx, request)
}

func (w *DWH) getBlacklist(ctx context.Context, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	var filters []*filter
	if request.OwnerID > "0" {
		filters = append(filters, newFilter("AdderID", eq, request.OwnerID, "AND"))
	}
	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Blacklists",
		filters:  filters,
		sortings: []*pb.SortingOption{{Field: "rowid", Order: pb.SortingOrder_Desc}},
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetBlacklist")
	}
	defer rows.Close()

	var addees []string
	for rows.Next() {
		var (
			adderID string
			addeeID string
		)
		if err := rows.Scan(&adderID, &addeeID); err != nil {
			w.logger.Error("failed to scan blacklist entry", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetBlacklist")
		}

		addees = append(addees, addeeID)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetBlacklist")
	}

	return &pb.BlacklistReply{
		OwnerID:   request.OwnerID,
		Addresses: addees,
	}, nil
}

func (w *DWH) GetValidators(ctx context.Context, request *pb.ValidatorsRequest) (*pb.ValidatorsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getValidators(ctx, request)
}

func (w *DWH) getValidators(ctx context.Context, request *pb.ValidatorsRequest) (*pb.ValidatorsReply, error) {
	var filters []*filter
	if request.ValidatorLevel != nil {
		level := request.ValidatorLevel
		filters = append(filters, newFilter("Level", opsTranslator[level.Operator], level.Value, "AND"))
	}
	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Validators",
		filters:  filters,
		sortings: request.Sortings,
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetValidators")
	}
	defer rows.Close()

	var out []*pb.Validator
	for rows.Next() {
		validator, err := w.decodeValidator(rows)
		if err != nil {
			w.logger.Error("failed to decodeValidator", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetValidators")
		}

		out = append(out, validator)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetValidators")
	}

	return &pb.ValidatorsReply{Validators: out}, nil
}

func (w *DWH) GetDealChangeRequests(ctx context.Context, request *pb.ID) (*pb.DealChangeRequestsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealChangeRequests(ctx, request)
}

func (w *DWH) getDealChangeRequests(ctx context.Context, request *pb.ID) (*pb.DealChangeRequestsReply, error) {
	rows, err := w.db.Query(w.commands["selectDealChangeRequestsByID"], request.Id)
	if err != nil {
		w.logger.Error("failed to selectDealChangeRequestsByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDealChangeRequests")
	}

	var out []*pb.DealChangeRequest
	for rows.Next() {
		changeRequest, err := w.decodeDealChangeRequest(rows)
		if err != nil {
			w.logger.Error("failed to decodeDealChangeRequest", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetDealChangeRequests")
		}
		out = append(out, changeRequest)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetDealChangeRequests")
	}

	return &pb.DealChangeRequestsReply{Requests: out}, nil
}

func (w *DWH) GetWorkers(ctx context.Context, request *pb.WorkersRequest) (*pb.WorkersReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getWorkers(ctx, request)
}

func (w *DWH) getWorkers(ctx context.Context, request *pb.WorkersRequest) (*pb.WorkersReply, error) {
	var filters []*filter
	if request.MasterID > "0" {
		filters = append(filters, newFilter("Level", eq, request.MasterID, "AND"))
	}
	rows, _, err := runQuery(w.db, &queryOpts{
		table:    "Workers",
		filters:  filters,
		sortings: []*pb.SortingOption{{Field: "rowid", Order: pb.SortingOrder_Desc}},
		offset:   request.Offset,
		limit:    request.Limit,
	})
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetWorkers")
	}
	defer rows.Close()

	var out []*pb.DWHWorker
	for rows.Next() {
		worker, err := w.decodeWorker(rows)
		if err != nil {
			w.logger.Error("failed to decodeWorker", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetWorkers")
		}
		out = append(out, worker)
	}

	if err := rows.Err(); err != nil {
		w.logger.Error("failed to read rows from db", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to GetWorkers")
	}

	return &pb.WorkersReply{
		Workers: out,
	}, nil
}

func (w *DWH) monitorBlockchain() error {
	w.logger.Info("starting monitoring")

	retryTime := time.Millisecond * 100
	for {
		if err := w.watchMarketEvents(); err != nil {
			w.logger.Error("failed to watch market events, retrying",
				zap.Error(err), zap.Duration("retry_time", retryTime))
			retryTime *= 2
			if retryTime > time.Second*10 {
				retryTime = time.Second * 10
			}
			time.Sleep(retryTime)
		}
	}
}

func (w *DWH) watchMarketEvents() error {
	lastKnownBlock, err := w.getLastKnownBlockTS()
	if err != nil {
		if err := w.updateLastKnownBlockTS(0); err != nil {
			return err
		}
		lastKnownBlock = 0
	}

	dealEvents, err := w.blockchain.GetEvents(context.Background(), big.NewInt(lastKnownBlock))
	if err != nil {
		return err
	}

	for event := range dealEvents {
		if err := w.updateLastKnownBlockTS(int64(event.BlockNumber)); err != nil {
			w.logger.Error("failed to updateLastKnownBlock", zap.Error(err),
				zap.Uint64("block_number", event.BlockNumber))
		}

		if err := w.processEvent(event); err != nil {
			w.logger.Error("failed to processEvent", zap.Error(err), zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).Name()))
		}

		w.logger.Info("processed event", zap.String("event_type", reflect.TypeOf(event.Data).String()))
	}

	return errors.New("events channel closed")
}

func (w *DWH) processEvent(event *blockchain.Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch value := event.Data.(type) {
	case *blockchain.DealOpenedData:
		return w.onDealOpened(value.ID)
	case *blockchain.DealUpdatedData:
		return w.onDealUpdated(value.ID)
	case *blockchain.OrderPlacedData:
		return w.onOrderPlaced(event.TS, value.ID)
	case *blockchain.OrderUpdatedData:
		return w.onOrderUpdated(value.ID)
	case *blockchain.DealChangeRequestSentData:
		return w.onDealChangeRequestSent(event.TS, value.ID)
	case *blockchain.DealChangeRequestUpdatedData:
		return w.onDealChangeRequestUpdated(event.TS, value.ID)
	case *blockchain.BilledData:
		return w.onBilled(event.TS, value.ID, value.PayedAmount)
	case *blockchain.WorkerAnnouncedData:
		return w.onWorkerAnnounced(value.MasterID.String(), value.SlaveID.String())
	case *blockchain.WorkerConfirmedData:
		return w.onWorkerConfirmed(value.MasterID.String(), value.SlaveID.String())
	case *blockchain.WorkerRemovedData:
		return w.onWorkerRemoved(value.MasterID.String(), value.SlaveID.String())
	case *blockchain.AddedToBlacklistData:
		return w.onAddedToBlacklist(value.AdderID.String(), value.AddeeID.String())
	case *blockchain.RemovedFromBlacklistData:
		w.onRemovedFromBlacklist(value.RemoverID.String(), value.RemoveeID.String())
	case *blockchain.ValidatorCreatedData:
		return w.onValidatorCreated(value.ID)
	case *blockchain.ValidatorDeletedData:
		return w.onValidatorDeleted(value.ID)
	case *blockchain.CertificateCreatedData:
		return w.onCertificateCreated(value.ID)
	case *blockchain.ErrorData:
		w.logger.Error("received error from events channel", zap.Error(value.Err), zap.String("topic", value.Topic))
	}

	return nil
}

func (w *DWH) onDealOpened(dealID *big.Int) error {
	deal, err := w.blockchain.GetDealInfo(w.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	ask, err := w.getOrderDetails(w.ctx, &pb.ID{Id: deal.AskID})
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Ask)")
	}

	bid, err := w.getOrderDetails(w.ctx, &pb.ID{Id: deal.BidID})
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Bid)")
	}

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		return nil
	}

	var hasActiveChangeRequests bool
	if _, err := w.getDealChangeRequests(w.ctx, &pb.ID{Id: deal.Id}); err == nil {
		hasActiveChangeRequests = true
	}

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	_, err = tx.Exec(
		w.commands["insertDeal"],
		deal.Id,
		deal.SupplierID,
		deal.ConsumerID,
		deal.MasterID,
		deal.AskID,
		deal.BidID,
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
		deal.Benchmarks.CPUSysbenchMulti(),
		deal.Benchmarks.CPUSysbenchOne(),
		deal.Benchmarks.CPUCores(),
		deal.Benchmarks.RAMSize(),
		deal.Benchmarks.StorageSize(),
		deal.Benchmarks.NetTrafficIn(),
		deal.Benchmarks.NetTrafficOut(),
		deal.Benchmarks.GPUCount(),
		deal.Benchmarks.GPUMem(),
		deal.Benchmarks.GPUEthHashrate(),
		deal.Benchmarks.GPUCashHashrate(),
		deal.Benchmarks.GPURedshift(),
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrapf(err, "failed to insertDeal")
	}

	_, err = tx.Exec(
		w.commands["insertDealCondition"],
		deal.SupplierID,
		deal.ConsumerID,
		deal.MasterID,
		deal.Duration,
		deal.Price.PaddedString(),
		deal.StartTime.Seconds,
		0,
		deal.TotalPayout.PaddedString(),
		deal.Id,
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrapf(err, "onDealOpened: failed to insert into DealConditions")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) onDealUpdated(dealID *big.Int) error {
	deal, err := w.blockchain.GetDealInfo(context.Background(), dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		tx, err := w.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}

		_, err = tx.Exec(w.commands["deleteDeal"], deal.Id)
		if err != nil {
			w.logger.Info("failed to delete closed Deal (possibly old log entry)", zap.Error(err),
				zap.String("deal_id", deal.Id))

			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return nil
		}

		if _, err := tx.Exec(w.commands["deleteOrder"], deal.AskID); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteOrder")
		}

		if _, err := tx.Exec(w.commands["deleteOrder"], deal.BidID); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteOrder")
		}

		askProfile, err := w.getProfileInfo(w.ctx, &pb.ID{Id: deal.SupplierID}, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to getProfileInfo")
		}

		if err := w.updateProfileStats(tx, pb.OrderType_ASK, deal.SupplierID, askProfile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileStats")
		}

		bidProfile, err := w.getProfileInfo(w.ctx, &pb.ID{Id: deal.ConsumerID}, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to getProfileInfo")
		}

		if err := w.updateProfileStats(tx, pb.OrderType_BID, deal.ConsumerID, bidProfile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileStats")
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "transaction commit failed")
		}

		return nil
	}

	_, err = w.db.Exec(
		w.commands["updateDeal"],
		deal.Duration,
		deal.Price.PaddedString(),
		deal.StartTime.Seconds,
		deal.EndTime.Seconds,
		uint64(deal.Status),
		deal.BlockedBalance.PaddedString(),
		deal.TotalPayout.PaddedString(),
		deal.LastBillTS.Seconds,
		deal.Id,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert into Deals")
	}

	return nil
}

func (w *DWH) onDealChangeRequestSent(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := w.blockchain.GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	if changeRequest.Status != pb.ChangeRequestStatus_REQUEST_CREATED {
		return errors.New("onDealChangeRequest event points to DealChangeRequest with .Status != Created")
	}

	// Sanity check: if more than 1 CR of one type is created for a Deal, we delete old CRs.
	rows, err := w.db.Query(
		w.commands["selectDealChangeRequests"],
		changeRequest.DealID,
		changeRequest.RequestType,
		changeRequest.Status,
	)
	if err != nil {
		return errors.New("failed to get (possibly) expired DealChangeRequests")
	}

	var expiredChangeRequests []*pb.DealChangeRequest
	for rows.Next() {
		if expiredChangeRequest, err := w.decodeDealChangeRequest(rows); err != nil {
			rows.Close()
			return errors.Wrap(err, "failed to decodeDealChangeRequest")
		} else {
			expiredChangeRequests = append(expiredChangeRequests, expiredChangeRequest)
		}
	}

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "failed to fetch all DealChangeRequest(s) from db")
	}

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	for _, expiredChangeRequest := range expiredChangeRequests {
		if _, err := tx.Exec(w.commands["deleteDealChangeRequest"], expiredChangeRequest.Id); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteDealChangeRequest")
		} else {
			w.logger.Warn("deleted expired DealChangeRequest", zap.String("id", expiredChangeRequest.Id))
		}
	}

	_, err = tx.Exec(
		w.commands["insertDealChangeRequest"],
		changeRequest.Id,
		eventTS,
		changeRequest.RequestType,
		changeRequest.Duration,
		changeRequest.Price.PaddedString(),
		changeRequest.Status,
		changeRequest.DealID,
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to insertDealChangeRequest")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return err
}

func (w *DWH) onDealChangeRequestUpdated(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := w.blockchain.GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	switch changeRequest.Status {
	case pb.ChangeRequestStatus_REQUEST_REJECTED:
		_, err := w.db.Exec(
			w.commands["updateDealChangeRequest"],
			changeRequest.Status,
			changeRequest.Id,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to update DealChangeRequest %s", changeRequest.Id)
		}
	case pb.ChangeRequestStatus_REQUEST_ACCEPTED:
		deal, err := w.getDealDetails(w.ctx, &pb.ID{Id: changeRequest.DealID})
		if err != nil {
			return errors.Wrap(err, "failed to getDealDetails")
		}

		tx, err := w.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}

		if err := w.updateDealConditionEndTime(tx, deal.GetDeal().Id, eventTS); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateDealConditionEndTime")
		}
		_, err = tx.Exec(
			w.commands["insertDealCondition"],
			deal.GetDeal().SupplierID,
			deal.GetDeal().ConsumerID,
			deal.GetDeal().MasterID,
			changeRequest.Duration,
			changeRequest.Price.PaddedString(),
			eventTS,
			0,
			"0",
			deal.GetDeal().Id,
		)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to insertDealCondition")
		}

		_, err = tx.Exec(w.commands["deleteDealChangeRequest"], changeRequest.Id)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id)
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "transaction commit failed")
		}
	default:
		_, err := w.db.Exec(w.commands["deleteDealChangeRequest"], changeRequest.Id)
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id)
		}
	}

	return nil
}

func (w *DWH) onBilled(eventTS uint64, dealID, payedAmount *big.Int) error {
	dealConditionsReply, err := w.getDealConditions(w.ctx, &pb.DealConditionsRequest{DealID: dealID.String()})
	if err != nil {
		return errors.Wrap(err, "failed to GetDealConditions (last)")
	}

	dealCondition := dealConditionsReply.Conditions[0]

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	newTotalPayout := big.NewInt(0)
	newTotalPayout.Add(dealCondition.TotalPayout.Unwrap(), payedAmount)
	_, err = tx.Exec(
		w.commands["updateDealConditionPayout"],
		util.BigIntToPaddedString(newTotalPayout),
		dealCondition.Id,
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to update DealCondition")
	}

	_, err = tx.Exec(w.commands["insertDealPayment"], eventTS, util.BigIntToPaddedString(payedAmount),
		dealID.String())
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "insertDealPayment failed")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) onOrderPlaced(eventTS uint64, orderID *big.Int) error {
	order, err := w.blockchain.GetOrderInfo(context.Background(), orderID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetOrderInfo")
	}

	if order.OrderStatus != pb.OrderStatus_ORDER_ACTIVE || order.OrderType == pb.OrderType_ANY {
		return nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	profile, err := w.getProfileInfo(w.ctx, &pb.ID{Id: order.AuthorID}, false)
	if err != nil {
		var askOrders, bidOrders = 0, 0
		if order.OrderType == pb.OrderType_ASK {
			askOrders = 1
		} else {
			bidOrders = 1
		}
		_, err = tx.Exec(w.commands["insertProfileUserID"], order.AuthorID, askOrders, bidOrders)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to insertProfileUserID")
		}

		profile = &pb.Profile{
			UserID:       order.AuthorID,
			Certificates: []byte{},
		}
	} else {
		if err := w.updateProfileStats(tx, order.OrderType, order.AuthorID, profile, 1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileStats")
		}
	}

	_, err = tx.Exec(
		w.commands["insertOrder"],
		order.Id,
		eventTS,
		order.DealID,
		uint64(order.OrderType),
		uint64(order.OrderStatus),
		order.AuthorID,
		order.CounterpartyID,
		order.Duration,
		order.Price.PaddedString(),
		order.Netflags,
		uint64(order.IdentityLevel),
		order.Blacklist,
		order.Tag,
		order.FrozenSum.PaddedString(),
		profile.IdentityLevel,
		profile.Name,
		profile.Country,
		profile.Certificates,
		order.Benchmarks.CPUSysbenchMulti(),
		order.Benchmarks.CPUSysbenchOne(),
		order.Benchmarks.CPUCores(),
		order.Benchmarks.RAMSize(),
		order.Benchmarks.StorageSize(),
		order.Benchmarks.NetTrafficIn(),
		order.Benchmarks.NetTrafficOut(),
		order.Benchmarks.GPUCount(),
		order.Benchmarks.GPUMem(),
		order.Benchmarks.GPUEthHashrate(),
		order.Benchmarks.GPUCashHashrate(),
		order.Benchmarks.GPURedshift(),
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrapf(err, "failed to insertOrder")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) onOrderUpdated(orderID *big.Int) error {
	order, err := w.blockchain.GetOrderInfo(w.ctx, orderID)
	if err != nil {
		return errors.Wrap(err, "failed to GetOrderInfo")
	}

	// If order was updated, but no deal is associated with it, delete the order.
	if order.DealID <= "0" {
		tx, err := w.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}

		if _, err := tx.Exec(w.commands["deleteOrder"], orderID.String()); err != nil {
			w.logger.Info("failed to delete Order (possibly old log entry)", zap.Error(err),
				zap.String("order_id", orderID.String()))

			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return nil
		}

		profile, err := w.getProfileInfo(w.ctx, &pb.ID{Id: order.AuthorID}, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to getProfileInfo")
		}

		if err := w.updateProfileStats(tx, order.OrderType, order.AuthorID, profile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileStats")
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "transaction commit failed")
		}
	}

	return nil
}

func (w *DWH) onWorkerAnnounced(masterID, slaveID string) error {
	_, err := w.db.Exec(
		w.commands["insertWorker"],
		masterID,
		slaveID,
		false,
	)
	if err != nil {
		return errors.Wrap(err, "onWorkerAnnounced failed")
	}

	return nil
}

func (w *DWH) onWorkerConfirmed(masterID, slaveID string) error {
	_, err := w.db.Exec(
		w.commands["updateWorker"],
		true,
		masterID,
		slaveID,
	)
	if err != nil {
		return errors.Wrap(err, "onWorkerConfirmed failed")
	}

	return nil
}

func (w *DWH) onWorkerRemoved(masterID, slaveID string) error {
	_, err := w.db.Exec(
		w.commands["deleteWorker"],
		masterID,
		slaveID,
	)
	if err != nil {
		return errors.Wrap(err, "onWorkerRemoved failed")
	}

	return nil
}

func (w *DWH) onAddedToBlacklist(adderID, addeeID string) error {
	_, err := w.db.Exec(
		w.commands["insertBlacklistEntry"],
		adderID,
		addeeID,
	)
	if err != nil {
		return errors.Wrap(err, "onAddedToBlacklist failed")
	}

	return nil
}

func (w *DWH) onRemovedFromBlacklist(removerID, removeeID string) error {
	_, err := w.db.Exec(
		w.commands["deleteBlacklistEntry"],
		removerID,
		removeeID,
	)
	if err != nil {
		return errors.Wrap(err, "onRemovedFromBlacklist failed")
	}

	return nil
}

func (w *DWH) onValidatorCreated(validatorID common.Address) error {
	validator, err := w.blockchain.GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	_, err = w.db.Exec(w.commands["insertValidator"], validator.Id, validator.Level)
	if err != nil {
		return errors.Wrap(err, "failed to insert Validator")
	}

	return nil
}

func (w *DWH) onValidatorDeleted(validatorID common.Address) error {
	validator, err := w.blockchain.GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	_, err = w.db.Exec(w.commands["updateValidator"], validator.Level, validator.Id)
	if err != nil {
		return errors.Wrap(err, "failed to update Validator")
	}

	return nil
}

func (w *DWH) onCertificateCreated(certificateID *big.Int) error {
	attr, err := w.blockchain.GetCertificate(w.ctx, certificateID)
	if err != nil {
		return errors.Wrap(err, "failed to GetAttr")
	}

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	_, err = tx.Exec(w.commands["insertCertificate"],
		attr.OwnerID, attr.Attribute, (attr.Attribute/uint64(math.Pow(10, 2)))%10, attr.Value, attr.ValidatorID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to insert Certificate")
	}

	// Create a Profile entry if it doesn't exist yet.
	if _, err := w.getProfileInfo(w.ctx, &pb.ID{Id: attr.OwnerID}, false); err != nil {
		_, err = tx.Exec(w.commands["insertProfileUserID"], attr.OwnerID, 0, 0)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to insertProfileUserID")
		}
	}

	// Update distinct Profile columns.
	switch attr.Attribute {
	case CertificateName:
		_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], attributeToString[attr.Attribute]),
			string(attr.Value), attr.OwnerID)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileName")
		}
	case CertificateCountry:
		_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], attributeToString[attr.Attribute]),
			string(attr.Value), attr.OwnerID)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileCountry")
		}
	}

	// Update certificates blob.
	rows, err := tx.Query(w.commands["selectCertificates"], attr.OwnerID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to getCertificatesByUseID")
	}

	var (
		certificates     []*pb.Certificate
		maxIdentityLevel uint64
	)
	for rows.Next() {
		if certificate, err := w.decodeCertificate(rows); err != nil {
			w.logger.Error("failed to decodeCertificate", zap.Error(err))
		} else {
			certificates = append(certificates, certificate)
			if certificate.IdentityLevel > maxIdentityLevel {
				maxIdentityLevel = certificate.IdentityLevel
			}
		}
	}

	certificateAttrsBytes, err := json.Marshal(certificates)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to marshal certificates")
	}

	_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], "Certificates"),
		certificateAttrsBytes, attr.OwnerID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateProfileCertificates (Certificates)")
	}

	_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], "IdentityLevel"),
		maxIdentityLevel, attr.OwnerID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateProfileCertificates (Level)")
	}

	profile, err := w.getProfileInfoTx(tx, &pb.ID{Id: attr.OwnerID})
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to getProfileInfo")
	}

	_, err = tx.Exec(w.commands["updateOrders"],
		profile.IdentityLevel,
		profile.Name,
		profile.Country,
		profile.Certificates,
		profile.UserID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateOrders")
	}

	_, err = tx.Exec(w.commands["updateDealsSupplier"], profile.Certificates, profile.UserID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateDealsSupplier")
	}

	_, err = tx.Exec(w.commands["updateDealsConsumer"], profile.Certificates, profile.UserID)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateDealsConsumer")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) updateProfileStats(tx *sql.Tx, orderType pb.OrderType, authorID string, profile *pb.Profile, update int) error {
	var (
		cmd   string
		value int
	)
	if orderType == pb.OrderType_ASK {
		cmd, value = fmt.Sprintf(w.commands["updateProfile"], "ActiveAsks"), int(profile.ActiveAsks)+update
	} else {
		cmd, value = fmt.Sprintf(w.commands["updateProfile"], "ActiveBids"), int(profile.ActiveBids)+update
	}

	_, err := tx.Exec(cmd, value, authorID)
	if err != nil {
		return errors.Wrap(err, "failed to updateProfile")
	}

	return nil
}

func (w *DWH) decodeDeal(rows *sql.Rows) (*pb.DWHDeal, error) {
	var (
		id                   string
		supplierID           string
		consumerID           string
		masterID             string
		askID                string
		bidID                string
		price                string
		duration             uint64
		startTime            int64
		endTime              int64
		status               uint64
		blockedBalance       string
		totalPayout          string
		lastBillTS           int64
		netflags             uint64
		askIdentityLevel     uint64
		bidIdentityLevel     uint64
		supplierCertificates []byte
		consumerCertificates []byte
		activeChangeRequest  bool
		cpuSysbenchMulti     uint64
		cpuSysbenchOne       uint64
		cpuCores             uint64
		ramSize              uint64
		storageSize          uint64
		netTrafficIn         uint64
		netTrafficOut        uint64
		gpuCount             uint64
		gpuMem               uint64
		gpuEthHashrate       uint64
		gpuCashHashrate      uint64
		gpuRedshift          uint64
	)
	if err := rows.Scan(
		&id,
		&supplierID,
		&consumerID,
		&masterID,
		&askID,
		&bidID,
		&duration,
		&price,
		&startTime,
		&endTime,
		&status,
		&blockedBalance,
		&totalPayout,
		&lastBillTS,
		&netflags,
		&askIdentityLevel,
		&bidIdentityLevel,
		&supplierCertificates,
		&consumerCertificates,
		&activeChangeRequest,
		&cpuSysbenchMulti,
		&cpuSysbenchOne,
		&cpuCores,
		&ramSize,
		&storageSize,
		&netTrafficIn,
		&netTrafficOut,
		&gpuCount,
		&gpuMem,
		&gpuEthHashrate,
		&gpuCashHashrate,
		&gpuRedshift,
	); err != nil {
		w.logger.Error("failed to scan deal row", zap.Error(err))
		return nil, err
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigBlockedBalance := new(big.Int)
	bigBlockedBalance.SetString(blockedBalance, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(totalPayout, 10)

	return &pb.DWHDeal{
		Deal: &pb.Deal{
			Id:             id,
			SupplierID:     supplierID,
			ConsumerID:     consumerID,
			MasterID:       masterID,
			AskID:          askID,
			BidID:          bidID,
			Price:          pb.NewBigInt(bigPrice),
			Duration:       duration,
			StartTime:      &pb.Timestamp{Seconds: startTime},
			EndTime:        &pb.Timestamp{Seconds: endTime},
			Status:         pb.DealStatus(status),
			BlockedBalance: pb.NewBigInt(bigBlockedBalance),
			TotalPayout:    pb.NewBigInt(bigTotalPayout),
			LastBillTS:     &pb.Timestamp{Seconds: lastBillTS},
			Benchmarks: &pb.Benchmarks{
				Values: []uint64{
					cpuSysbenchMulti,
					cpuSysbenchOne,
					cpuCores,
					ramSize,
					storageSize,
					netTrafficIn,
					netTrafficOut,
					gpuCount,
					gpuMem,
					gpuEthHashrate,
					gpuCashHashrate,
					gpuRedshift,
				},
			},
		},
		Netflags:             netflags,
		AskIdentityLevel:     askIdentityLevel,
		BidIdentityLevel:     bidIdentityLevel,
		SupplierCertificates: supplierCertificates,
		ConsumerCertificates: consumerCertificates,
		ActiveChangeRequest:  activeChangeRequest,
	}, nil
}

func (w *DWH) decodeDealChangeRequest(rows *sql.Rows) (*pb.DealChangeRequest, error) {
	var (
		changeRequestID string
		createdTS       uint64
		requestType     uint64
		duration        uint64
		price           string
		status          uint64
		dealID          string
	)
	if err := rows.Scan(
		&changeRequestID,
		&createdTS,
		&requestType,
		&duration,
		&price,
		&status,
		&dealID,
	); err != nil {
		w.logger.Error("failed to scan DealChangeRequest row", zap.Error(err))
		return nil, err
	}
	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)

	return &pb.DealChangeRequest{
		Id:          changeRequestID,
		DealID:      dealID,
		RequestType: pb.OrderType(requestType),
		Duration:    duration,
		Price:       pb.NewBigInt(bigPrice),
		Status:      pb.ChangeRequestStatus(status),
	}, nil
}

func (w *DWH) decodeOrder(rows *sql.Rows) (*pb.DWHOrder, error) {
	var (
		id                   string
		createdTS            uint64
		dealID               string
		orderType            uint64
		status               uint64
		author               string
		counterAgent         string
		price                string
		duration             uint64
		netflags             uint64
		identityLevel        uint64
		blacklist            string
		tag                  []byte
		frozenSum            string
		creatorIdentityLevel uint64
		creatorName          string
		creatorCountry       string
		creatorCertificates  []byte
		cpuSysbenchMulti     uint64
		cpuSysbenchOne       uint64
		cpuCores             uint64
		ramSize              uint64
		storageSize          uint64
		netTrafficIn         uint64
		netTrafficOut        uint64
		gpuCount             uint64
		gpuMem               uint64
		gpuEthHashrate       uint64
		gpuCashHashrate      uint64
		gpuRedshift          uint64
	)
	if err := rows.Scan(
		&id,
		&createdTS,
		&dealID,
		&orderType,
		&status,
		&author,
		&counterAgent,
		&duration,
		&price,
		&netflags,
		&identityLevel,
		&blacklist,
		&tag,
		&frozenSum,
		&creatorIdentityLevel,
		&creatorName,
		&creatorCountry,
		&creatorCertificates,
		&cpuSysbenchMulti,
		&cpuSysbenchOne,
		&cpuCores,
		&ramSize,
		&storageSize,
		&netTrafficIn,
		&netTrafficOut,
		&gpuCount,
		&gpuMem,
		&gpuEthHashrate,
		&gpuCashHashrate,
		&gpuRedshift,
	); err != nil {
		w.logger.Error("failed to scan order row", zap.Error(err))
		return nil, err
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigFrozenSum := new(big.Int)
	bigFrozenSum.SetString(frozenSum, 10)

	return &pb.DWHOrder{
		Order: &pb.Order{
			Id:             id,
			DealID:         dealID,
			OrderType:      pb.OrderType(orderType),
			OrderStatus:    pb.OrderStatus(status),
			AuthorID:       author,
			CounterpartyID: counterAgent,
			Duration:       duration,
			Price:          pb.NewBigInt(bigPrice),
			Netflags:       netflags,
			IdentityLevel:  pb.IdentityLevel(identityLevel),
			Blacklist:      blacklist,
			Tag:            tag,
			FrozenSum:      pb.NewBigInt(bigFrozenSum),
			Benchmarks: &pb.Benchmarks{
				Values: []uint64{
					cpuSysbenchMulti,
					cpuSysbenchOne,
					cpuCores,
					ramSize,
					storageSize,
					netTrafficIn,
					netTrafficOut,
					gpuCount,
					gpuMem,
					gpuEthHashrate,
					gpuCashHashrate,
					gpuRedshift,
				},
			},
		},

		CreatedTS:            createdTS,
		CreatorIdentityLevel: creatorIdentityLevel,
		CreatorName:          creatorName,
		CreatorCountry:       creatorCountry,
		CreatorCertificates:  creatorCertificates,
	}, nil
}

func (w *DWH) decodeDealCondition(rows *sql.Rows) (*pb.DealCondition, error) {
	var (
		rowid       uint64
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
		&rowid,
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
		w.logger.Error("failed to scan DealCondition row", zap.Error(err))
		return nil, err
	}

	bigPrice := new(big.Int)
	bigPrice.SetString(price, 10)
	bigTotalPayout := new(big.Int)
	bigTotalPayout.SetString(totalPayout, 10)

	return &pb.DealCondition{
		Id:          rowid,
		SupplierID:  supplierID,
		ConsumerID:  consumerID,
		MasterID:    masterID,
		Price:       pb.NewBigInt(bigPrice),
		Duration:    duration,
		StartTime:   &pb.Timestamp{Seconds: startTime},
		EndTime:     &pb.Timestamp{Seconds: endTime},
		TotalPayout: pb.NewBigInt(bigTotalPayout),
		DealID:      dealID,
	}, nil
}

func (w *DWH) decodeCertificate(rows *sql.Rows) (*pb.Certificate, error) {
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
			OwnerID:       ownerID,
			Attribute:     attribute,
			IdentityLevel: identityLevel,
			Value:         value,
			ValidatorID:   validatorID,
		}, nil
	}
}

func (w *DWH) decodeProfile(rows *sql.Rows) (*pb.Profile, error) {
	var (
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
		w.logger.Error("failed to scan deal row", zap.Error(err))
		return nil, err
	}

	return &pb.Profile{
		UserID:         userID,
		IdentityLevel:  identityLevel,
		Name:           name,
		Country:        country,
		IsCorporation:  isCorporation,
		IsProfessional: isProfessional,
		Certificates:   certificates,
		ActiveAsks:     activeAsks,
		ActiveBids:     activeBids,
	}, nil
}

func (w *DWH) decodeValidator(rows *sql.Rows) (*pb.Validator, error) {
	var (
		validatorID string
		level       uint64
	)
	if err := rows.Scan(&validatorID, &level); err != nil {
		return nil, errors.Wrap(err, "failed to scan Validator row")
	}

	return &pb.Validator{
		Id:    validatorID,
		Level: level,
	}, nil
}

func (w *DWH) decodeWorker(rows *sql.Rows) (*pb.DWHWorker, error) {
	var (
		masterID  string
		slaveID   string
		confirmed bool
	)
	if err := rows.Scan(&masterID, &slaveID, &confirmed); err != nil {
		return nil, errors.Wrap(err, "failed to scan Worker row")
	}

	return &pb.DWHWorker{
		MasterID:  masterID,
		SlaveID:   slaveID,
		Confirmed: confirmed,
	}, nil
}

func (w *DWH) addBenchmarksConditions(benches *pb.DWHBenchmarkConditions, filters *[]*filter) {
	if benches.CPUSysbenchMulti != nil {
		if benches.CPUSysbenchMulti.Max > 0 {
			*filters = append(*filters, newFilter("CPUSysbenchMulti", lte, benches.CPUSysbenchMulti.Max, "AND"))
		}
		*filters = append(*filters, newFilter("CPUSysbenchMulti", gte, benches.CPUSysbenchMulti.Min, "AND"))
	}
	if benches.CPUSysbenchOne != nil {
		if benches.CPUSysbenchOne.Max > 0 {
			*filters = append(*filters, newFilter("CPUSysbenchOne", lte, benches.CPUSysbenchOne.Max, "AND"))
		}
		*filters = append(*filters, newFilter("CPUSysbenchOne", gte, benches.CPUSysbenchOne.Min, "AND"))
	}
	if benches.CPUCores != nil {
		if benches.CPUCores.Max > 0 {
			*filters = append(*filters, newFilter("CPUCores", lte, benches.CPUCores.Max, "AND"))
		}
		*filters = append(*filters, newFilter("CPUCores", gte, benches.CPUCores.Min, "AND"))
	}
	if benches.RAMSize != nil {
		if benches.RAMSize.Max > 0 {
			*filters = append(*filters, newFilter("RAMSize", lte, benches.RAMSize.Max, "AND"))
		}
		*filters = append(*filters, newFilter("RAMSize", gte, benches.RAMSize.Min, "AND"))
	}
	if benches.StorageSize != nil {
		if benches.StorageSize.Max > 0 {
			*filters = append(*filters, newFilter("StorageSize", lte, benches.StorageSize.Max, "AND"))
		}
		*filters = append(*filters, newFilter("StorageSize", gte, benches.StorageSize.Min, "AND"))
	}
	if benches.NetTrafficIn != nil {
		if benches.NetTrafficIn.Max > 0 {
			*filters = append(*filters, newFilter("NetTrafficIn", lte, benches.NetTrafficIn.Max, "AND"))
		}
		*filters = append(*filters, newFilter("NetTrafficIn", gte, benches.NetTrafficIn.Min, "AND"))
	}
	if benches.NetTrafficOut != nil {
		if benches.NetTrafficOut.Max > 0 {
			*filters = append(*filters, newFilter("NetTrafficOut", lte, benches.NetTrafficOut.Max, "AND"))
		}
		*filters = append(*filters, newFilter("NetTrafficOut", gte, benches.NetTrafficOut.Min, "AND"))
	}
	if benches.GPUCount != nil {
		if benches.GPUCount.Max > 0 {
			*filters = append(*filters, newFilter("GPUCount", lte, benches.GPUCount.Max, "AND"))
		}
		*filters = append(*filters, newFilter("GPUCount", gte, benches.GPUCount.Min, "AND"))
	}
	if benches.GPUMem != nil {
		if benches.GPUMem.Max > 0 {
			*filters = append(*filters, newFilter("GPUMem", lte, benches.GPUMem.Max, "AND"))
		}
		*filters = append(*filters, newFilter("GPUMem", gte, benches.GPUMem.Min, "AND"))
	}
	if benches.GPUEthHashrate != nil {
		if benches.GPUEthHashrate.Max > 0 {
			*filters = append(*filters, newFilter("GPUEthHashrate", lte, benches.GPUEthHashrate.Max, "AND"))
		}
		*filters = append(*filters, newFilter("GPUEthHashrate", gte, benches.GPUEthHashrate.Min, "AND"))
	}
	if benches.GPUCashHashrate != nil {
		if benches.GPUCashHashrate.Max > 0 {
			*filters = append(*filters, newFilter("GPUCashHashrate", lte, benches.GPUCashHashrate.Max, "AND"))
		}
		*filters = append(*filters, newFilter("GPUCashHashrate", gte, benches.GPUCashHashrate.Min, "AND"))
	}
	if benches.GPURedshift != nil {
		if benches.GPURedshift.Max > 0 {
			*filters = append(*filters, newFilter("GPURedshift", lte, benches.GPURedshift.Max, "AND"))
		}
		*filters = append(*filters, newFilter("GPURedshift", gte, benches.GPURedshift.Min, "AND"))
	}
}

func (w *DWH) getLastKnownBlockTS() (int64, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	rows, err := w.db.Query(w.commands["selectLastKnownBlock"])
	if err != nil {
		return -1, errors.Wrap(err, "failed to selectLastKnownBlock")
	}
	defer rows.Close()

	if ok := rows.Next(); !ok {
		return -1, errors.New("selectLastKnownBlock: no entries")
	}

	var lastKnownBlock int64
	if err := rows.Scan(&lastKnownBlock); err != nil {
		return -1, errors.Wrapf(err, "failed to parse last known block number")
	}

	return lastKnownBlock, nil
}

func (w *DWH) updateLastKnownBlockTS(blockNumber int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.db.Exec(w.commands["updateLastKnownBlock"], blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (w *DWH) updateDealConditionEndTime(tx *sql.Tx, dealID string, eventTS uint64) error {
	dealConditionsReply, err := w.getDealConditions(w.ctx, &pb.DealConditionsRequest{DealID: dealID})
	if err != nil {
		return errors.Wrap(err, "failed to getDealConditions")
	}

	dealCondition := dealConditionsReply.Conditions[0]

	if _, err := tx.Exec(w.commands["updateDealConditionEndTime"], eventTS, dealCondition.Id); err != nil {
		return errors.Wrap(err, "failed to update DealCondition")
	}

	return nil
}
