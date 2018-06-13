package dwh

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
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
	mu             sync.RWMutex
	ctx            context.Context
	cfg            *Config
	key            *ecdsa.PrivateKey
	cancel         context.CancelFunc
	grpc           *grpc.Server
	http           *rest.Server
	logger         *zap.Logger
	db             *sql.DB
	creds          credentials.TransportCredentials
	certRotator    util.HitlessCertRotator
	blockchain     blockchain.API
	storage        storage
	numBenchmarks  uint64
	lastKnownBlock uint64
}

func NewDWH(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (*DWH, error) {
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
	m.logger.Info("starting with backend", zap.String("backend", m.cfg.Storage.Backend),
		zap.String("endpoint", m.cfg.Storage.Endpoint))
	var err error
	m.db, err = sql.Open(m.cfg.Storage.Backend, m.cfg.Storage.Endpoint)
	if err != nil {
		m.Stop()
		return err
	}

	bch, err := blockchain.NewAPI(blockchain.WithConfig(m.cfg.Blockchain))
	if err != nil {
		m.Stop()
		return errors.Wrap(err, "failed to create NewAPI")
	}
	m.blockchain = bch

	if err := m.setupDBts(); err != nil {
		m.Stop()
		return errors.WithMessage(err, "failed to setupDBts")
	}

	go m.monitorBlockchain()
	if m.cfg.ColdStart {
		if err := m.coldStart(); err != nil {
			m.Stop()
			return errors.Wrap(err, "failed to coldStart")
		}
	}

	wg := errgroup.Group{}
	wg.Go(m.serveGRPC)
	wg.Go(m.serveHTTP)

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

func (m *DWH) setupDBts() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	numBenchmarks, err := m.blockchain.Market().GetNumBenchmarks(m.ctx)
	if err != nil {
		m.stop()
		return errors.Wrap(err, "failed to GetNumBenchmarks")
	}
	if numBenchmarks >= NumMaxBenchmarks {
		m.stop()
		return errors.New("market number of benchmarks is greater than NumMaxBenchmarks")
	}

	var storage *sqlStorage
	switch m.cfg.Storage.Backend {
	case "sqlite3":
		_, err := m.db.Exec(`PRAGMA foreign_keys=ON`)
		if err != nil {
			return errors.Wrapf(err, "failed to enable foreign key support (%s)", m.cfg.Storage.Backend)
		}
		storage = newSQLiteStorage(numBenchmarks)
	case "postgres":
		storage = newPostgresStorage(numBenchmarks)
	default:
		return errors.Errorf("unsupported backend: %s", m.cfg.Storage.Backend)
	}

	if err := storage.Setup(m.db); err != nil {
		return errors.Wrap(err, "failed to setup storage")
	}

	m.numBenchmarks = numBenchmarks
	m.storage = storage
	return nil
}

func (m *DWH) serveGRPC() error {
	m.mu.Lock()
	certRotator, TLSConfig, err := util.NewHitlessCertRotator(m.ctx, m.key)
	if err != nil {
		m.mu.Unlock()
		return err
	}

	m.certRotator = certRotator
	m.creds = util.NewTLS(TLSConfig)
	m.grpc = xgrpc.NewServer(m.logger, xgrpc.Credentials(m.creds), xgrpc.DefaultTraceInterceptor())
	pb.RegisterDWHServer(m.grpc, m)
	grpc_prometheus.Register(m.grpc)

	lis, err := net.Listen("tcp", m.cfg.GRPCListenAddr)
	if err != nil {
		m.mu.Unlock()
		return errors.Wrapf(err, "failed to listen on %s", m.cfg.GRPCListenAddr)
	}

	m.mu.Unlock()
	return m.grpc.Serve(lis)
}

func (m *DWH) serveHTTP() error {
	m.mu.Lock()
	options := []rest.Option{rest.WithContext(m.ctx)}
	lis, err := net.Listen("tcp", m.cfg.HTTPListenAddr)
	if err != nil {
		m.mu.Unlock()
		return errors.WithMessage(err, "failed to create http listener")
	}

	options = append(options, rest.WithListener(lis))
	srv, err := rest.NewServer(options...)
	if err != nil {
		m.mu.Unlock()
		return errors.WithMessage(err, "failed to create rest server")
	}

	err = srv.RegisterService((*pb.DWHServer)(nil), m)
	if err != nil {
		m.mu.Unlock()
		return errors.WithMessage(err, "failed to RegisterService")
	}

	m.http = srv
	m.mu.Unlock()
	return srv.Serve()
}

func (m *DWH) GetDeals(ctx context.Context, request *pb.DealsRequest) (*pb.DWHDealsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	deals, count, err := m.storage.GetDeals(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetDeals", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDeals")
	}

	return &pb.DWHDealsReply{Deals: deals, Count: count}, nil
}

func (m *DWH) GetDealDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHDeal, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetDealByID(conn, request.Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetDealDetails", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealDetails")
	}

	return out, nil
}

