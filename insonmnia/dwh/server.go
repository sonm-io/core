package dwh

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
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
)

type DWH struct {
	mu            sync.RWMutex
	ctx           context.Context
	cfg           *Config
	cancel        context.CancelFunc
	grpc          *grpc.Server
	http          *rest.Server
	logger        *zap.Logger
	db            *sql.DB
	creds         credentials.TransportCredentials
	certRotator   util.HitlessCertRotator
	blockchain    blockchain.API
	storage       storage
	numBenchmarks uint64
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
			w.logger.Warn("failed to coldStart", zap.Error(err))
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
		w.logger.Warn("failed to GetDeals", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDeals")
	}

	return &pb.DWHDealsReply{Deals: deals, Count: count}, nil
}

func (w *DWH) GetDealDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHDeal, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetDealByID(conn, request.Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetDealDetails", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealDetails")
	}

	return out, nil
}

func (w *DWH) GetDealConditions(ctx context.Context, request *pb.DealConditionsRequest) (*pb.DealConditionsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	dealConditions, count, err := w.storage.GetDealConditions(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetDealConditions", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealConditions")
	}

	return &pb.DealConditionsReply{Conditions: dealConditions, Count: count}, nil
}

func (w *DWH) GetOrders(ctx context.Context, request *pb.OrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	orders, count, err := w.storage.GetOrders(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetOrders", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (w *DWH) GetMatchingOrders(ctx context.Context, request *pb.MatchingOrdersRequest) (*pb.DWHOrdersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	orders, count, err := w.storage.GetMatchingOrders(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetMatchingOrders", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetMatchingOrders")
	}

	return &pb.DWHOrdersReply{Orders: orders, Count: count}, nil
}

func (w *DWH) GetOrderDetails(ctx context.Context, request *pb.BigInt) (*pb.DWHOrder, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetOrderByID(conn, request.Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetOrderDetails", zap.Error(err), zap.Any("request", *request))
		return nil, errors.Wrap(err, "failed to GetOrderDetails")
	}

	return out, nil
}

func (w *DWH) GetProfiles(ctx context.Context, request *pb.ProfilesRequest) (*pb.ProfilesReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	profiles, count, err := w.storage.GetProfiles(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetProfiles", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfiles")
	}

	return &pb.ProfilesReply{Profiles: profiles, Count: count}, nil
}

func (w *DWH) GetProfileInfo(ctx context.Context, request *pb.EthID) (*pb.Profile, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetProfileByID(conn, request.GetId().Unwrap())
	if err != nil {
		w.logger.Warn("failed to GetProfileInfo", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetProfileInfo")
	}

	return out, nil
}

func (w *DWH) GetBlacklist(ctx context.Context, request *pb.BlacklistRequest) (*pb.BlacklistReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.storage.GetBlacklist(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetBlacklist", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetBlacklist")
	}

	return out, nil
}

func (w *DWH) GetValidators(ctx context.Context, request *pb.ValidatorsRequest) (*pb.ValidatorsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	validators, count, err := w.storage.GetValidators(conn, request)
	if err != nil {
		w.logger.Warn("failed to GetValidators", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetValidators")
	}

	return &pb.ValidatorsReply{Validators: validators, Count: count}, nil
}

func (w *DWH) GetDealChangeRequests(ctx context.Context, request *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	out, err := w.getDealChangeRequests(conn, request)
	if err != nil {
		w.logger.Error("failed to GetDealChangeRequests", zap.Error(err), zap.Any("request", *request))
		return nil, status.Error(codes.NotFound, "failed to GetDealChangeRequests")
	}

	return &pb.DealChangeRequestsReply{Requests: out}, nil
}

func (w *DWH) getDealChangeRequests(conn queryConn, request *pb.BigInt) ([]*pb.DealChangeRequest, error) {
	return w.storage.GetDealChangeRequestsByID(conn, request.Unwrap())
}

func (w *DWH) GetWorkers(ctx context.Context, request *pb.WorkersRequest) (*pb.WorkersReply, error) {
	conn := newSimpleConn(w.db)
	defer conn.Finish()

	workers, count, err := w.storage.GetWorkers(conn, request)
	if err != nil {
		w.logger.Error("failed to GetWorkers", zap.Error(err), zap.Any("request", *request))
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
				w.logger.Warn("failed to watch market events, retrying", zap.Error(err))
			}
		}
	}
}

