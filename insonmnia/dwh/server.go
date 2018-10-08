package dwh

import (
	"context"
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/sonm-io/core/proto"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type DWH struct {
	logger      *zap.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cfg         *DWHConfig
	key         *ecdsa.PrivateKey
	cancel      context.CancelFunc
	grpc        *grpc.Server
	http        *rest.Server
	db          *sql.DB
	creds       credentials.TransportCredentials
	certRotator util.HitlessCertRotator
	blockchain  blockchain.API
	storage     *sqlStorage
	lastEvent   *blockchain.Event
	stats       *sonm.DWHStatsReply
	outOfSync   bool
}

func NewDWH(ctx context.Context, cfg *DWHConfig, key *ecdsa.PrivateKey) (*DWH, error) {
	ctx, cancel := context.WithCancel(ctx)
	w := &DWH{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		key:    key,
		logger: log.GetLogger(ctx),
	}
	return w, nil
}

func (m *DWH) Serve() error {
	m.logger.Info("starting with backend", zap.String("endpoint", m.cfg.Storage.Endpoint))
	var err error
	m.db, err = sql.Open("postgres", m.cfg.Storage.Endpoint)
	if err != nil {
		m.Stop()
		return err
	}

	bch, err := blockchain.NewAPI(m.ctx, blockchain.WithConfig(m.cfg.Blockchain))
	if err != nil {
		m.Stop()
		return fmt.Errorf("failed to create NewAPI: %v", err)
	}
	m.blockchain = bch

	numBenchmarks, err := m.blockchain.Market().GetNumBenchmarks(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to GetNumBenchmarks: %v", err)
	}

	m.storage = newPostgresStorage(numBenchmarks)

	wg := errgroup.Group{}
	wg.Go(m.serveGRPC)
	wg.Go(m.serveHTTP)
	wg.Go(m.monitorStatistics)
	wg.Go(m.monitorSync)

	return wg.Wait()
}

func (m *DWH) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stop()
}

func (m *DWH) stop() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.db != nil {
		m.db.Close()
	}
	if m.grpc != nil {
		m.grpc.Stop()
	}
	if m.http != nil {
		m.http.Close()
	}
}

func (m *DWH) serveGRPC() error {
	lis, err := func() (net.Listener, error) {
		m.mu.Lock()
		defer m.mu.Unlock()

		certRotator, TLSConfig, err := util.NewHitlessCertRotator(m.ctx, m.key)
		if err != nil {
			return nil, err
		}

		m.certRotator = certRotator
		m.creds = util.NewTLS(TLSConfig)
		m.grpc = xgrpc.NewServer(
			m.logger,
			xgrpc.Credentials(m.creds),
			xgrpc.DefaultTraceInterceptor(),
			xgrpc.UnaryServerInterceptor(m.unaryInterceptor),
		)
		sonm.RegisterDWHServer(m.grpc, m)
		grpc_prometheus.Register(m.grpc)

		lis, err := net.Listen("tcp", m.cfg.GRPCListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen on %s: %v", m.cfg.GRPCListenAddr, err)
		}

		return lis, nil
	}()
	if err != nil {
		return err
	}

	return m.grpc.Serve(lis)
}

func (m *DWH) serveHTTP() error {
	lis, err := func() (net.Listener, error) {
		m.mu.Lock()
		defer m.mu.Unlock()

		options := []rest.Option{rest.WithLog(m.logger)}
		lis, err := net.Listen("tcp", m.cfg.HTTPListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create http listener: %v", err)
		}

		srv := rest.NewServer(options...)

		err = srv.RegisterService((*sonm.DWHServer)(nil), m)
		if err != nil {
			return nil, fmt.Errorf("failed to RegisterService: %v", err)
		}
		m.http = srv

		return lis, err
	}()
	if err != nil {
		return err
	}

	return m.http.Serve(lis)
}

