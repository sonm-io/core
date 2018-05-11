package dwh

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net"
	"reflect"
	"strings"
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

const (
	eventRetryTime = time.Second * 3
	numWorkers     = 40
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
	runQuery    QueryRunner
}

func NewDWH(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (*DWH, error) {
	ctx, cancel := context.WithCancel(ctx)
	w := &DWH{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		logger: log.GetLogger(ctx),
	}

	setupDB, ok := setupDBCallbacks[cfg.Storage.Backend]
	if !ok {
		cancel()
		return nil, errors.Errorf("unsupported backend: %s", cfg.Storage.Backend)
	}

	if err := setupDB(w); err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to setupDB")
	}

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		w.db.Close()
		cancel()
		return nil, err
	}

	w.certRotator = certRotator
	w.creds = util.NewTLS(TLSConfig)
	w.grpc = xgrpc.NewServer(w.logger, xgrpc.Credentials(w.creds), xgrpc.DefaultTraceInterceptor())
	pb.RegisterDWHServer(w.grpc, w)
	grpc_prometheus.Register(w.grpc)

	return w, nil
}

func (w *DWH) Serve() error {
	if w.cfg.Blockchain != nil {
		bch, err := blockchain.NewAPI(blockchain.WithConfig(w.cfg.Blockchain))
		if err != nil {
			return errors.Wrap(err, "failed to create NewAPI")
		}
		w.blockchain = bch
		go w.monitorBlockchain()
	} else {
		w.logger.Info("monitoring disabled")
	}

	w.logger.Info("starting with backend", zap.String("backend", w.cfg.Storage.Backend),
		zap.String("endpoint", w.cfg.Storage.Endpoint))

	wg := errgroup.Group{}
	wg.Go(w.serveGRPC)
	wg.Go(w.serveHTTP)

	return wg.Wait()
}