func (m *DWH) GetDealConditions(ctx context.Context, request *pb.DealConditionsRequest) (*pb.DealConditionsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	dealConditions, count, err := m.storage.GetDealConditions(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetDealConditions", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealConditions")
	}

	return &pb.DealConditionsReply{Conditions: dealConditions, Count: count}, nil
}

func (m *DWH) GetOrders(ctx context.Context, request *pb.OrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	orders, count, err := m.storage.GetOrders(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetOrders", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (m *DWH) GetMatchingOrders(ctx context.Context, request *pb.MatchingOrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	orders, count, err := m.storage.GetMatchingOrders(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetMatchingOrders", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetMatchingOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (m *DWH) GetOrderDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHOrder, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetOrderByID(conn, request.Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetOrderDetails", util.LaconicError(err), zap.Any("request", *request))
		return nil, errors.Wrap(err, "failed to GetOrderDetails")
	}

	return out, nil
}

func (m *DWH) GetProfiles(ctx context.Context, request *pb.ProfilesRequest) (*pb.ProfilesReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	profiles, count, err := m.storage.GetProfiles(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetProfiles", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfiles")
	}

	return &pb.ProfilesReply{Profiles: profiles, Count: count}, nil
}

func (m *DWH) GetProfileInfo(ctx context.Context, request *pb.EthID) (*pb.Profile, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetProfileByID(conn, request.GetId().Unwrap())
	if err != nil {
		m.logger.Warn("failed to GetProfileInfo", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfileInfo")
	}

	return out, nil
}

func (m *DWH) GetBlacklist(ctx context.Context, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetBlacklist(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetBlacklist", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (m *DWH) GetBlacklistsContainingUser(ctx context.Context, r *pb.BlacklistRequest) (*pb.BlacklistsContainingUserReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.storage.GetBlacklistsContainingUser(conn, r)
	if err != nil {
		m.logger.Warn("failed to GetBlacklistsContainingUser", util.LaconicError(err), zap.Any("request", *r))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (m *DWH) GetValidators(ctx context.Context, request *pb.ValidatorsRequest) (*pb.ValidatorsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	validators, count, err := m.storage.GetValidators(conn, request)
	if err != nil {
		m.logger.Warn("failed to GetValidators", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetValidators")
	}

	return &pb.ValidatorsReply{Validators: validators, Count: count}, nil
}

func (m *DWH) GetDealChangeRequests(ctx context.Context, request *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	out, err := m.getDealChangeRequests(conn, request)
	if err != nil {
		m.logger.Error("failed to GetDealChangeRequests", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealChangeRequests")
	}

	return &pb.DealChangeRequestsReply{Requests: out}, nil
}

func (m *DWH) getDealChangeRequests(conn queryConn, request *pb.BigInt) ([]*pb.DealChangeRequest, error) {
	return m.storage.GetDealChangeRequestsByDealID(conn, request.Unwrap())
}

func (m *DWH) GetWorkers(ctx context.Context, request *pb.WorkersRequest) (*pb.WorkersReply, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	workers, count, err := m.storage.GetWorkers(conn, request)
	if err != nil {
		m.logger.Error("failed to GetWorkers", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetWorkers")
	}

	return &pb.WorkersReply{Workers: workers, Count: count}, nil
}

func (m *DWH) monitorBlockchain() error {
	m.logger.Info("starting monitoring")

	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("context cancelled (monitorBlockchain)")
			return nil
		default:
			if err := m.watchMarketEvents(); err != nil {
				m.logger.Warn("failed to watch market events, retrying", util.LaconicError(err))
			}
		}
	}
}

func (m *DWH) watchMarketEvents() error {
	var err error
	m.lastKnownBlock, err = m.getLastKnownBlock()
	if err != nil {
		if err := m.insertLastKnownBlock(0); err != nil {
			return err
		}
		m.lastKnownBlock = 0
	}

	m.logger.Info("starting from block", zap.Uint64("block_number", m.lastKnownBlock))
	events, err := m.blockchain.Events().GetEvents(m.ctx, big.NewInt(0).SetUint64(m.lastKnownBlock))
	if err != nil {
		return err
	}

	var (
		eventsCount int
		dispatcher  = newEventDispatcher(m.logger)
		tk          = time.NewTicker(time.Millisecond * 500)
	)
	defer tk.Stop()

	// Store events by their type, run events of each type in parallel after a timeout
	// or after a certain number of events is accumulated.
	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("context cancelled (watchMarketEvents)")
			return nil
		case <-tk.C:
			m.processEvents(dispatcher)
			eventsCount, dispatcher = 0, newEventDispatcher(m.logger)
		case event, ok := <-events:
			if !ok {
				return errors.New("events channel closed")
			}
			m.processBlockBoundary(event)
			dispatcher.Add(event)
			eventsCount++
			if eventsCount >= m.cfg.NumWorkers {
				m.processEvents(dispatcher)
				eventsCount, dispatcher = 0, newEventDispatcher(m.logger)
			}
		}
	}
}

func (m *DWH) processEvents(dispatcher *eventsDispatcher) {
	m.processEventsGroup(dispatcher.WorkersAnnounced)
	m.processEventsGroup(dispatcher.WorkersConfirmed)
	m.processEventsGroup(dispatcher.ValidatorsCreated)
	m.processEventsGroup(dispatcher.CertificatesCreated)
	m.processEventsGroup(dispatcher.OrdersOpened)
	m.processEventsGroup(dispatcher.DealsOpened)
	m.processEventsGroup(dispatcher.DealChangeRequestsSent)
	m.processEventsGroup(dispatcher.Billed)
	m.processEventsGroup(dispatcher.DealChangeRequestsUpdated)
	m.processEventsGroup(dispatcher.OrdersClosed)
	m.processEventsGroup(dispatcher.DealsClosed)
	m.processEventsGroup(dispatcher.ValidatorsDeleted)
	m.processEventsGroup(dispatcher.AddedToBlacklist)
	m.processEventsGroup(dispatcher.RemovedFromBlacklist)
	m.processEventsGroup(dispatcher.WorkersRemoved)
	m.processEventsGroup(dispatcher.Other)
}

func (m *DWH) processEventsGroup(events []*blockchain.Event) {
	wg := &sync.WaitGroup{}
	for _, event := range events {
		wg.Add(1)
		go func(wg *sync.WaitGroup, event *blockchain.Event) {
			defer wg.Done()
			var (
				err        error
				numRetries = 60
			)
			for numRetries > 0 {
				if err = m.processEvent(event); err != nil {
					m.logger.Warn("failed to processEvent, retrying", util.LaconicError(err),
						zap.Uint64("block_number", event.BlockNumber),
						zap.String("event_type", reflect.TypeOf(event.Data).String()),
						zap.Any("event_data", event.Data))
				} else {
					m.logger.Debug("processed event", zap.Uint64("block_number", event.BlockNumber),
						zap.String("event_type", reflect.TypeOf(event.Data).String()),
						zap.Any("event_data", event.Data))
					return
				}
				numRetries--
				time.Sleep(time.Second)
			}
			m.logger.Warn("failed to processEvent, STATE IS INCONSISTENT", util.LaconicError(err),
				zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).String()),
				zap.Any("event_data", event.Data))
		}(wg, event)
	}
	wg.Wait()
}

func (m *DWH) processEvent(event *blockchain.Event) error {
	switch value := event.Data.(type) {
	case *blockchain.DealOpenedData:
		return m.onDealOpened(value.ID)
	case *blockchain.DealUpdatedData:
		return m.onDealUpdated(value.ID)
	case *blockchain.OrderPlacedData:
		return m.onOrderPlaced(event.TS, value.ID)
	case *blockchain.OrderUpdatedData:
		return m.onOrderUpdated(value.ID)
	case *blockchain.DealChangeRequestSentData:
		return m.onDealChangeRequestSent(event.TS, value.ID)
	case *blockchain.DealChangeRequestUpdatedData:
		return m.onDealChangeRequestUpdated(event.TS, value.ID)
	case *blockchain.BilledData:
		return m.onBilled(event.TS, value.DealID, value.PaidAmount)
	case *blockchain.WorkerAnnouncedData:
		return m.onWorkerAnnounced(value.MasterID, value.WorkerID)
	case *blockchain.WorkerConfirmedData:
		return m.onWorkerConfirmed(value.MasterID, value.WorkerID)
	case *blockchain.WorkerRemovedData:
		return m.onWorkerRemoved(value.MasterID, value.WorkerID)
	case *blockchain.AddedToBlacklistData:
		return m.onAddedToBlacklist(value.AdderID, value.AddeeID)
	case *blockchain.RemovedFromBlacklistData:
		m.onRemovedFromBlacklist(value.RemoverID, value.RemoveeID)
	case *blockchain.ValidatorCreatedData:
		return m.onValidatorCreated(value.ID)
	case *blockchain.ValidatorDeletedData:
		return m.onValidatorDeleted(value.ID)
	case *blockchain.CertificateCreatedData:
		return m.onCertificateCreated(value.ID)
	}

	return nil
}

func (m *DWH) onDealOpened(dealID *big.Int) error {
	deal, err := m.blockchain.Market().GetDealInfo(m.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		if err := m.storage.StoreStaleID(newSimpleConn(m.db), dealID, "Deal"); err != nil {
			return errors.Wrap(err, "failed to StoreStaleID")
		}
		m.logger.Debug("skipping inactive deal", zap.String("deal_id", dealID.String()))
		return nil
	}

	if err := m.checkBenchmarks(deal.Benchmarks); err != nil {
		return err
	}

	err = m.storage.InsertDeal(conn, deal)
	if err != nil {
		return errors.Wrapf(err, "failed to insertDeal")
	}

	err = m.storage.InsertDealCondition(conn,
		&pb.DealCondition{
			SupplierID:  deal.SupplierID,
			ConsumerID:  deal.ConsumerID,
			MasterID:    deal.MasterID,
			Duration:    deal.Duration,
			Price:       deal.Price,
			StartTime:   deal.StartTime,
			EndTime:     &pb.Timestamp{},
			TotalPayout: deal.TotalPayout,
			DealID:      deal.Id,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "onDealOpened: failed to insertDealCondition")
	}

	return nil
}

func (m *DWH) onDealUpdated(dealID *big.Int) error {
	deal, err := m.blockchain.Market().GetDealInfo(m.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale:
	if ok, err := m.storage.CheckStaleID(conn, dealID, "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			m.removeStaleEntityID(dealID, "Deal")
			return nil
		}
	}

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		err = m.storage.DeleteDeal(conn, deal.Id.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to delete deal (possibly old log entry)")
		}

		if err := m.storage.DeleteOrder(conn, deal.AskID.Unwrap()); err != nil {
			return errors.Wrap(err, "failed to deleteOrder")
		}
		if err := m.storage.DeleteOrder(conn, deal.BidID.Unwrap()); err != nil {
			return errors.Wrap(err, "failed to deleteOrder")
		}

		return nil
	}

	if err := m.storage.UpdateDeal(conn, deal); err != nil {
		return errors.Wrapf(err, "failed to UpdateDeal")
	}

	return nil
}

