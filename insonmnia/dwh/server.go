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

const (
	eventRetryTime = time.Second * 3
)

type DWH struct {
	mu                sync.RWMutex
	ctx               context.Context
	cfg               *Config
	cancel            context.CancelFunc
	grpc              *grpc.Server
	http              *rest.Server
	logger            *zap.Logger
	db                *sql.DB
	creds             credentials.TransportCredentials
	certRotator       util.HitlessCertRotator
	blockchain        blockchain.API
	storage           storage
	numBenchmarks     uint64
	blockEndCallbacks []func() error
	lastKnownBlock    uint64
}

func NewDWH(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (*DWH, error) {
	ctx, cancel := context.WithCancel(ctx)
	w := &DWH{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		logger: log.GetLogger(ctx),
	}

	bch, err := blockchain.NewAPI(blockchain.WithConfig(w.cfg.Blockchain))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to create NewAPI")
	}
	w.blockchain = bch

	numBenchmarks, err := bch.Market().GetNumBenchmarks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to GetNumBenchmarks")
	}

	if numBenchmarks >= NumMaxBenchmarks {
		return nil, errors.New("market number of benchmarks is greater than NumMaxBenchmarks")
	}

	w.numBenchmarks = numBenchmarks

	w.db, err = sql.Open(w.cfg.Storage.Backend, w.cfg.Storage.Endpoint)
	if err != nil {
		return nil, err
	}

	switch w.cfg.Storage.Backend {
	case "sqlite3":
		err = w.setupSQLite(w.db, numBenchmarks)
	case "postgres":
		err = w.setupPostgres(w.db, numBenchmarks)
	default:
		err = errors.Errorf("unsupported backend: %s", cfg.Storage.Backend)
	}
	if err != nil {
		w.db.Close()
		cancel()
		return nil, err
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
	w.logger.Info("starting with backend", zap.String("backend", w.cfg.Storage.Backend),
		zap.String("endpoint", w.cfg.Storage.Endpoint))

	if w.cfg.Blockchain != nil {
		go w.monitorBlockchain()
	} else {
		w.logger.Info("monitoring disabled")
	}

	if w.cfg.ColdStart != nil {
		if err := w.coldStart(); err != nil {
			w.logger.Warn("failed to coldStart", util.LaconicError(err))
			return errors.Wrap(err, "failed to coldStart")
		}
	}

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
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	deals, count, err := w.storage.GetDeals(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetDeals", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDeals")
	}

	return &pb.DWHDealsReply{Deals: deals, Count: count}, nil
}

func (w *DWH) GetDealDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHDeal, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetDealByID(conn, request.Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetDealDetails", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealDetails")
	}

	return out, nil
}

func (w *DWH) GetDealConditions(ctx context.Context, request *pb.DealConditionsRequest) (*pb.DealConditionsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	dealConditions, count, err := w.storage.GetDealConditions(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetDealConditions", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealConditions")
	}

	return &pb.DealConditionsReply{Conditions: dealConditions, Count: count}, nil
}

func (w *DWH) GetOrders(ctx context.Context, request *pb.OrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	orders, count, err := w.storage.GetOrders(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetOrders", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (w *DWH) GetMatchingOrders(ctx context.Context, request *pb.MatchingOrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	orders, count, err := w.storage.GetMatchingOrders(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetMatchingOrders", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetMatchingOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (w *DWH) GetOrderDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHOrder, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetOrderByID(conn, request.Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetOrderDetails", util.LaconicError(err), zap.Any("request", *request))
		return nil, errors.Wrap(err, "failed to GetOrderDetails")
	}

	return out, nil
}

func (w *DWH) GetProfiles(ctx context.Context, request *pb.ProfilesRequest) (*pb.ProfilesReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	profiles, count, err := w.storage.GetProfiles(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetProfiles", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfiles")
	}

	return &pb.ProfilesReply{Profiles: profiles, Count: count}, nil
}

func (w *DWH) GetProfileInfo(ctx context.Context, request *pb.EthID) (*pb.Profile, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetProfileByID(conn, request.GetId().Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetProfileInfo", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfileInfo")
	}

	return out, nil
}

func (w *DWH) GetBlacklist(ctx context.Context, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetBlacklist(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetBlacklist", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (w *DWH) GetValidators(ctx context.Context, request *pb.ValidatorsRequest) (*pb.ValidatorsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	validators, count, err := w.storage.GetValidators(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetValidators", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetValidators")
	}

	return &pb.ValidatorsReply{Validators: validators, Count: count}, nil
}