func (w *DWH) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.cancel()
	w.db.Close()
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
	if request.SupplierID != nil && !request.SupplierID.IsZero() {
		filters = append(filters, newFilter("SupplierID", eq, request.SupplierID.Unwrap().Hex(), "AND"))
	}
	if request.ConsumerID != nil && !request.ConsumerID.IsZero() {
		filters = append(filters, newFilter("ConsumerID", eq, request.ConsumerID.Unwrap().Hex(), "AND"))
	}
	if request.MasterID != nil && !request.MasterID.IsZero() {
		filters = append(filters, newFilter("MasterID", eq, request.MasterID.Unwrap().Hex(), "AND"))
	}
	if request.AskID != nil && !request.AskID.IsZero() {
		filters = append(filters, newFilter("AskID", eq, request.AskID, "AND"))
	}
	if request.BidID != nil && !request.BidID.IsZero() {
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

	rows, count, err := w.runQuery(w.db, &queryOpts{
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

	return &pb.DWHDealsReply{Deals: deals, Count: count}, nil
}

func (w *DWH) GetDealDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHDeal, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealDetails(ctx, request)
}

func (w *DWH) getDealDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHDeal, error) {
	rows, err := w.db.Query(w.commands["selectDealByID"], request.Unwrap().String())
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
		request.Sortings = []*pb.SortingOption{{Field: "Id", Order: pb.SortingOrder_Desc}}
	}

	filters = append(filters, newFilter("DealID", eq, request.DealID.Unwrap().String(), "AND"))
	rows, _, err := w.runQuery(w.db, &queryOpts{
		table:    "DealConditions",
		filters:  filters,
		sortings: request.Sortings,
		offset:   request.Offset,
		limit:    request.Limit,
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
	if request.DealID != nil && !request.DealID.IsZero() {
		filters = append(filters, newFilter("DealID", eq, request.DealID.Unwrap().String(), "AND"))
	}
	if request.Type > 0 {
		filters = append(filters, newFilter("Type", eq, request.Type, "AND"))
	}
	if request.AuthorID != nil && !request.AuthorID.IsZero() {
		filters = append(filters, newFilter("AuthorID", eq, request.AuthorID.Unwrap().Hex(), "AND"))
	}
	if request.CounterpartyID != nil && !request.CounterpartyID.IsZero() {
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
		w.addBenchmarksConditions(request.Benchmarks, &filters)
	}
	rows, count, err := w.runQuery(w.db, &queryOpts{
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

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
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
	if order.Order.CounterpartyID != nil && !order.Order.CounterpartyID.IsZero() {
		filters = append(filters, newFilter("AuthorID", eq, order.Order.CounterpartyID.Unwrap().Hex(), "AND"))
	}
	counterpartyFilter := newFilter("CounterpartyID", eq, "", "OR")
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

	rows, count, err := w.runQuery(w.db, &queryOpts{
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

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (w *DWH) GetOrderDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHOrder, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getOrderDetails(ctx, request)
}

func (w *DWH) getOrderDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHOrder, error) {
	rows, err := w.db.Query(w.commands["selectOrderByID"], request.Unwrap().String())
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
		table:    "Profiles",
		filters:  filters,
		sortings: filterSortings(request.Sortings, ProfilesColumns),
		offset:   request.Offset,
		limit:    request.Limit,
	}
	if request.BlacklistQuery != nil && request.BlacklistQuery.OwnerID != nil {
		opts.selectAs = "AS p"
		switch request.BlacklistQuery.Option {
		case pb.BlacklistOption_WithoutMatching:
			opts.customFilter = &customFilter{
				clause: w.commands["profileNotInBlacklist"],
				values: []interface{}{request.BlacklistQuery.OwnerID.Unwrap().Hex()},
			}
		case pb.BlacklistOption_OnlyMatching:
			opts.customFilter = &customFilter{
				clause: w.commands["profileInBlacklist"],
				values: []interface{}{request.BlacklistQuery.OwnerID.Unwrap().Hex()},
			}
		}
	}

	rows, count, err := w.runQuery(w.db, opts)
	if err != nil {
		w.logger.Error("failed to runQuery", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfiles")
	}
	defer rows.Close()

	var out []*pb.Profile
	for rows.Next() {
		if profile, err := w.decodeProfile(rows); err != nil {
			w.logger.Error("failed to decodeProfile", zap.Error(err), zap.Any("request", request))
			return nil, status.Error(codes.Internal, "failed to GetProfiles")
		} else {
			out = append(out, profile)
		}
	}
	if err := rows.Err(); err != nil {
		w.logger.Error("failed to fetch profiles from db", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetProfiles")
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
			if blacklistedAddrs[profile.UserID.Unwrap().Hex()] {
				profile.IsBlacklisted = true
			}
		}
	}

	return &pb.ProfilesReply{Profiles: out, Count: count}, nil
}

func (w *DWH) GetProfileInfo(ctx context.Context, request *pb.EthID) (*pb.Profile, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getProfileInfo(ctx, request.GetId(), true)
}

func (w *DWH) getProfileInfo(ctx context.Context, request *pb.EthAddress, logErrors bool) (*pb.Profile, error) {
	rows, err := w.db.Query(w.commands["selectProfileByID"], request.Unwrap().Hex())
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

func (w *DWH) getProfileInfoTx(tx *sql.Tx, request *pb.EthAddress) (*pb.Profile, error) {
	rows, err := tx.Query(w.commands["selectProfileByID"], request.Unwrap().Hex())
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
	if request.OwnerID != nil && !request.OwnerID.IsZero() {
		filters = append(filters, newFilter("AdderID", eq, request.OwnerID.Unwrap().Hex(), "AND"))
	}
	rows, count, err := w.runQuery(w.db, &queryOpts{
		table:    "Blacklists",
		filters:  filters,
		sortings: []*pb.SortingOption{},
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
		Count:     count,
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
	rows, count, err := w.runQuery(w.db, &queryOpts{
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

	return &pb.ValidatorsReply{Validators: out, Count: count}, nil
}

func (w *DWH) GetDealChangeRequests(ctx context.Context, request *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.getDealChangeRequests(ctx, request)
}

func (w *DWH) getDealChangeRequests(ctx context.Context, request *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	rows, err := w.db.Query(w.commands["selectDealChangeRequestsByID"], request.Unwrap().String())
	if err != nil {
		w.logger.Error("failed to selectDealChangeRequestsByID", zap.Error(err), zap.Any("request", request))
		return nil, status.Error(codes.Internal, "failed to GetDealChangeRequests")
	}
	defer rows.Close()

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
	if request.MasterID != nil && !request.MasterID.IsZero() {
		filters = append(filters, newFilter("Level", eq, request.MasterID, "AND"))
	}
	rows, count, err := w.runQuery(w.db, &queryOpts{
		table:    "Workers",
		filters:  filters,
		sortings: []*pb.SortingOption{},
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
		Count:   count,
	}, nil
}

func (w *DWH) monitorBlockchain() error {
	w.logger.Info("starting monitoring")

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("context cancelled (monitorBlockchain)")
			return nil
		default:
			if err := w.watchMarketEvents(); err != nil {
				w.logger.Error("failed to watch market events, retrying", zap.Error(err))
			}
		}
	}
}

func (w *DWH) watchMarketEvents() error {
	lastKnownBlock, err := w.getLastKnownBlockTS()
	if err != nil {
		if err := w.insertLastKnownBlockTS(0); err != nil {
			return err
		}
		lastKnownBlock = 0
	}

	w.logger.Info("starting from block", zap.Int64("block_number", lastKnownBlock))

	events, err := w.blockchain.Events().GetEvents(w.ctx, big.NewInt(lastKnownBlock))
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	for workerID := 0; workerID < numWorkers; workerID++ {
		wg.Add(1)
		go w.runEventWorker(wg, workerID, events)
	}
	wg.Wait()

	return nil
}

func (w *DWH) runEventWorker(wg *sync.WaitGroup, workerID int, events chan *blockchain.Event) {
	defer wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("context cancelled (watchMarketEvents)", zap.Int("worker_id", workerID))
			return
		case event, ok := <-events:
			if !ok {
				w.logger.Info("events channel closed", zap.Int("worker_id", workerID))
				return
			}
			if err := w.updateLastKnownBlockTS(int64(event.BlockNumber)); err != nil {
				w.logger.Error("failed to updateLastKnownBlock", zap.Error(err),
					zap.Uint64("block_number", event.BlockNumber), zap.Int("worker_id", workerID))
			}
			// Events in the same block can come in arbitrary order. If two events have to be processed
			// in a specific order (e.g., OrderPlaced > DealOpened), we need to retry if the order is
			// messed up.
			if err := w.processEvent(event); err != nil {
				if strings.Contains(err.Error(), "constraint") {
					continue
				}
				w.logger.Error("failed to processEvent, retrying", zap.Error(err),
					zap.Uint64("block_number", event.BlockNumber),
					zap.String("event_type", reflect.TypeOf(event.Data).String()),
					zap.Any("event_data", event.Data), zap.Int("worker_id", workerID))
				w.retryEvent(event)
			}
			w.logger.Debug("processed event", zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).String()),
				zap.Any("event_data", event.Data), zap.Int("worker_id", workerID))
		}
	}
}

func (w *DWH) processEvent(event *blockchain.Event) error {
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
		return w.onBilled(event.TS, value.DealID, value.PaidAmount)
	case *blockchain.WorkerAnnouncedData:
		return w.onWorkerAnnounced(value.MasterID.Hex(), value.SlaveID.Hex())
	case *blockchain.WorkerConfirmedData:
		return w.onWorkerConfirmed(value.MasterID.Hex(), value.SlaveID.Hex())
	case *blockchain.WorkerRemovedData:
		return w.onWorkerRemoved(value.MasterID.Hex(), value.SlaveID.Hex())
	case *blockchain.AddedToBlacklistData:
		return w.onAddedToBlacklist(value.AdderID.Hex(), value.AddeeID.Hex())
	case *blockchain.RemovedFromBlacklistData:
		w.onRemovedFromBlacklist(value.RemoverID.Hex(), value.RemoveeID.Hex())
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

func (w *DWH) retryEvent(event *blockchain.Event) {
	timer := time.NewTimer(eventRetryTime)
	select {
	case <-w.ctx.Done():
		w.logger.Info("context cancelled while retrying event",
			zap.Uint64("block_number", event.BlockNumber),
			zap.String("event_type", reflect.TypeOf(event.Data).Name()))
		return
	case <-timer.C:
		if err := w.processEvent(event); err != nil {
			w.logger.Error("failed to retry processEvent", zap.Error(err),
				zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).String()),
				zap.Any("event_data", event.Data))
		}
	}
}

func (w *DWH) onDealOpened(dealID *big.Int) error {
	deal, err := w.blockchain.Market().GetDealInfo(w.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	ask, err := w.getOrderDetails(w.ctx, deal.AskID)
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Ask)")
	}

	bid, err := w.getOrderDetails(w.ctx, deal.BidID)
	if err != nil {
		return errors.Wrapf(err, "failed to getOrderDetails (Bid)")
	}

	var hasActiveChangeRequests bool
	if _, err := w.getDealChangeRequests(w.ctx, deal.Id); err == nil {
		hasActiveChangeRequests = true
	}

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	if err := CheckBenchmarks(deal.Benchmarks); err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return err
	}

	_, err = tx.Exec(
		w.commands["insertDeal"],
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
		deal.SupplierID.Unwrap().Hex(),
		deal.ConsumerID.Unwrap().Hex(),
		deal.MasterID.Unwrap().Hex(),
		deal.Duration,
		deal.Price.PaddedString(),
		deal.StartTime.Seconds,
		0,
		deal.TotalPayout.PaddedString(),
		deal.Id.Unwrap().String(),
	)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrapf(err, "onDealOpened: failed to insertDealCondition")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) onDealUpdated(dealID *big.Int) error {
	deal, err := w.blockchain.Market().GetDealInfo(w.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		tx, err := w.db.Begin()
		if err != nil {
			return errors.Wrap(err, "failed to begin transaction")
		}

		_, err = tx.Exec(w.commands["deleteDeal"], deal.Id.Unwrap().String())
		if err != nil {
			w.logger.Info("failed to delete closed Deal (possibly old log entry)", zap.Error(err),
				zap.String("deal_id", deal.Id.Unwrap().String()))

			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return nil
		}

		if _, err := tx.Exec(w.commands["deleteOrder"], deal.AskID.Unwrap().String()); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteOrder")
		}

		if _, err := tx.Exec(w.commands["deleteOrder"], deal.BidID.Unwrap().String()); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteOrder")
		}

		askProfile, err := w.getProfileInfo(w.ctx, deal.SupplierID, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to getProfileInfo")
		}

		if err := w.updateProfileStats(tx, pb.OrderType_ASK, deal.SupplierID.Unwrap().Hex(), askProfile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to updateProfileStats (OrderID: `%s`)", deal.AskID.Unwrap().String())
		}

		bidProfile, err := w.getProfileInfo(w.ctx, deal.ConsumerID, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to getProfileInfo")
		}

		if err := w.updateProfileStats(tx, pb.OrderType_BID, deal.ConsumerID.Unwrap().Hex(), bidProfile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to updateProfileStats (OrderID: `%s`)", deal.BidID.Unwrap().String())
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
		deal.Id.Unwrap().String(),
	)
	if err != nil {
		return errors.Wrapf(err, "failed to insert into Deals")
	}

	return nil
}