func (m *DWH) onDealChangeRequestSent(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := m.blockchain.Market().GetDealChangeRequestInfo(m.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := m.storage.CheckStaleID(conn, changeRequest.DealID.Unwrap(), "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			m.logger.Debug("skipping DealChangeRequestSent event for inactive deal")
			return nil
		}
	}

	if changeRequest.Status != pb.ChangeRequestStatus_REQUEST_CREATED {
		m.logger.Info("onDealChangeRequest event points to DealChangeRequest with .Status != Created",
			zap.String("actual_status", pb.ChangeRequestStatus_name[int32(changeRequest.Status)]))
		return nil
	}

	// Sanity check: if more than 1 CR of one type is created for a Deal, we delete old CRs.
	expiredChangeRequests, err := m.storage.GetDealChangeRequests(conn, changeRequest)
	if err != nil {
		return errors.New("failed to get (possibly) expired DealChangeRequests")
	}

	for _, expiredChangeRequest := range expiredChangeRequests {
		err := m.storage.DeleteDealChangeRequest(conn, expiredChangeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to deleteDealChangeRequest")
		} else {
			m.logger.Warn("deleted expired DealChangeRequest",
				zap.String("id", expiredChangeRequest.Id.Unwrap().String()))
		}
	}

	changeRequest.CreatedTS = &pb.Timestamp{Seconds: int64(eventTS)}
	if err := m.storage.InsertDealChangeRequest(conn, changeRequest); err != nil {
		return errors.Wrapf(err, "failed to InsertDealChangeRequest (%s)", changeRequest.Id.Unwrap().String())
	}

	return err
}