func (w *DWH) GetDealChangeRequests(ctx context.Context, request *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.getDealChangeRequests(conn, request)
	if err != nil {
		w.logger.Error("failed to GetDealChangeRequests", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealChangeRequests")
	}

	return &pb.DealChangeRequestsReply{Requests: out}, nil
}

func (w *DWH) getDealChangeRequests(conn queryConn, request *pb.BigInt) ([]*pb.DealChangeRequest, error) {
	return w.storage.GetDealChangeRequestsByDealID(conn, request.Unwrap())
}

func (w *DWH) GetWorkers(ctx context.Context, request *pb.WorkersRequest) (*pb.WorkersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	workers, count, err := w.storage.GetWorkers(conn, request)
	if err != nil {
		w.logger.Error("failed to GetWorkers", util.LaconicError(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetWorkers")
	}

	return &pb.WorkersReply{Workers: workers, Count: count}, nil
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
				w.logger.Warn("failed to watch market events, retrying", util.LaconicError(err))
			}
		}
	}
}

func (w *DWH) watchMarketEvents() error {
	var err error
	w.lastKnownBlock, err = w.getLastKnownBlock()
	if err != nil {
		if err := w.insertLastKnownBlock(0); err != nil {
			return err
		}
		w.lastKnownBlock = 0
	}

	w.logger.Info("starting from block", zap.Uint64("block_number", w.lastKnownBlock))
	events, err := w.blockchain.Events().GetEvents(w.ctx, big.NewInt(0).SetUint64(w.lastKnownBlock))
	if err != nil {
		return err
	}

	jobs := make(chan *blockchain.Event)
	for workerID := 0; workerID < w.cfg.NumWorkers; workerID++ {
		go w.runEventWorker(workerID, jobs)
	}

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("context cancelled (watchMarketEvents)")
			return nil
		case event, ok := <-events:
			if !ok {
				close(jobs)
				return errors.New("events channel closed")
			}

			w.processBlockBoundary(event)
			jobs <- event
		}
	}
}