func (m *DWH) monitorNumBenchmarks() error {
	lastBlock, err := m.blockchain.Events().GetLastBlock(m.ctx)
	if err != nil {
		return err
	}

	filter := m.blockchain.Events().GetMarketFilter(big.NewInt(0).SetUint64(lastBlock))
	events, err := m.blockchain.Events().GetEvents(m.ctx, filter)
	if err != nil {
		return err
	}

	for {
		event, ok := <-events
		if !ok {
			return errors.New("events channel closed")
		}
		if _, ok := event.Data.(*blockchain.NumBenchmarksUpdatedData); ok {
			if m.storage, err = setupDB(m.ctx, m.db, m.blockchain); err != nil {
				return fmt.Errorf("failed to setupDB after NumBenchmarksUpdated event: %v", err)
			}

			if err := m.storage.CreateIndices(m.db); err != nil {
				return fmt.Errorf("failed to CreateIndices (onNumBenchmarksUpdated): %v", err)
			}
		}
	}
}

func (m *DWH) monitorStatistics() error {
	tk := util.NewImmediateTicker(time.Second)
	for {
		select {
		case <-tk.C:
			func() {
				conn := newSimpleConn(m.db)
				defer conn.Finish()

				if stats, err := m.storage.getStats(conn); err != nil {
					m.logger.Warn("failed to getStats", zap.Error(err))
				} else {
					m.mu.Lock()
					m.stats = stats
					m.mu.Unlock()
				}
			}()
		case <-m.ctx.Done():
			return errors.New("monitorStatistics: context cancelled")
		}
	}
}

func (m *DWH) monitorSync() error {
	const maxBacklog = 50
	tk := util.NewImmediateTicker(time.Minute)
	for {
		select {
		case <-tk.C:
			func() {
				conn := newSimpleConn(m.db)
				defer conn.Finish()

				lastEvent, err := m.storage.GetLastEvent(conn)
				if err != nil {
					m.logger.Warn("failed to get last event", zap.Error(err))
					return
				}

				lastBlock, err := m.blockchain.Events().GetLastBlock(m.ctx)
				if err != nil {
					m.logger.Warn("failed to get last block", zap.Error(err))
					return
				}

				m.mu.Lock()
				m.outOfSync = int64(lastBlock)-int64(lastEvent.BlockNumber) > maxBacklog
				m.mu.Unlock()
			}()
		case <-m.ctx.Done():
			return errors.New("monitorSync: context cancelled")
		}
	}
}

// unaryInterceptor RLocks DWH for all incoming requests. This is needed because some events (e.g.,
// NumBenchmarksUpdated) can alter `m.storage` state.
func (m *DWH) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.outOfSync {
		return nil, status.Error(codes.Internal, "DWH is out of sync, please wait")
	}
	return handler(ctx, req)
}

func (m *DWH) GetDeals(ctx context.Context, request *sonm.DealsRequest) (*sonm.DWHDealsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	deals, count, err := m.storage.GetDeals(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetDeals", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDeals")
	}

	return &sonm.DWHDealsReply{Deals: deals, Count: count}, nil
}

func (m *DWH) GetDealDetails(ctx context.Context, request *sonm.BigInt) (*sonm.DWHDeal, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetDealByID(conn, request.Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetDealDetails", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealDetails")
	}

	return out, nil
}

func (m *DWH) GetDealConditions(ctx context.Context, request *sonm.DealConditionsRequest) (*sonm.DealConditionsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	dealConditions, count, err := m.storage.GetDealConditions(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetDealConditions", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealConditions")
	}

	return &sonm.DealConditionsReply{Conditions: dealConditions, Count: count}, nil
}