func (m *DWH) onDealChangeRequestUpdated(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := m.blockchain.Market().GetDealChangeRequestInfo(m.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := m.storage.CheckStaleID(conn, changeRequest.DealID.Unwrap(), "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			m.logger.Debug("skipping DealChangeRequestUpdated event for inactive deal")
			return nil
		}
	}

	switch changeRequest.Status {
	case pb.ChangeRequestStatus_REQUEST_REJECTED:
		err := m.storage.UpdateDealChangeRequest(conn, changeRequest)
		if err != nil {
			return errors.Wrapf(err, "failed to update DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	case pb.ChangeRequestStatus_REQUEST_ACCEPTED:
		deal, err := m.storage.GetDealByID(conn, changeRequest.DealID.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to storage.GetDealByID")
		}

		deal.Deal.Duration = changeRequest.Duration
		deal.Deal.Price = changeRequest.Price
		if err := m.storage.UpdateDeal(conn, deal.Deal); err != nil {
			return errors.WithMessage(err, "failed to UpdateDeal")
		}

		if err := m.updateDealConditionEndTime(conn, deal.GetDeal().Id, eventTS); err != nil {
			return errors.Wrap(err, "failed to updateDealConditionEndTime")
		}

		err = m.storage.InsertDealCondition(conn,
			&pb.DealCondition{
				SupplierID:  deal.GetDeal().SupplierID,
				ConsumerID:  deal.GetDeal().ConsumerID,
				MasterID:    deal.GetDeal().MasterID,
				Duration:    changeRequest.Duration,
				Price:       changeRequest.Price,
				StartTime:   &pb.Timestamp{Seconds: int64(eventTS)},
				EndTime:     &pb.Timestamp{},
				TotalPayout: pb.NewBigIntFromInt(0),
				DealID:      deal.GetDeal().Id,
			},
		)
		if err != nil {
			return errors.Wrap(err, "failed to insertDealCondition")
		}

		err = m.storage.DeleteDealChangeRequest(conn, changeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	default:
		err := m.storage.DeleteDealChangeRequest(conn, changeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	}

	return nil
}

func (m *DWH) onBilled(eventTS uint64, dealID, payedAmount *big.Int) error {
	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := m.storage.CheckStaleID(conn, dealID, "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			m.logger.Debug("skipping Billed event for inactive deal")
			return nil
		}
	}

	if err := m.updateDealPayout(conn, dealID, payedAmount, eventTS); err != nil {
		return errors.Wrap(err, "failed to updateDealPayout")
	}

	dealConditions, _, err := m.storage.GetDealConditions(conn, &pb.DealConditionsRequest{DealID: pb.NewBigInt(dealID)})
	if err != nil {
		return errors.Wrap(err, "failed to GetDealConditions (last)")
	}

	if len(dealConditions) < 1 {
		return errors.Errorf("no deal conditions found for deal `%s`", dealID.String())
	}

	err = m.storage.UpdateDealConditionPayout(conn, dealConditions[0].Id,
		big.NewInt(0).Add(dealConditions[0].TotalPayout.Unwrap(), payedAmount))
	if err != nil {
		return errors.Wrap(err, "failed to UpdateDealConditionPayout")
	}

	if err != nil {
		return errors.Wrap(err, "insertDealPayment failed")
	}

	return nil
}

func (m *DWH) updateDealPayout(conn queryConn, dealID, payedAmount *big.Int, billTS uint64) error {
	deal, err := m.storage.GetDealByID(conn, dealID)
	if err != nil {
		return errors.Wrap(err, "failed to GetDealByID")
	}

	newDealTotalPayout := big.NewInt(0).Add(deal.Deal.TotalPayout.Unwrap(), payedAmount)
	err = m.storage.UpdateDealPayout(conn, dealID, newDealTotalPayout, billTS)
	if err != nil {
		return errors.Wrap(err, "failed to updateDealPayout")
	}

	return nil
}

func (m *DWH) onOrderPlaced(eventTS uint64, orderID *big.Int) error {
	order, err := m.blockchain.Market().GetOrderInfo(m.ctx, orderID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetOrderInfo")
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if order.OrderStatus == pb.OrderStatus_ORDER_INACTIVE && order.DealID.IsZero() {
		if err := m.storage.StoreStaleID(conn, orderID, "Order"); err != nil {
			return errors.Wrap(err, "failed to StoreStaleID")
		}
		m.logger.Debug("skipping inactive order", zap.String("order_id", orderID.String()))
		return nil
	}

	var userID common.Address
	if order.OrderType == pb.OrderType_ASK {
		// For Ask orders, try to get this Author's masterID, use AuthorID if not found.
		userID, err = m.storage.GetMasterByWorker(conn, order.GetAuthorID().Unwrap())
		if err != nil {
			m.logger.Warn("failed to GetMasterByWorker", util.LaconicError(err),
				zap.String("author_id", order.GetAuthorID().Unwrap().Hex()))
			userID = order.GetAuthorID().Unwrap()
		}
	} else {
		userID = order.GetAuthorID().Unwrap()
	}

	profile, err := m.storage.GetProfileByID(conn, userID)
	if err != nil {
		certificates, _ := json.Marshal([]*pb.Certificate{})
		profile = &pb.Profile{UserID: order.AuthorID, Certificates: string(certificates)}
	} else {
		if order.OrderStatus == pb.OrderStatus_ORDER_ACTIVE {
			if err := m.updateProfileStats(conn, order.OrderType, userID, 1); err != nil {
				return errors.Wrap(err, "failed to updateProfileStats")
			}
		}
	}

	if order.DealID == nil {
		order.DealID = pb.NewBigIntFromInt(0)
	}

	if err := m.checkBenchmarks(order.Benchmarks); err != nil {
		return err
	}

	err = m.storage.InsertOrder(conn, &pb.DWHOrder{
		CreatedTS:            &pb.Timestamp{Seconds: int64(eventTS)},
		CreatorIdentityLevel: profile.IdentityLevel,
		CreatorName:          profile.Name,
		CreatorCountry:       profile.Country,
		CreatorCertificates:  []byte(profile.Certificates),
		MasterID:             pb.NewEthAddress(userID),
		Order: &pb.Order{
			Id:             order.Id,
			DealID:         order.DealID,
			OrderType:      order.OrderType,
			OrderStatus:    order.OrderStatus,
			AuthorID:       order.AuthorID,
			CounterpartyID: order.CounterpartyID,
			Duration:       order.Duration,
			Price:          order.Price,
			Netflags:       order.Netflags,
			IdentityLevel:  order.IdentityLevel,
			Blacklist:      order.Blacklist,
			Tag:            order.Tag,
			FrozenSum:      order.FrozenSum,
			Benchmarks:     order.Benchmarks,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "failed to insertOrder")
	}

	return nil
}

func (m *DWH) onOrderUpdated(orderID *big.Int) error {
	marketOrder, err := m.blockchain.Market().GetOrderInfo(m.ctx, orderID)
	if err != nil {
		return errors.Wrap(err, "failed to GetOrderInfo")
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If the order was known to be inactive, delete it from the list of inactive entities
	// and skip.
	if ok, err := m.storage.CheckStaleID(conn, orderID, "Order"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			m.removeStaleEntityID(orderID, "Order")
			return nil
		}
	}

	// A situation is possible when user places an Ask order without specifying her `MasterID` (and we take
	// `AuthorID` for `MasterID`), and afterwards the user *does* specify her master. To avoid inconsistency,
	// we always use the user ID that was chosen in `onOrderPlaced` (i.e., the one that is already stored in DB).
	dwhOrder, err := m.storage.GetOrderByID(conn, marketOrder.GetId().Unwrap())
	if err != nil {
		return errors.Wrap(err, "failed to GetOrderByID")
	}

	var userID common.Address
	if marketOrder.OrderType == pb.OrderType_ASK {
		userID = dwhOrder.GetMasterID().Unwrap()
	} else {
		userID = marketOrder.GetAuthorID().Unwrap()
	}

	// If order was updated, but no deal is associated with it, delete the order.
	if marketOrder.DealID.IsZero() {
		if err := m.storage.DeleteOrder(conn, orderID); err != nil {
			m.logger.Info("failed to delete Order (possibly old log entry)", util.LaconicError(err),
				zap.String("order_id", orderID.String()))
		}
	} else {
		// Otherwise update order status.
		err := m.storage.UpdateOrderStatus(conn, marketOrder.Id.Unwrap(), marketOrder.OrderStatus)
		if err != nil {
			return errors.Wrap(err, "failed to updateOrderStatus (possibly old log entry)")
		}
	}

	if dwhOrder.GetOrder().OrderStatus == pb.OrderStatus_ORDER_ACTIVE {
		if err := m.updateProfileStats(conn, marketOrder.OrderType, userID, -1); err != nil {
			return errors.Wrapf(err, "failed to updateProfileStats (AuthorID: `%s`)", marketOrder.AuthorID.Unwrap().String())
		}
	}

	return nil
}

func (m *DWH) onWorkerAnnounced(masterID, slaveID common.Address) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if ok, err := m.storage.CheckWorkerExists(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "failed to CheckWorker")
	} else {
		if ok {
			// Worker already exists, skipping.
			return nil
		}
	}

	if err := m.storage.InsertWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerAnnounced failed")
	}

	return nil
}

func (m *DWH) onWorkerConfirmed(masterID, slaveID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.UpdateWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerConfirmed failed")
	}

	return nil
}

func (m *DWH) onWorkerRemoved(masterID, slaveID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.DeleteWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerRemoved failed")
	}

	return nil
}