func (w *DWH) runEventWorker(workerID int, events chan *blockchain.Event) {
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("context cancelled (worker)", zap.Int("worker_id", workerID))
			return
		case event, ok := <-events:
			if !ok {
				w.logger.Info("events channel closed", zap.Int("worker_id", workerID))
				return
			}
			if err := w.processEvent(event); err != nil {
				w.logger.Warn("failed to processEvent, retrying", util.LaconicError(err),
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
		w.logger.Warn("received error from events channel", zap.Error(value.Err), zap.String("topic", value.Topic))
	}

	return nil
}

func (w *DWH) retryEvent(event *blockchain.Event) {
	timer := time.NewTimer(eventRetryTime)
	select {
	case <-w.ctx.Done():
		w.logger.Info("context cancelled while retrying event",
			zap.Uint64("block_number", event.BlockNumber),
			zap.String("event_type", reflect.TypeOf(event.Data).String()))
		return
	case <-timer.C:
		if err := w.processEvent(event); err != nil {
			w.logger.Warn("failed to retry processEvent", util.LaconicError(err),
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

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		if err := w.storage.StoreStaleID(newSimpleConn(w.db), dealID, "Deal"); err != nil {
			return errors.Wrap(err, "failed to StoreStaleID")
		}
		w.logger.Debug("skipping inactive deal", zap.String("deal_id", dealID.String()))
		return nil
	}

	if err := w.checkBenchmarks(deal.Benchmarks); err != nil {
		return err
	}

	err = w.storage.InsertDeal(conn, deal)
	if err != nil {
		return errors.Wrapf(err, "failed to insertDeal")
	}

	err = w.storage.InsertDealCondition(conn,
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

func (w *DWH) onDealUpdated(dealID *big.Int) error {
	deal, err := w.blockchain.Market().GetDealInfo(w.ctx, dealID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetDealInfo")
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale:
	if ok, err := w.storage.CheckStaleID(conn, dealID, "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			w.addBlockEndCallback(func() error { return w.removeStaleEntityID(dealID, "Deal") })
			return nil
		}
	}

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		err = w.storage.DeleteDeal(conn, deal.Id.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to delete deal (possibly old log entry)")
		}
		if err := w.storage.DeleteOrder(conn, deal.AskID.Unwrap()); err != nil {
			return errors.Wrap(err, "failed to deleteOrder")
		}
		if err := w.storage.DeleteOrder(conn, deal.BidID.Unwrap()); err != nil {
			return errors.Wrap(err, "failed to deleteOrder")
		}

		return nil
	}

	if err := w.storage.UpdateDeal(conn, deal); err != nil {
		return errors.Wrapf(err, "failed to UpdateDeal")
	}

	return nil
}

func (w *DWH) onDealChangeRequestSent(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := w.blockchain.Market().GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := w.storage.CheckStaleID(conn, changeRequest.DealID.Unwrap(), "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			w.logger.Debug("skipping DealChangeRequestSent event for inactive deal")
			return nil
		}
	}

	if changeRequest.Status != pb.ChangeRequestStatus_REQUEST_CREATED {
		w.logger.Info("onDealChangeRequest event points to DealChangeRequest with .Status != Created",
			zap.String("actual_status", pb.ChangeRequestStatus_name[int32(changeRequest.Status)]))
		return nil
	}

	// Sanity check: if more than 1 CR of one type is created for a Deal, we delete old CRs.
	expiredChangeRequests, err := w.storage.GetDealChangeRequests(conn, changeRequest)
	if err != nil {
		return errors.New("failed to get (possibly) expired DealChangeRequests")
	}

	for _, expiredChangeRequest := range expiredChangeRequests {
		err := w.storage.DeleteDealChangeRequest(conn, expiredChangeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to deleteDealChangeRequest")
		} else {
			w.logger.Warn("deleted expired DealChangeRequest",
				zap.String("id", expiredChangeRequest.Id.Unwrap().String()))
		}
	}

	changeRequest.CreatedTS = &pb.Timestamp{Seconds: int64(eventTS)}
	if err := w.storage.InsertDealChangeRequest(conn, changeRequest); err != nil {
		return errors.Wrap(err, "failed to insertDealChangeRequest")
	}

	return err
}

func (w *DWH) onDealChangeRequestUpdated(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := w.blockchain.Market().GetDealChangeRequestInfo(w.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := w.storage.CheckStaleID(conn, changeRequest.DealID.Unwrap(), "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			w.logger.Debug("skipping DealChangeRequestUpdated event for inactive deal")
			return nil
		}
	}

	switch changeRequest.Status {
	case pb.ChangeRequestStatus_REQUEST_REJECTED:
		err := w.storage.UpdateDealChangeRequest(conn, changeRequest)
		if err != nil {
			return errors.Wrapf(err, "failed to update DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	case pb.ChangeRequestStatus_REQUEST_ACCEPTED:
		deal, err := w.storage.GetDealByID(conn, changeRequest.DealID.Unwrap())
		if err != nil {
			return errors.Wrap(err, "failed to storage.GetDealByID")
		}

		if err := w.updateDealConditionEndTime(conn, deal.GetDeal().Id, eventTS); err != nil {
			return errors.Wrap(err, "failed to updateDealConditionEndTime")
		}
		err = w.storage.InsertDealCondition(conn,
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

		err = w.storage.DeleteDealChangeRequest(conn, changeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	default:
		err := w.storage.DeleteDealChangeRequest(conn, changeRequest.Id.Unwrap())
		if err != nil {
			return errors.Wrapf(err, "failed to delete DealChangeRequest %s", changeRequest.Id.Unwrap().String())
		}
	}

	return nil
}

func (w *DWH) onBilled(eventTS uint64, dealID, payedAmount *big.Int) error {
	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If deal is known to be stale, skip.
	if ok, err := w.storage.CheckStaleID(conn, dealID, "Deal"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			w.logger.Debug("skipping Billed event for inactive deal")
			return nil
		}
	}

	if err := w.updateDealPayout(conn, dealID, payedAmount, eventTS); err != nil {
		return errors.Wrap(err, "failed to updateDealPayout")
	}

	dealConditions, _, err := w.storage.GetDealConditions(conn, &pb.DealConditionsRequest{DealID: pb.NewBigInt(dealID)})
	if err != nil {
		return errors.Wrap(err, "failed to GetDealConditions (last)")
	}

	if len(dealConditions) < 1 {
		return errors.Errorf("no deal conditions found for deal `%s`", dealID.String())
	}

	err = w.storage.UpdateDealConditionPayout(conn, dealConditions[0].Id,
		big.NewInt(0).Add(dealConditions[0].TotalPayout.Unwrap(), payedAmount))
	if err != nil {
		return errors.Wrap(err, "failed to UpdateDealConditionPayout")
	}

	if err != nil {
		return errors.Wrap(err, "insertDealPayment failed")
	}

	return nil
}

func (w *DWH) updateDealPayout(conn queryConn, dealID, payedAmount *big.Int, billTS uint64) error {
	deal, err := w.storage.GetDealByID(conn, dealID)
	if err != nil {
		return errors.Wrap(err, "failed to GetDealByID")
	}

	newDealTotalPayout := big.NewInt(0).Add(deal.Deal.TotalPayout.Unwrap(), payedAmount)
	err = w.storage.UpdateDealPayout(conn, dealID, newDealTotalPayout, billTS)
	if err != nil {
		return errors.Wrap(err, "failed to updateDealPayout")
	}

	return nil
}

func (w *DWH) onOrderPlaced(eventTS uint64, orderID *big.Int) error {
	order, err := w.blockchain.Market().GetOrderInfo(w.ctx, orderID)
	if err != nil {
		return errors.Wrapf(err, "failed to GetOrderInfo")
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if order.OrderStatus == pb.OrderStatus_ORDER_INACTIVE && order.DealID.IsZero() {
		if err := w.storage.StoreStaleID(conn, orderID, "Order"); err != nil {
			return errors.Wrap(err, "failed to StoreStaleID")
		}
		w.logger.Debug("skipping inactive order", zap.String("order_id", orderID.String()))
		return nil
	}

	profile, err := w.storage.GetProfileByID(conn, order.AuthorID.Unwrap())
	if err != nil {
		certificates, _ := json.Marshal([]*pb.Certificate{})
		profile = &pb.Profile{UserID: order.AuthorID, Certificates: string(certificates)}
	} else {
		if err := w.updateProfileStats(conn, order, 1); err != nil {
			return errors.Wrap(err, "failed to updateProfileStats")
		}
	}

	if order.DealID == nil {
		order.DealID = pb.NewBigIntFromInt(0)
	}

	if err := w.checkBenchmarks(order.Benchmarks); err != nil {
		return err
	}

	err = w.storage.InsertOrder(conn, &pb.DWHOrder{
		CreatedTS:            &pb.Timestamp{Seconds: int64(eventTS)},
		CreatorIdentityLevel: profile.IdentityLevel,
		CreatorName:          profile.Name,
		CreatorCountry:       profile.Country,
		CreatorCertificates:  []byte(profile.Certificates),
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

func (w *DWH) onOrderUpdated(orderID *big.Int) error {
	order, err := w.blockchain.Market().GetOrderInfo(w.ctx, orderID)
	if err != nil {
		return errors.Wrap(err, "failed to GetOrderInfo")
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	// If the order was known to be inactive, delete it from the list of inactive entities
	// and skip.
	if ok, err := w.storage.CheckStaleID(conn, orderID, "Order"); err != nil {
		return errors.Wrap(err, "failed to CheckStaleID")
	} else {
		if ok {
			w.removeStaleEntityID(orderID, "Order")
			return nil
		}
	}

	// If order was updated, but no deal is associated with it, delete the order.
	if order.DealID.IsZero() {
		if err := w.storage.DeleteOrder(conn, orderID); err != nil {
			w.logger.Info("failed to delete Order (possibly old log entry)", util.LaconicError(err),
				zap.String("order_id", orderID.String()))
		}
	} else {
		// Otherwise update order status.
		err := w.storage.UpdateOrderStatus(conn, order.Id.Unwrap(), order.OrderStatus)
		if err != nil {
			return errors.Wrap(err, "failed to updateOrderStatus (possibly old log entry)")
		}
	}

	if err := w.updateProfileStats(conn, order, -1); err != nil {
		return errors.Wrapf(err, "failed to updateProfileStats (AuthorID: `%s`)", order.AuthorID.Unwrap().String())
	}

	return nil
}

func (w *DWH) onWorkerAnnounced(masterID, slaveID string) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.InsertWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerAnnounced failed")
	}

	return nil
}

func (w *DWH) onWorkerConfirmed(masterID, slaveID string) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.UpdateWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerConfirmed failed")
	}

	return nil
}

func (w *DWH) onWorkerRemoved(masterID, slaveID string) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.DeleteWorker(conn, masterID, slaveID); err != nil {
		return errors.Wrap(err, "onWorkerRemoved failed")
	}

	return nil
}

func (w *DWH) onAddedToBlacklist(adderID, addeeID string) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.InsertBlacklistEntry(conn, adderID, addeeID); err != nil {
		return errors.Wrap(err, "onAddedToBlacklist failed")
	}

	return nil
}

func (w *DWH) onRemovedFromBlacklist(removerID, removeeID string) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.DeleteBlacklistEntry(conn, removerID, removeeID); err != nil {
		return errors.Wrap(err, "onRemovedFromBlacklist failed")
	}

	return nil
}

func (w *DWH) onValidatorCreated(validatorID common.Address) error {
	validator, err := w.blockchain.ProfileRegistry().GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.InsertValidator(conn, validator); err != nil {
		return errors.Wrap(err, "failed to insertValidator")
	}

	return nil
}

func (w *DWH) onValidatorDeleted(validatorID common.Address) error {
	validator, err := w.blockchain.ProfileRegistry().GetValidator(w.ctx, validatorID)
	if err != nil {
		return errors.Wrapf(err, "failed to get validator `%s`", validatorID.String())
	}

	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.UpdateValidator(conn, validator); err != nil {
		return errors.Wrap(err, "failed to updateValidator")
	}

	return nil
}

func (w *DWH) onCertificateCreated(certificateID *big.Int) error {
	certificate, err := w.blockchain.ProfileRegistry().GetCertificate(w.ctx, certificateID)
	if err != nil {
		return errors.Wrap(err, "failed to GetCertificate")
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

	if err = w.storage.InsertCertificate(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to insertCertificate")
	}

	if err := w.updateProfile(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to updateProfile")
	}

	if err := w.updateEntitiesByProfile(conn, certificate); err != nil {
		return errors.Wrap(err, "failed to updateEntitiesByProfile")
	}

	return nil
}

func (w *DWH) updateProfile(conn queryConn, certificate *pb.Certificate) error {
	certBytes, _ := json.Marshal([]*pb.Certificate{})
	err := w.storage.InsertProfileUserID(conn, &pb.Profile{
		UserID:       certificate.OwnerID,
		Certificates: string(certBytes),
		ActiveAsks:   0,
		ActiveBids:   0,
	})
	if err != nil {
		return errors.Wrap(err, "failed to insertProfileUserID")
	}

	// Update distinct Profile columns.
	switch certificate.Attribute {
	case CertificateName, CertificateCountry:
		err := w.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), attributeToString[certificate.Attribute],
			string(certificate.Value))
		if err != nil {
			return errors.Wrapf(err, "failed to UpdateProfile (%s)", attributeToString[certificate.Attribute])
		}
	}

	// Update certificates blob.
	certificates, err := w.storage.GetCertificates(conn, certificate.OwnerID.Unwrap())
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

	err = w.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), "Certificates", certificateAttrsBytes)
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Certificates)")
	}

	err = w.storage.UpdateProfile(conn, certificate.OwnerID.Unwrap(), "IdentityLevel", maxIdentityLevel)
	if err != nil {
		return errors.Wrap(err, "failed to updateProfileCertificates (Level)")
	}

	return nil
}