func (w *DWH) onDealChangeRequestSent(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := w.blockchain.Market().GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if changeRequest.Status != pb.ChangeRequestStatus_REQUEST_CREATED {
		w.logger.Info("onDealChangeRequest event points to DealChangeRequest with .Status != Created",
			zap.String("actual_status", pb.ChangeRequestStatus_name[int32(changeRequest.Status)]))
		return nil
	}

	// Sanity check: if more than 1 CR of one type is created for a Deal, we delete old CRs.
	rows, err := w.db.Query(
		w.commands["selectDealChangeRequests"],
		changeRequest.DealID.Unwrap().String(),
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
		_, err := tx.Exec(w.commands["deleteDealChangeRequest"], expiredChangeRequest.Id.Unwrap().String())
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to deleteDealChangeRequest")
		} else {
			w.logger.Warn("deleted expired DealChangeRequest",
				zap.String("id", expiredChangeRequest.Id.Unwrap().String()))
		}
	}

	_, err = tx.Exec(
		w.commands["insertDealChangeRequest"],
		changeRequest.Id.Unwrap().String(),
		eventTS,
		changeRequest.RequestType,
		changeRequest.Duration,
		changeRequest.Price.PaddedString(),
		changeRequest.Status,
		changeRequest.DealID.Unwrap().String(),
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
	changeRequest, err := w.blockchain.Market().GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	switch changeRequest.Status {
	case pb.ChangeRequestStatus_REQUEST_REJECTED:
		_, err := w.db.Exec(
			w.commands["updateDealChangeRequest"],
			changeRequest.Status,
			changeRequest.Id.Unwrap().String(),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to update DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	case pb.ChangeRequestStatus_REQUEST_ACCEPTED:
		deal, err := w.getDealDetails(w.ctx, changeRequest.DealID)
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
			deal.GetDeal().SupplierID.Unwrap().Hex(),
			deal.GetDeal().ConsumerID.Unwrap().Hex(),
			deal.GetDeal().MasterID.Unwrap().Hex(),
			changeRequest.Duration,
			changeRequest.Price.PaddedString(),
			eventTS,
			0,
			"0",
			deal.GetDeal().Id.Unwrap().String(),
		)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to insertDealCondition")
		}

		_, err = tx.Exec(w.commands["deleteDealChangeRequest"], changeRequest.Id.Unwrap().String())
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "transaction commit failed")
		}
	default:
		_, err := w.db.Exec(w.commands["deleteDealChangeRequest"], changeRequest.Id.Unwrap().String())
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	}

	return nil
}