func (m *DWH) onAddedToBlacklist(adderID, addeeID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertBlacklistEntry(conn, adderID, addeeID); err != nil {
		return errors.Wrap(err, "onAddedToBlacklist failed")
	}

	return nil
}

func (m *DWH) onRemovedFromBlacklist(removerID, removeeID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.DeleteBlacklistEntry(conn, removerID, removeeID); err != nil {
		return errors.Wrap(err, "onRemovedFromBlacklist failed")
	}

	return nil
}

func (m *DWH) onValidatorCreated(validatorID common.Address) error {
	validator, err := m.blockchain.ProfileRegistry().GetValidator(m.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertOrUpdateValidator(conn, validator); err != nil {
		return errors.Wrap(err, "failed to insertValidator")
	}

	return nil
}

func (m *DWH) onValidatorDeleted(validatorID common.Address) error {
	validator, err := m.blockchain.ProfileRegistry().GetValidator(m.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.UpdateValidator(conn, validator); err != nil {
		return errors.Wrap(err, "failed to InsertOrUpdateValidator")
	}

	return nil
}

func (m *DWH) onCertificateCreated(certificateID *big.Int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	certificate, err := m.blockchain.ProfileRegistry().GetCertificate(m.ctx, certificateID)
	if err != nil {
		return errors.Wrap(err, "failed to GetCertificate")
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if err = m.storage.InsertCertificate(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to insertCertificate")
	}

	if err := m.updateProfile(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to updateProfile")
	}

	if err := m.updateEntitiesByProfile(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to updateEntitiesByProfile")
	}

	return nil
}

func (m *DWH) updateProfile(conn queryConn, certificate *pb.Certificate) error {
	_, activeAsks, err := m.storage.GetOrders(conn, &pb.OrdersRequest{
		Type:      pb.OrderType_ASK,
		MasterID:  certificate.OwnerID,
		WithCount: true})
	if err != nil {
		return errors.WithMessage(err, "failed to get active ASKs count")
	}

	_, activeBids, err := m.storage.GetOrders(conn, &pb.OrdersRequest{
		Type:      pb.OrderType_BID,
		MasterID:  certificate.OwnerID,
		WithCount: true})
	if err != nil {
		return errors.WithMessage(err, "failed to get active BIDs count")
	}

	certBytes, _ := json.Marshal([]*pb.Certificate{})
	err = m.storage.InsertProfileUserID(conn, &pb.Profile{
		UserID:       certificate.OwnerID,
		Certificates: string(certBytes),
		ActiveAsks:   activeAsks,
		ActiveBids:   activeBids,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insertProfileUserID")
	}

	// Update distinct Profile columns.
	switch certificate.Attribute {
	case CertificateName, CertificateCountry:
		err := m.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), attributeToString[certificate.Attribute],
			string(certificate.Value))
		if err != nil {
			return errors.Wrapf(err, "failed to UpdateProfile (%s)", attributeToString[certificate.Attribute])
		}
	}

	// Update certificates blob.
	certificates, err := m.storage.GetCertificates(conn, certificate.OwnerID.Unwrap())
	if err != nil {
		return errors.Wrap(err, "failed to GetCertificates")
	}

	certificateAttrsBytes, err := json.Marshal(certificates)
	if err != nil {
		return errors.Wrap(err, "failed to marshal certificates")
	}

	var maxIdentityLevel uint64
	for _, certificate := range certificates {
		if certificate.IdentityLevel > maxIdentityLevel {
			maxIdentityLevel = certificate.IdentityLevel
		}
	}

	err = m.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), "Certificates", certificateAttrsBytes)
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Certificates)")
	}

	err = m.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), "IdentityLevel", maxIdentityLevel)
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Level)")
	}

	return nil
}