func (m *DWH) GetOrders(ctx context.Context, request *sonm.OrdersRequest) (*sonm.DWHOrdersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	orders, count, err := m.storage.GetOrders(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetOrders", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetOrders")
	}

	return &sonm.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (m *DWH) GetMatchingOrders(ctx context.Context, request *sonm.MatchingOrdersRequest) (*sonm.DWHOrdersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	orders, count, err := m.storage.GetMatchingOrders(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetMatchingOrders", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetMatchingOrders")
	}

	return &sonm.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (m *DWH) GetOrderDetails(ctx context.Context, request *sonm.BigInt) (*sonm.DWHOrder, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetOrderByID(conn, request.Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetOrderDetails", zap.Error(err), zap.Any("request", *request))
		return nil, fmt.Errorf("failed to GetOrderDetails: %v", err)
	}

	return out, nil
}

func (m *DWH) GetProfiles(ctx context.Context, request *sonm.ProfilesRequest) (*sonm.ProfilesReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	profiles, count, err := m.storage.GetProfiles(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetProfiles", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfiles")
	}

	return &sonm.ProfilesReply{Profiles: profiles, Count: count}, nil
}

func (m *DWH) GetProfileInfo(ctx context.Context, request *sonm.EthID) (*sonm.Profile, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetProfileByID(conn, request.GetId().Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetProfileInfo", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfileInfo")
	}

	certs, err := m.storage.GetCertificates(conn, request.GetId().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("failed to GetCertificates: %v", err)
	}

	certsEncoded, err := json.Marshal(certs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s certificates: %v", request.GetId().Unwrap(), err)
	}

	out.Certificates = string(certsEncoded)

	return out, nil
}

func (m *DWH) GetBlacklist(ctx context.Context, request *sonm.BlacklistRequest) (*sonm.BlacklistReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetBlacklist(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetBlacklist", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (m *DWH) GetBlacklistsContainingUser(ctx context.Context, r *sonm.BlacklistRequest) (*sonm.BlacklistsContainingUserReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetBlacklistsContainingUser(conn, r)
	if err != nil {
		m.logger.Warn("failed to GetBlacklistsContainingUser", zap.Error(err), zap.Any("request", *r))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (m *DWH) GetValidators(ctx context.Context, request *sonm.ValidatorsRequest) (*sonm.ValidatorsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	validators, count, err := m.storage.GetValidators(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetValidators", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetValidators")
	}

	return &sonm.ValidatorsReply{Validators: validators, Count: count}, nil
}

func (m *DWH) GetDealChangeRequests(ctx context.Context, dealID *sonm.BigInt) (*sonm.DealChangeRequestsReply, error) {
	return m.GetChangeRequests(ctx, &sonm.ChangeRequestsRequest{
		DealID:     dealID,
		OnlyActive: true,
	})
}

func (m *DWH) GetChangeRequests(ctx context.Context, request *sonm.ChangeRequestsRequest) (*sonm.DealChangeRequestsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetDealChangeRequestsByDealID(conn, request.DealID.Unwrap(), request.OnlyActive)
	if err != nil {
		m.logger.Error("failed to GetDealChangeRequests", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealChangeRequests")
	}

	return &sonm.DealChangeRequestsReply{
		Requests: out,
	}, nil
}

func (m *DWH) GetWorkers(ctx context.Context, request *sonm.WorkersRequest) (*sonm.WorkersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	workers, count, err := m.storage.GetWorkers(conn, request)
	if err != nil {
		m.logger.Error("failed to GetWorkers", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetWorkers")
	}

	return &sonm.WorkersReply{Workers: workers, Count: count}, nil
}

func (m *DWH) GetStats(ctx context.Context, request *sonm.Empty) (*sonm.DWHStatsReply, error) {
	return m.stats, nil
}

func (m *DWH) GetOrdersByIDs(ctx context.Context, request *sonm.OrdersByIDsRequest) (*sonm.DWHOrdersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	orders, count, err := m.storage.GetOrdersByIDs(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetOrdersByIDs", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetWorkers")
	}

	return &sonm.DWHOrdersReply{
		Orders: orders,
		Count:  count,
	}, nil
}