func (w *DWH) onBilled(eventTS uint64, dealID, payedAmount *big.Int) error {
	dealConditionsReply, err := w.getDealConditions(w.ctx, &pb.DealConditionsRequest{DealID: pb.NewBigInt(dealID)})
	if err != nil {
		return errors.Wrap(err, "failed to GetDealConditions (last)")
	}

	if len(dealConditionsReply.Conditions) < 1 {
		return errors.Errorf("no deal conditions found for deal `%s`", dealID.String())
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
	order, err := w.blockchain.Market().GetOrderInfo(w.ctx, orderID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetOrderInfo")
	}

	if order.OrderStatus == pb.OrderStatus_ORDER_INACTIVE && order.DealID.IsZero() {
		w.logger.Info("skipping inactive order", zap.String("order_id", order.Id.Unwrap().String()))
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	profile, err := w.getProfileInfo(w.ctx, order.AuthorID, false)
	if err != nil {
		var askOrders, bidOrders = 0, 0
		if order.OrderType == pb.OrderType_ASK {
			askOrders = 1
		} else {
			bidOrders = 1
		}
		_, err = tx.Exec(w.commands["insertProfileUserID"], order.AuthorID.Unwrap().Hex(), askOrders, bidOrders)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to insertProfileUserID")
		}

		profile = &pb.Profile{
			UserID:       order.AuthorID,
			Certificates: "",
		}
	} else {
		if err := w.updateProfileStats(tx, order.OrderType, order.AuthorID.Unwrap().Hex(), profile, 1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrap(err, "failed to updateProfileStats")
		}
	}

	if order.DealID == nil {
		order.DealID = pb.NewBigIntFromInt(0)
	}

	if err := CheckBenchmarks(order.Benchmarks); err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return err
	}

	_, err = tx.Exec(
		w.commands["insertOrder"],
		order.Id.Unwrap().String(),
		eventTS,
		order.DealID.Unwrap().String(),
		uint64(order.OrderType),
		uint64(order.OrderStatus),
		order.AuthorID.Unwrap().Hex(),
		order.CounterpartyID.Unwrap().Hex(),
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
		[]byte(profile.Certificates),
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
	order, err := w.blockchain.Market().GetOrderInfo(w.ctx, orderID)
	if err != nil {
		return errors.Wrap(err, "failed to GetOrderInfo")
	}

	if order.OrderStatus == pb.OrderStatus_ORDER_INACTIVE && order.DealID.IsZero() {
		w.logger.Info("skipping inactive order", zap.String("order_id", order.Id.Unwrap().String()))
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// If order was updated, but no deal is associated with it, delete the order.
	if order.DealID.IsZero() {
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

		profile, err := w.getProfileInfo(w.ctx, order.AuthorID, false)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to getProfileInfo (AuthorID: `%s`)", order.AuthorID)
		}

		if err := w.updateProfileStats(tx, order.OrderType, order.AuthorID.Unwrap().Hex(), profile, -1); err != nil {
			if err := tx.Rollback(); err != nil {
				w.logger.Error("transaction rollback failed", zap.Error(err))
			}

			return errors.Wrapf(err, "failed to updateProfileStats (AuthorID: `%s`)", order.AuthorID.Unwrap().String())
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "transaction commit failed")
		}
	}

	return nil
}