func (m *DWH) updateEntitiesByProfile(conn queryConn, certificate *pb.Certificate) error {
	profile, err := m.storage.GetProfileByID(conn, certificate.OwnerID.Unwrap())
	if err != nil {
		return errors.Wrap(err, "failed to getProfileInfo")
	}

	if err := m.storage.UpdateOrders(conn, profile); err != nil {
		return errors.Wrap(err, "failed to updateOrders")
	}

	if err = m.storage.UpdateDealsSupplier(conn, profile); err != nil {
		return errors.Wrap(err, "failed to updateDealsSupplier")
	}

	err = m.storage.UpdateDealsConsumer(conn, profile)
	if err != nil {
		return errors.Wrap(err, "failed to updateDealsConsumer")
	}

	return nil
}

func (m *DWH) updateProfileStats(conn queryConn, orderType pb.OrderType, userID common.Address, update int) error {
	var field string
	if orderType == pb.OrderType_ASK {
		field = "ActiveAsks"
	} else {
		field = "ActiveBids"
	}

	if err := m.storage.UpdateProfileStats(conn, userID, field, update); err != nil {
		return errors.Wrap(err, "failed to UpdateProfileStats")
	}

	return nil
}

// coldStart waits till last seen block number gets to `w.cfg.ColdStart.UpToBlock` and then tries to create indices.
func (m *DWH) coldStart() error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	targetBlock, err := m.blockchain.Events().GetLastBlock(m.ctx)
	if err != nil {
		return errors.WithMessage(err, "failed to GetLastBlock")
	}
	var retries = 5
	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("stopped coldStart routine")
			return nil
		case <-ticker.C:
			targetBlockReached, err := m.maybeCreateIndices(targetBlock)
			if err != nil {
				if retries == 0 {
					m.logger.Warn("failed to CreateIndices, exiting")
					return err
				}
				retries--
				m.logger.Warn("failed to CreateIndices, retrying", zap.Int("retries_left", retries))
			}
			if targetBlockReached {
				m.logger.Info("CreateIndices success")
				return nil
			}
		}
	}
}