func (w *DWH) updateEntitiesByProfile(conn queryConn, certificate *pb.Certificate) error {
	profile, err := w.storage.GetProfileByID(conn, certificate.OwnerID.Unwrap())
	if err != nil {
		return errors.Wrap(err, "failed to getProfileInfo")
	}

	if err := w.storage.UpdateOrders(conn, profile); err != nil {
		return errors.Wrap(err, "failed to updateOrders")
	}

	if err = w.storage.UpdateDealsSupplier(conn, profile); err != nil {
		return errors.Wrap(err, "failed to updateDealsSupplier")
	}

	err = w.storage.UpdateDealsConsumer(conn, profile)
	if err != nil {
		return errors.Wrap(err, "failed to updateDealsConsumer")
	}

	return nil
}

func (w *DWH) updateProfileStats(conn queryConn, order *pb.Order, update int) error {
	var field string
	if order.OrderType == pb.OrderType_ASK {
		field = "ActiveAsks"
	} else {
		field = "ActiveBids"
	}

	if err := w.storage.UpdateProfileStats(conn, order.AuthorID.Unwrap(), field, update); err != nil {
		return errors.Wrap(err, "failed to UpdateProfileStats")
	}

	return nil
}

// coldStart waits till last seen block number gets to `w.cfg.ColdStart.UpToBlock` and then tries to create indices.
func (w *DWH) coldStart() error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	var retries = 5
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Info("stopped coldStart routine")
			return nil
		case <-ticker.C:
			targetBlockReached, err := w.maybeCreateIndices()
			if err != nil {
				if retries == 0 {
					w.logger.Warn("failed to CreateIndices, exiting")
					return err
				}
				retries--
				w.logger.Warn("failed to CreateIndices, retrying", zap.Int("retries_left", retries))
			}
			if targetBlockReached {
				w.logger.Info("CreateIndices success")
				return nil
			}
		}
	}
}