func (w *DWH) onWorkerAnnounced(masterID, slaveID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

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
	w.mu.Lock()
	defer w.mu.Unlock()

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
	w.mu.Lock()
	defer w.mu.Unlock()

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
	w.mu.Lock()
	defer w.mu.Unlock()

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
	w.mu.Lock()
	defer w.mu.Unlock()

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
	validator, err := w.blockchain.ProfileRegistry().GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	_, err = w.db.Exec(w.commands["insertValidator"], validator.Id.Unwrap().Hex(), validator.Level)
	if err != nil {
		return errors.Wrap(err, "failed to insertValidator")
	}

	return nil
}

func (w *DWH) onValidatorDeleted(validatorID common.Address) error {
	validator, err := w.blockchain.ProfileRegistry().GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	_, err = w.db.Exec(w.commands["updateValidator"], validator.Level, validator.Id.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateValidator")
	}

	return nil
}

func (w *DWH) onCertificateCreated(certificateID *big.Int) error {
	certificate, err := w.blockchain.ProfileRegistry().GetCertificate(w.ctx, certificateID)
	if err != nil {
		return errors.Wrap(err, "failed to GetCertificate")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	tx, err := w.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	_, err = tx.Exec(w.commands["insertCertificate"],
		certificate.OwnerID.Unwrap().Hex(),
		certificate.Attribute,
		(certificate.Attribute/uint64(math.Pow(10, 2)))%10,
		certificate.Value,
		certificate.ValidatorID.Unwrap().Hex())
	if err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to insertCertificate")
	}

	if err := w.updateProfile(tx, certificate); err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateProfile")
	}

	if err := w.updateEntitiesByProfile(tx, certificate); err != nil {
		if err := tx.Rollback(); err != nil {
			w.logger.Error("transaction rollback failed", zap.Error(err))
		}

		return errors.Wrap(err, "failed to updateEntitiesByProfile")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction commit failed")
	}

	return nil
}