func (m *DWH) maybeCreateIndices(targetBlock uint64) (targetBlockReached bool, err error) {
	lastBlock, err := m.getLastKnownBlock()
	if err != nil {
		return false, err
	}

	m.logger.Info("current block (waiting to CreateIndices)", zap.Uint64("block_number", lastBlock))
	if lastBlock >= targetBlock {
		if err := m.storage.CreateIndices(m.db); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (m *DWH) getLastKnownBlock() (uint64, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	return m.storage.GetLastKnownBlock(conn)
}

func (m *DWH) updateLastKnownBlock(blockNumber int64) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.UpdateLastKnownBlock(conn, blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (m *DWH) insertLastKnownBlock(blockNumber int64) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertLastKnownBlock(conn, blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (m *DWH) updateDealConditionEndTime(conn queryConn, dealID *pb.BigInt, eventTS uint64) error {
	dealConditions, _, err := m.storage.GetDealConditions(conn, &pb.DealConditionsRequest{DealID: dealID})
	if err != nil {
		return errors.Wrap(err, "failed to getDealConditions")
	}

	dealCondition := dealConditions[0]
	if err := m.storage.UpdateDealConditionEndTime(conn, dealCondition.Id, eventTS); err != nil {
		return errors.Wrap(err, "failed to update DealCondition")
	}

	return nil
}

func (m *DWH) checkBenchmarks(benches *pb.Benchmarks) error {
	if uint64(len(benches.Values)) != m.numBenchmarks {
		return errors.Errorf("expected %d benchmarks, got %d", m.numBenchmarks, len(benches.Values))
	}

	for idx, bench := range benches.Values {
		if bench >= MaxBenchmark {
			return errors.Errorf("benchmark %d is greater that %d", idx, MaxBenchmark)
		}
	}

	return nil
}

func (m *DWH) removeStaleEntityID(id *big.Int, entity string) error {
	m.logger.Debug("removing stale entity from cache", zap.String("entity", entity), zap.String("id", id.String()))
	if err := m.storage.RemoveStaleID(newSimpleConn(m.db), id, entity); err != nil {
		return errors.Wrapf(err, "failed to RemoveStaleID (%s %s)", entity, id.String())
	}

	return nil
}

func (m *DWH) processBlockBoundary(event *blockchain.Event) {
	if m.lastKnownBlock != event.BlockNumber {
		m.lastKnownBlock = event.BlockNumber
		for {
			if err := m.updateLastKnownBlock(int64(event.BlockNumber)); err != nil {
				m.logger.Warn("failed to updateLastKnownBlock", util.LaconicError(err))
			} else {
				return
			}
		}
	}
}

type eventsDispatcher struct {
	logger                    *zap.Logger
	ValidatorsCreated         []*blockchain.Event
	CertificatesCreated       []*blockchain.Event
	OrdersOpened              []*blockchain.Event
	DealsOpened               []*blockchain.Event
	DealChangeRequestsSent    []*blockchain.Event
	DealChangeRequestsUpdated []*blockchain.Event
	Billed                    []*blockchain.Event
	OrdersClosed              []*blockchain.Event
	DealsClosed               []*blockchain.Event
	ValidatorsDeleted         []*blockchain.Event
	AddedToBlacklist          []*blockchain.Event
	RemovedFromBlacklist      []*blockchain.Event
	WorkersAnnounced          []*blockchain.Event
	WorkersConfirmed          []*blockchain.Event
	WorkersRemoved            []*blockchain.Event
	Other                     []*blockchain.Event
}

func newEventDispatcher(logger *zap.Logger) *eventsDispatcher {
	return &eventsDispatcher{logger: logger}
}

func (m *eventsDispatcher) Add(event *blockchain.Event) {
	switch data := event.Data.(type) {
	case *blockchain.ValidatorCreatedData:
		m.ValidatorsCreated = append(m.ValidatorsCreated, event)
	case *blockchain.ValidatorDeletedData:
		m.ValidatorsDeleted = append(m.ValidatorsDeleted, event)
	case *blockchain.CertificateCreatedData:
		m.CertificatesCreated = append(m.CertificatesCreated, event)
	case *blockchain.DealOpenedData:
		m.DealsOpened = append(m.DealsOpened, event)
	case *blockchain.DealUpdatedData:
		m.DealsClosed = append(m.DealsClosed, event)
	case *blockchain.OrderPlacedData:
		m.OrdersOpened = append(m.OrdersOpened, event)
	case *blockchain.OrderUpdatedData:
		m.OrdersClosed = append(m.OrdersClosed, event)
	case *blockchain.DealChangeRequestSentData:
		m.DealChangeRequestsSent = append(m.DealChangeRequestsSent, event)
	case *blockchain.DealChangeRequestUpdatedData:
		m.DealChangeRequestsUpdated = append(m.DealChangeRequestsUpdated, event)
	case *blockchain.BilledData:
		m.Billed = append(m.Billed, event)
	case *blockchain.AddedToBlacklistData:
		m.AddedToBlacklist = append(m.AddedToBlacklist, event)
	case *blockchain.RemovedFromBlacklistData:
		m.RemovedFromBlacklist = append(m.RemovedFromBlacklist, event)
	case *blockchain.WorkerAnnouncedData:
		m.WorkersAnnounced = append(m.WorkersAnnounced, event)
	case *blockchain.WorkerConfirmedData:
		m.WorkersConfirmed = append(m.WorkersConfirmed, event)
	case *blockchain.WorkerRemovedData:
		m.WorkersRemoved = append(m.WorkersRemoved, event)
	case *blockchain.ErrorData:
		m.logger.Warn("received error from events channel", zap.Error(data.Err), zap.String("topic", data.Topic))
	default:
		m.Other = append(m.Other, event)
	}
}