func (w *DWH) watchMarketEvents() error {
	lastKnownBlock, err := w.getLastKnownBlock()
	if err != nil {
		if err := w.insertLastKnownBlock(0); err != nil {
			return err
		}
		lastKnownBlock = 0
	}

	w.logger.Info("starting from block", zap.Uint64("block_number", lastKnownBlock))
	events, err := w.blockchain.Events().GetEvents(w.ctx, big.NewInt(0).SetUint64(lastKnownBlock))
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	for workerID := 0; workerID < w.cfg.NumWorkers; workerID++ {
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
			if err := w.updateLastKnownBlock(int64(event.BlockNumber)); err != nil {
				w.logger.Warn("failed to updateLastKnownBlock", zap.Error(err),
					zap.Uint64("block_number", event.BlockNumber), zap.Int("worker_id", workerID))
			}
			// Events in the same block can come in arbitrary order. If two events have to be processed
			// in a specific order (e.g., OrderPlaced > DealOpened), we need to retry if the order is
			// messed up.
			if err := w.processEvent(event); err != nil {
				if strings.Contains(err.Error(), "constraint") {
					continue
				}
				w.logger.Warn("failed to processEvent, retrying", zap.Error(err),
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
			w.logger.Warn("failed to retry processEvent", zap.Error(err),
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

	if deal.Status == pb.DealStatus_DEAL_CLOSED {
		w.logger.Info("skipping inactive deal", zap.String("deal_id", dealID.String()))
		return nil
	}

	if err := w.checkBenchmarks(deal.Benchmarks); err != nil {
		return err
	}

	conn, err := newTxConn(w.db, w.logger)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer conn.Finish()

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

	err = w.storage.InsertDealPayment(conn, &pb.DealPayment{
		DealID:      pb.NewBigInt(dealID),
		PayedAmount: pb.NewBigInt(payedAmount),
		PaymentTS:   &pb.Timestamp{Seconds: int64(eventTS)},
	})
	if err != nil {
		return errors.Wrap(err, "insertDealPayment failed")
	}

	return nil
}

func (w *DWH) updateDealPayout(conn queryConn, dealID, payedAmount *big.Int, billTS uint64) error {
	deal, err := w.storage.GetDealByID(conn, dealID)
	if err != nil {
		return errors.Wrap(err, "failed to storage.GetDealByID")
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

	profile, err := w.storage.GetProfileByID(conn, order.AuthorID.Unwrap())
	if err != nil {
		var askOrders, bidOrders uint64 = 0, 0
		if order.OrderType == pb.OrderType_ASK {
			askOrders = 1
		} else {
			bidOrders = 1
		}
		certificates, _ := json.Marshal([]*pb.Certificate{})
		profile = &pb.Profile{
			UserID:       order.AuthorID,
			Certificates: string(certificates),
			ActiveAsks:   askOrders,
			ActiveBids:   bidOrders,
		}
		err = w.storage.InsertProfileUserID(conn, profile)
		if err != nil {
			return errors.Wrap(err, "failed to insertProfileUserID")
		}
	} else {
		if err := w.updateProfileStats(conn, order, profile, 1); err != nil {
			return errors.Wrap(err, "failed to updateProfileStats")
		}
	}

	if order.OrderStatus == pb.OrderStatus_ORDER_INACTIVE && order.DealID.IsZero() {
		w.logger.Info("skipping inactive order", zap.String("order_id", order.Id.Unwrap().String()))
		return nil
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

	// If order was updated, but no deal is associated with it, delete the order.
	if order.DealID.IsZero() {
		if err := w.storage.DeleteOrder(conn, orderID); err != nil {
			w.logger.Info("failed to delete Order (possibly old log entry)", zap.Error(err),
				zap.String("order_id", orderID.String()))
		}
	} else {
		// Otherwise update order status.
		err := w.storage.UpdateOrderStatus(conn, order.Id.Unwrap(), order.OrderStatus)
		if err != nil {
			return errors.Wrap(err, "failed to updateOrderStatus (possibly old log entry)")
		}
	}

	profile, err := w.storage.GetProfileByID(conn, order.AuthorID.Unwrap())
	if err != nil {
		return errors.Wrapf(err, "failed to getProfileInfo (AuthorID: `%s`)", order.AuthorID)
	}

	if err := w.updateProfileStats(conn, order, profile, -1); err != nil {
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
	// Create a Profile entry if it doesn't exist yet.
	certBytes, _ := json.Marshal([]*pb.Certificate{})
	if _, err := w.storage.GetProfileByID(conn, certificate.OwnerID.Unwrap()); err != nil {
		err = w.storage.InsertProfileUserID(conn, &pb.Profile{
			UserID:       certificate.OwnerID,
			Certificates: string(certBytes),
			ActiveAsks:   0,
			ActiveBids:   0,
		})
		if err != nil {
			return errors.Wrap(err, "failed to insertProfileUserID")
		}
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

func (w *DWH) updateProfileStats(conn queryConn, order *pb.Order, profile *pb.Profile, update int) error {
	var (
		field string
		value int
	)
	if order.OrderType == pb.OrderType_ASK {
		updateResult := int(profile.ActiveAsks) + update
		if updateResult < 0 {
			return errors.Errorf("updateProfileStats resulted in a negative Asks value (UserID: `%s`)",
				order.AuthorID.Unwrap().Hex())
		}
		field, value = "ActiveAsks", updateResult
	} else {
		updateResult := int(profile.ActiveBids) + update
		if updateResult < 0 {
			return errors.Errorf("updateProfileStats resulted in a negative Bids value (UserID: `%s`)",
				order.AuthorID.Unwrap().Hex())
		}
		field, value = "ActiveBids", updateResult
	}

	if err := w.storage.UpdateProfile(conn, order.AuthorID.Unwrap(), field, value); err != nil {
		return errors.Wrap(err, "failed to updateProfile")
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

func (w *DWH) addBenchmarksConditions(benches map[uint64]*pb.MaxMinUint64, filters *[]*filter) {
	for benchID, condition := range benches {
		if condition.Max > 0 {
			*filters = append(*filters, newFilter(getBenchmarkColumn(benchID), lte, condition.Max, "AND"))
		}
		if condition.Min > 0 {
			*filters = append(*filters, newFilter(getBenchmarkColumn(benchID), gte, condition.Max, "AND"))
		}
	}
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