func (w *DWH) updateProfile(tx *sql.Tx, certificate *pb.Certificate) error {
	// Create a Profile entry if it doesn't exist yet.
	if _, err := w.getProfileInfo(w.ctx, certificate.OwnerID, false); err != nil {
		_, err = tx.Exec(w.commands["insertProfileUserID"], certificate.OwnerID.Unwrap().Hex(), 0, 0)
		if err != nil {
			return errors.Wrap(err, "failed to insertProfileUserID")
		}
	}

	// Update distinct Profile columns.
	switch certificate.Attribute {
	case CertificateName:
		_, err := tx.Exec(fmt.Sprintf(w.commands["updateProfile"], attributeToString[certificate.Attribute]),
			string(certificate.Value), certificate.OwnerID.Unwrap().Hex())
		if err != nil {
			return errors.Wrap(err, "failed to updateProfileName")
		}
	case CertificateCountry:
		_, err := tx.Exec(fmt.Sprintf(w.commands["updateProfile"], attributeToString[certificate.Attribute]),
			string(certificate.Value), certificate.OwnerID.Unwrap().Hex())
		if err != nil {
			return errors.Wrap(err, "failed to updateProfileCountry")
		}
	}

	// Update certificates blob.
	rows, err := tx.Query(w.commands["selectCertificates"], certificate.OwnerID.Unwrap().Hex())
	if err != nil {
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
		return errors.Wrap(err, "failed to marshal certificates")
	}

	_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], "Certificates"),
		certificateAttrsBytes, certificate.OwnerID.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Certificates)")
	}

	_, err = tx.Exec(fmt.Sprintf(w.commands["updateProfile"], "IdentityLevel"),
		maxIdentityLevel, certificate.OwnerID.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Level)")
	}

	return nil
}

func (w *DWH) updateEntitiesByProfile(tx *sql.Tx, certificate *pb.Certificate) error {
	profile, err := w.getProfileInfoTx(tx, certificate.OwnerID)
	if err != nil {
		return errors.Wrap(err, "failed to getProfileInfo")
	}

	_, err = tx.Exec(w.commands["updateOrders"],
		profile.IdentityLevel,
		profile.Name,
		profile.Country,
		profile.Certificates,
		profile.UserID.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateOrders")
	}

	_, err = tx.Exec(w.commands["updateDealsSupplier"], []byte(profile.Certificates), profile.UserID.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateDealsSupplier")
	}

	_, err = tx.Exec(w.commands["updateDealsConsumer"], []byte(profile.Certificates), profile.UserID.Unwrap().Hex())
	if err != nil {
		return errors.Wrap(err, "failed to updateDealsConsumer")
	}

	return nil
}