func (w *DWH) maybeCreateIndices() (targetBlockReached bool, err error) {
	lastBlock, err := w.getLastKnownBlock()
	if err != nil {
		return false, err
	}

	w.logger.Info("current block (waiting to CreateIndices)", zap.Uint64("block_number", lastBlock))
	if lastBlock >= w.cfg.ColdStart.UpToBlock {
		if err := w.storage.CreateIndices(w.db); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (w *DWH) getLastKnownBlock() (uint64, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	return w.storage.GetLastKnownBlock(conn)
}

func (w *DWH) updateLastKnownBlock(blockNumber int64) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.UpdateLastKnownBlock(conn, blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (w *DWH) insertLastKnownBlock(blockNumber int64) error {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	if err := w.storage.InsertLastKnownBlock(conn, blockNumber); err != nil {
		return errors.Wrap(err, "failed to updateLastKnownBlock")
	}

	return nil
}

func (w *DWH) updateDealConditionEndTime(conn queryConn, dealID *pb.BigInt, eventTS uint64) error {
	dealConditions, _, err := w.storage.GetDealConditions(conn, &pb.DealConditionsRequest{DealID: dealID})
	if err != nil {
		return errors.Wrap(err, "failed to getDealConditions")
	}

	dealCondition := dealConditions[0]
	if err := w.storage.UpdateDealConditionEndTime(conn, dealCondition.Id, eventTS); err != nil {
		return errors.Wrap(err, "failed to update DealCondition")
	}

	return nil
}

func (w *DWH) checkBenchmarks(benches *pb.Benchmarks) error {
	if uint64(len(benches.Values)) != w.numBenchmarks {
		return errors.Errorf("expected %d benchmarks, got %d", w.numBenchmarks, len(benches.Values))
	}

	for idx, bench := range benches.Values {
		if bench >= MaxBenchmark {
			return errors.Errorf("benchmark %d is greater that %d", idx, MaxBenchmark)
		}
	}

	return nil
}

func (w *DWH) removeStaleEntityID(id *big.Int, entity string) error {
	w.logger.Debug("removing stale entity from cache", zap.String("entity", entity), zap.String("id", id.String()))
	if err := w.storage.RemoveStaleID(newSimpleConn(w.db), id, entity); err != nil {
		return errors.Wrapf(err, "failed to RemoveStaleID (%s %s)", entity, id.String())
	}

	return nil
}

func (w *DWH) addBlockEndCallback(cb func() error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.blockEndCallbacks = append(w.blockEndCallbacks, cb)
}

func (w *DWH) processBlockBoundary(event *blockchain.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.lastKnownBlock != event.BlockNumber {
		go func(callbacks []func() error) {
			for _, cb := range callbacks {
				if err := cb(); err != nil {
					w.logger.Warn("failed to execute cb after block end", util.LaconicError(err))
				}
			}
		}(w.blockEndCallbacks[:])
		w.blockEndCallbacks = w.blockEndCallbacks[:0]
		w.lastKnownBlock = event.BlockNumber
		if err := w.updateLastKnownBlock(int64(event.BlockNumber)); err != nil {
			w.logger.Warn("failed to updateLastKnownBlock", util.LaconicError(err),
				zap.Uint64("block_number", event.BlockNumber))
		}
	}
}