func (w *DWH) updateProfileStats(tx *sql.Tx, orderType pb.OrderType, authorID string, profile *pb.Profile, update int) error {
	var (
		cmd   string
		value int
	)
	if orderType == pb.OrderType_ASK {
		updateResult := int(profile.ActiveAsks) + update
		if updateResult < 0 {
			return errors.Errorf("updateProfileStats resulted in a negative Asks value (UserID: `%s`)", authorID)
		}
		cmd, value = fmt.Sprintf(w.commands["updateProfile"], "ActiveAsks"), updateResult
	} else {
		updateResult := int(profile.ActiveBids) + update
		if updateResult < 0 {
			return errors.Errorf("updateProfileStats resulted in a negative Bids value (UserID: `%s`)", authorID)
		}
		cmd, value = fmt.Sprintf(w.commands["updateProfile"], "ActiveBids"), updateResult
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

	bigID, err := pb.NewBigIntFromString(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ID)")
	}

	bigAskID, err := pb.NewBigIntFromString(askID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (askID)")
	}

	bigBidID, err := pb.NewBigIntFromString(bidID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (bidID)")
	}

	return &pb.DWHDeal{
		Deal: &pb.Deal{
			Id:             bigID,
			SupplierID:     pb.NewEthAddress(common.HexToAddress(supplierID)),
			ConsumerID:     pb.NewEthAddress(common.HexToAddress(consumerID)),
			MasterID:       pb.NewEthAddress(common.HexToAddress(masterID)),
			AskID:          bigAskID,
			BidID:          bigBidID,
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

	bigPrice, err := pb.NewBigIntFromString(price)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (Price)")
	}
	bigFrozenSum, err := pb.NewBigIntFromString(frozenSum)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (FrozenSum)")
	}
	bigID, err := pb.NewBigIntFromString(id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (ID)")
	}
	bigDealID, err := pb.NewBigIntFromString(dealID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewBigIntFromString (DealID)")
	}

	return &pb.DWHOrder{
		Order: &pb.Order{
			Id:             bigID,
			DealID:         bigDealID,
			OrderType:      pb.OrderType(orderType),
			OrderStatus:    pb.OrderStatus(status),
			AuthorID:       pb.NewEthAddress(common.HexToAddress(author)),
			CounterpartyID: pb.NewEthAddress(common.HexToAddress(counterAgent)),
			Duration:       duration,
			Price:          bigPrice,
			Netflags:       netflags,
			IdentityLevel:  pb.IdentityLevel(identityLevel),
			Blacklist:      blacklist,
			Tag:            tag,
			FrozenSum:      bigFrozenSum,
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

		CreatedTS:            &pb.Timestamp{Seconds: int64(createdTS)},
		CreatorIdentityLevel: creatorIdentityLevel,
		CreatorName:          creatorName,
		CreatorCountry:       creatorCountry,
		CreatorCertificates:  creatorCertificates,
	}, nil
}

func (w *DWH) decodeDealCondition(rows *sql.Rows) (*pb.DealCondition, error) {
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
		w.logger.Error("failed to scan DealCondition row", zap.Error(err))
		return nil, err
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
			OwnerID:       pb.NewEthAddress(common.HexToAddress(ownerID)),
			Attribute:     attribute,
			IdentityLevel: identityLevel,
			Value:         value,
			ValidatorID:   pb.NewEthAddress(common.HexToAddress(validatorID)),
		}, nil
	}
}

func (w *DWH) decodeProfile(rows *sql.Rows) (*pb.Profile, error) {
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
		w.logger.Error("failed to scan deal row", zap.Error(err))
		return nil, err
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

func (w *DWH) decodeValidator(rows *sql.Rows) (*pb.Validator, error) {
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
		MasterID:  pb.NewEthAddress(common.HexToAddress(masterID)),
		SlaveID:   pb.NewEthAddress(common.HexToAddress(slaveID)),
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

func (w *DWH) insertLastKnownBlockTS(blockNumber int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.db.Exec(w.commands["insertLastKnownBlock"], blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (w *DWH) updateDealConditionEndTime(tx *sql.Tx, dealID *pb.BigInt, eventTS uint64) error {
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
