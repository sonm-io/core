package dwh

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type L1Processor struct {
	cfg        *L1ProcessorConfig
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.Logger
	db         *sql.DB
	blockchain blockchain.API
	storage    *sqlStorage
	lastEvent  *blockchain.Event
}

func NewL1Processor(ctx context.Context, cfg *L1ProcessorConfig) (*L1Processor, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &L1Processor{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		logger: log.GetLogger(ctx),
	}, nil
}

func (m *L1Processor) Start() error {
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

	if m.storage, err = setupDB(m.ctx, m.db, m.blockchain); err != nil {
		m.Stop()
		return fmt.Errorf("failed to setupDB: %v", err)
	}

	go func() {
		if m.cfg.ColdStart {
			if err := m.coldStart(); err != nil {
				m.logger.Warn("failed to coldStart", zap.Error(err))
				m.Stop()
			}
		} else {
			if err := m.storage.CreateIndices(m.db); err != nil {
				m.logger.Warn("failed to CreateIndices (Serve)", zap.Error(err))
				m.Stop()
			}
		}
	}()

	return m.monitorBlockchain()
}

func (m *L1Processor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stop()
}

func (m *L1Processor) stop() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.db != nil {
		m.db.Close()
	}
}

func (m *L1Processor) monitorBlockchain() error {
	m.logger.Info("starting monitoring")
	for {
		select {
		case <-m.ctx.Done():
			m.logger.Info("context cancelled (monitorBlockchain)")
			return nil
		default:
			if err := m.watchMarketEvents(); err != nil {
				m.logger.Warn("failed to watch market events, retrying", zap.Error(err))
			}
		}
	}
}

func (m *L1Processor) watchMarketEvents() error {
	var err error
	m.lastEvent, err = m.getLastEvent()
	if err != nil {
		m.lastEvent = &blockchain.Event{}
		if err := m.insertLastEvent(m.lastEvent); err != nil {
			return err
		}
	}

	m.logger.Info("starting from block", zap.Uint64("block_number", m.lastEvent.BlockNumber))
	filter := m.blockchain.Events().GetMarketFilter(big.NewInt(0).SetUint64(m.lastEvent.BlockNumber))
	events, err := m.blockchain.Events().GetEvents(m.ctx, filter)
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
		err := func() error {
			m.mu.Lock()
			defer m.mu.Unlock()

			select {
			case <-m.ctx.Done():
				return errors.New("watchMarketEvents: context cancelled)")
			case <-tk.C:
				m.processEvents(dispatcher)
				eventsCount, dispatcher = 0, newEventDispatcher(m.logger)
			case event, ok := <-events:
				if !ok {
					return errors.New("events channel closed")
				}

				if event.PrecedesOrEquals(m.lastEvent) {
					return nil
				}

				dispatcher.Add(event)
				m.lastEvent, eventsCount = event, eventsCount+1
				if eventsCount >= m.cfg.NumWorkers {
					m.processEvents(dispatcher)
					eventsCount, dispatcher = 0, newEventDispatcher(m.logger)
				}
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}
}

func (m *L1Processor) processEvents(dispatcher *eventsDispatcher) {
	m.processEventsSynchronous(dispatcher.NumBenchmarksUpdated)
	m.processEventsSynchronous(dispatcher.WorkersAnnounced)
	m.processEventsSynchronous(dispatcher.WorkersConfirmed)
	m.processEventsSynchronous(dispatcher.ValidatorsCreated)
	m.processEventsSynchronous(dispatcher.CertificatesCreated)
	m.processEventsAsync(dispatcher.OrdersOpened)
	m.processEventsAsync(dispatcher.DealsOpened)
	m.processEventsSynchronous(dispatcher.DealChangeRequestsSent)
	m.processEventsAsync(dispatcher.Billed)
	m.processEventsAsync(dispatcher.DealChangeRequestsUpdated)
	m.processEventsAsync(dispatcher.OrdersClosed)
	m.processEventsAsync(dispatcher.DealsClosed)
	m.processEventsAsync(dispatcher.ValidatorsDeleted)
	m.processEventsSynchronous(dispatcher.CertificatesUpdated)
	m.processEventsAsync(dispatcher.AddedToBlacklist)
	m.processEventsAsync(dispatcher.RemovedFromBlacklist)
	m.processEventsAsync(dispatcher.WorkersRemoved)
	m.processEventsAsync(dispatcher.Other)

	m.saveLastEvent()
}

func (m *L1Processor) processEventsSynchronous(events []*blockchain.Event) {
	for _, event := range events {
		m.processEventWithRetries(event)
	}
}

func (m *L1Processor) processEventsAsync(events []*blockchain.Event) {
	wg := &sync.WaitGroup{}
	for _, event := range events {
		wg.Add(1)
		go func(wg *sync.WaitGroup, event *blockchain.Event) {
			defer wg.Done()
			if err := m.processEventWithRetries(event); err != nil {
				m.logger.Warn("failed to processEvent, STATE IS INCONSISTENT", zap.Error(err),
					zap.Uint64("block_number", event.BlockNumber),
					zap.String("event_type", reflect.TypeOf(event.Data).String()),
					zap.Any("event_data", event.Data))
			}
		}(wg, event)
	}
	wg.Wait()
}

func (m *L1Processor) processEventWithRetries(event *blockchain.Event) error {
	var (
		err        error
		numRetries = 60
	)
	for numRetries > 0 {
		if err = m.processEvent(event); err != nil {
			m.logger.Warn("failed to processEvent, retrying", zap.Error(err),
				zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).String()),
				zap.Any("event_data", event.Data))
		} else {
			m.logger.Debug("processed event", zap.Uint64("block_number", event.BlockNumber),
				zap.String("event_type", reflect.TypeOf(event.Data).String()),
				zap.Any("event_data", event.Data))
			return nil
		}
		numRetries--
		time.Sleep(time.Second)
	}
	return err
}

func (m *L1Processor) processEvent(event *blockchain.Event) error {
	switch value := event.Data.(type) {
	case *blockchain.NumBenchmarksUpdatedData:
		return m.onNumBenchmarksUpdated(value.NumBenchmarks)
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
	case *blockchain.CertificateUpdatedData:
		return m.onCertificateUpdated(value.ID)
	}

	return nil
}

func (m *L1Processor) onNumBenchmarksUpdated(newNumBenchmarks uint64) error {
	var err error
	if m.storage, err = setupDB(m.ctx, m.db, m.blockchain); err != nil {
		return fmt.Errorf("failed to setupDB after NumBenchmarksUpdated event: %v", err)
	}

	if err := m.storage.CreateIndices(m.db); err != nil {
		return fmt.Errorf("failed to CreateIndices (onNumBenchmarksUpdated): %v", err)
	}

	return nil
}

func (m *L1Processor) onDealOpened(dealID *big.Int) error {
	deal, err := m.blockchain.Market().GetDealInfo(m.ctx, dealID)
	if err != nil {
		return fmt.Errorf("failed to GetDealInfo: %v", err)
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	err = m.storage.InsertDeal(conn, deal)
	if err != nil {
		return fmt.Errorf("failed to insertDeal: %v", err)
	}

	err = m.storage.InsertDealCondition(conn,
		&sonm.DealCondition{
			SupplierID:  deal.SupplierID,
			ConsumerID:  deal.ConsumerID,
			MasterID:    deal.MasterID,
			Duration:    deal.Duration,
			Price:       deal.Price,
			StartTime:   deal.StartTime,
			EndTime:     &sonm.Timestamp{},
			TotalPayout: deal.TotalPayout,
			DealID:      deal.Id,
		},
	)
	if err != nil {
		return fmt.Errorf("onDealOpened: failed to insertDealCondition: %v", err)
	}

	return nil
}

func (m *L1Processor) onDealUpdated(dealID *big.Int) error {
	deal, err := m.blockchain.Market().GetDealInfo(m.ctx, dealID)
	if err != nil {
		return fmt.Errorf("failed to GetDealInfo: %v", err)
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	if err := m.storage.UpdateDeal(conn, deal); err != nil {
		return fmt.Errorf("failed to UpdateDeal: %v", err)
	}

	return nil
}

func (m *L1Processor) onDealChangeRequestSent(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := m.blockchain.Market().GetDealChangeRequestInfo(m.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	changeRequest.CreatedTS = &sonm.Timestamp{Seconds: int64(eventTS)}
	if err := m.storage.InsertDealChangeRequest(conn, changeRequest); err != nil {
		return fmt.Errorf("failed to InsertDealChangeRequest (%s): %v", changeRequest.Id.Unwrap().String(), err)
	}

	return err
}

func (m *L1Processor) onDealChangeRequestUpdated(eventTS uint64, changeRequestID *big.Int) error {
	changeRequest, err := m.blockchain.Market().GetDealChangeRequestInfo(m.ctx, changeRequestID)
	if err != nil {
		return err
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	if changeRequest.Status == sonm.ChangeRequestStatus_REQUEST_ACCEPTED {
		deal, err := m.storage.GetDealByID(conn, changeRequest.DealID.Unwrap())
		if err != nil {
			return fmt.Errorf("failed to storage.GetDealByID: %v", err)
		}

		deal.Deal.Duration = changeRequest.Duration
		deal.Deal.Price = changeRequest.Price
		if err := m.storage.UpdateDeal(conn, deal.Deal); err != nil {
			return fmt.Errorf("failed to UpdateDeal: %v", err)
		}

		if err := m.updateDealConditionEndTime(conn, deal.GetDeal().Id, eventTS); err != nil {
			return fmt.Errorf("failed to updateDealConditionEndTime: %v", err)
		}

		err = m.storage.InsertDealCondition(conn,
			&sonm.DealCondition{
				SupplierID:  deal.GetDeal().SupplierID,
				ConsumerID:  deal.GetDeal().ConsumerID,
				MasterID:    deal.GetDeal().MasterID,
				Duration:    changeRequest.Duration,
				Price:       changeRequest.Price,
				StartTime:   &sonm.Timestamp{Seconds: int64(eventTS)},
				EndTime:     &sonm.Timestamp{},
				TotalPayout: sonm.NewBigIntFromInt(0),
				DealID:      deal.GetDeal().Id,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to insertDealCondition: %v", err)
		}
	}

	if err := m.storage.UpdateDealChangeRequest(conn, changeRequest); err != nil {
		return fmt.Errorf("failed to update DealChangeRequest %s: %v", changeRequest.Id.Unwrap().String(), err)
	}

	return nil
}

func (m *L1Processor) onBilled(eventTS uint64, dealID, payedAmount *big.Int) error {
	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	if err := m.updateDealPayout(conn, dealID, payedAmount, eventTS); err != nil {
		return fmt.Errorf("failed to updateDealPayout: %v", err)
	}

	dealConditions, _, err := m.storage.GetDealConditions(conn, &sonm.DealConditionsRequest{DealID: sonm.NewBigInt(dealID)})
	if err != nil {
		return fmt.Errorf("failed to GetDealConditions (last): %v", err)
	}

	if len(dealConditions) < 1 {
		return fmt.Errorf("no deal conditions found for deal `%s`: %v", dealID.String(), err)
	}

	err = m.storage.UpdateDealConditionPayout(conn, dealConditions[0].Id,
		big.NewInt(0).Add(dealConditions[0].TotalPayout.Unwrap(), payedAmount))
	if err != nil {
		return fmt.Errorf("failed to UpdateDealConditionPayout: %v", err)
	}

	if err != nil {
		return fmt.Errorf("insertDealPayment failed: %v", err)
	}

	return nil
}

func (m *L1Processor) updateDealPayout(conn queryConn, dealID, payedAmount *big.Int, billTS uint64) error {
	deal, err := m.storage.GetDealByID(conn, dealID)
	if err != nil {
		return fmt.Errorf("failed to GetDealByID: %v", err)
	}

	newDealTotalPayout := big.NewInt(0).Add(deal.Deal.TotalPayout.Unwrap(), payedAmount)
	err = m.storage.UpdateDealPayout(conn, dealID, newDealTotalPayout, billTS)
	if err != nil {
		return fmt.Errorf("failed to updateDealPayout: %v", err)
	}

	return nil
}

func (m *L1Processor) onOrderPlaced(eventTS uint64, orderID *big.Int) error {
	order, err := m.blockchain.Market().GetOrderInfo(m.ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to GetOrderInfo: %v", err)
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	var userID common.Address
	if order.OrderType == sonm.OrderType_ASK {
		// For Ask orders, try to get this Author's masterID, use AuthorID if not found.
		userID, err = m.storage.GetMasterByWorker(conn, order.GetAuthorID().Unwrap())
		if err != nil {
			m.logger.Warn("failed to GetMasterByWorker", zap.Error(err),
				zap.String("author_id", order.GetAuthorID().Unwrap().Hex()))
			userID = order.GetAuthorID().Unwrap()
		}
	} else {
		userID = order.GetAuthorID().Unwrap()
	}

	profile, err := m.storage.GetProfileByID(conn, userID)
	if err != nil {
		certificates, _ := json.Marshal([]*sonm.Certificate{})
		profile = &sonm.Profile{
			UserID:        order.AuthorID,
			Certificates:  string(certificates),
			IdentityLevel: uint64(sonm.IdentityLevel_ANONYMOUS),
		}
	} else {
		if err := m.updateProfileStats(conn, order.OrderType, userID, 1); err != nil {
			return fmt.Errorf("failed to updateProfileStats: %v", err)
		}
	}

	if order.DealID == nil {
		order.DealID = sonm.NewBigIntFromInt(0)
	}

	err = m.storage.InsertOrder(conn, &sonm.DWHOrder{
		CreatedTS:            &sonm.Timestamp{Seconds: int64(eventTS)},
		CreatorIdentityLevel: profile.IdentityLevel,
		CreatorName:          profile.Name,
		CreatorCountry:       profile.Country,
		CreatorCertificates:  []byte(profile.Certificates),
		MasterID:             sonm.NewEthAddress(userID),
		Order: &sonm.Order{
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
		return fmt.Errorf("failed to insertOrder: %v", err)
	}

	return nil
}

func (m *L1Processor) onOrderUpdated(orderID *big.Int) error {
	marketOrder, err := m.blockchain.Market().GetOrderInfo(m.ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to GetOrderInfo: %v", err)
	}

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	// A situation is possible when user places an Ask order without specifying her `MasterID` (and we take
	// `AuthorID` for `MasterID`), and afterwards the user *does* specify her master. To avoid inconsistency,
	// we always use the user ID that was chosen in `onOrderPlaced` (i.e., the one that is already stored in DB).
	dwhOrder, err := m.storage.GetOrderByID(conn, marketOrder.GetId().Unwrap())
	if err != nil {
		return fmt.Errorf("failed to GetOrderByID: %v", err)
	}

	var userID common.Address
	if marketOrder.OrderType == sonm.OrderType_ASK {
		userID = dwhOrder.GetMasterID().Unwrap()
	} else {
		userID = marketOrder.GetAuthorID().Unwrap()
	}

	// Otherwise update order status.
	if err := m.storage.UpdateOrder(conn, marketOrder); err != nil {
		return fmt.Errorf("failed to updateOrderStatus (possibly old log entry): %v", err)
	}

	if err := m.updateProfileStats(conn, marketOrder.OrderType, userID, -1); err != nil {
		return fmt.Errorf("failed to updateProfileStats (AuthorID: `%s`): %v", marketOrder.AuthorID.Unwrap().String(), err)
	}

	return nil
}

func (m *L1Processor) onWorkerAnnounced(masterID, slaveID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertWorker(conn, masterID, slaveID); err != nil {
		return fmt.Errorf("onWorkerAnnounced failed: %v", err)
	}

	return nil
}

func (m *L1Processor) onWorkerConfirmed(masterID, slaveID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.UpdateWorker(conn, masterID, slaveID); err != nil {
		return fmt.Errorf("onWorkerConfirmed failed: %v", err)
	}

	return nil
}

func (m *L1Processor) onWorkerRemoved(masterID, slaveID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.DeleteWorker(conn, masterID, slaveID); err != nil {
		return fmt.Errorf("onWorkerRemoved failed: %v", err)
	}

	return nil
}

func (m *L1Processor) onAddedToBlacklist(adderID, addeeID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertBlacklistEntry(conn, adderID, addeeID); err != nil {
		return fmt.Errorf("onAddedToBlacklist failed: %v", err)
	}

	return nil
}

func (m *L1Processor) onRemovedFromBlacklist(removerID, removeeID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.DeleteBlacklistEntry(conn, removerID, removeeID); err != nil {
		return fmt.Errorf("onRemovedFromBlacklist failed: %v", err)
	}

	return nil
}

func (m *L1Processor) onValidatorCreated(validatorID common.Address) error {
	validator, err := m.blockchain.ProfileRegistry().GetValidator(m.ctx, validatorID)
	if err != nil {
		return fmt.Errorf("failed to get validator `%s`: %v", validatorID.String(), err)
	}

	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertOrUpdateValidator(conn, validator); err != nil {
		return fmt.Errorf("failed to insertValidator: %v", err)
	}

	return nil
}

func (m *L1Processor) onValidatorDeleted(validatorID common.Address) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.DeactivateValidator(conn, validatorID); err != nil {
		return fmt.Errorf("failed to InsertOrUpdateValidator: %v", err)
	}

	return nil
}

func (m *L1Processor) onCertificateCreated(certificateID *big.Int) error {
	certificate, err := m.blockchain.ProfileRegistry().GetCertificate(m.ctx, certificateID)
	if err != nil {
		return fmt.Errorf("failed to GetCertificate: %v", err)
	}
	certificate.IdentityLevel = (certificate.Attribute / uint64(100)) % 10

	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	// Check if this certificate is assigned to a validator.
	validatorLevel, err := m.blockchain.ProfileRegistry().GetValidatorLevel(m.ctx, certificate.OwnerID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to GetValidatorLevel: %v", err)
	}
	if validatorLevel != 0 {
		// It's a validator.
		return m.storage.UpdateValidator(
			conn, certificate.OwnerID.Unwrap(), certificate.GetAttributeNameNormalized(), certificate.GetValue())
	}

	if err = m.storage.InsertCertificate(conn, certificate); err != nil {
		return fmt.Errorf("failed to insertCertificate: %v", err)
	}

	if err := m.updateProfile(conn, certificate); err != nil {
		return fmt.Errorf("failed to updateProfile: %v", err)
	}

	if err := m.updateEntitiesByProfile(conn, certificate.OwnerID.Unwrap()); err != nil {
		return fmt.Errorf("failed to updateEntitiesByProfile: %v", err)
	}

	return nil
}

func (m *L1Processor) updateProfile(conn queryConn, cert *sonm.Certificate) error {
	_, activeAsks, err := m.storage.GetOrders(conn, &sonm.OrdersRequest{
		Type:      sonm.OrderType_ASK,
		MasterID:  cert.OwnerID,
		WithCount: true})
	if err != nil {
		return fmt.Errorf("failed to get active ASKs count: %v", err)
	}

	_, activeBids, err := m.storage.GetOrders(conn, &sonm.OrdersRequest{
		Type:      sonm.OrderType_BID,
		MasterID:  cert.OwnerID,
		WithCount: true})
	if err != nil {
		return fmt.Errorf("failed to get active BIDs count: %v", err)
	}

	profile, err := m.storage.GetProfileByID(conn, cert.OwnerID.Unwrap())
	if err != nil {
		certBytes, _ := json.Marshal([]*sonm.Certificate{cert})
		profile = &sonm.Profile{
			UserID:        cert.OwnerID,
			Certificates:  string(certBytes),
			ActiveAsks:    activeAsks,
			ActiveBids:    activeBids,
			IdentityLevel: cert.IdentityLevel,
		}
		if err = m.storage.InsertProfileUserID(conn, profile); err != nil {
			return fmt.Errorf("failed to insertProfileUserID: %v", err)
		}
	}

	// Update distinct Profile columns.
	switch cert.Attribute {
	case CertificateName, CertificateCountry:
		err := m.storage.UpdateProfile(conn, cert.OwnerID.Unwrap(), attributeToString[cert.Attribute],
			string(cert.Value))
		if err != nil {
			return fmt.Errorf("failed to UpdateProfile (%s): %v err", attributeToString[cert.Attribute], err)
		}
	}

	if cert.IdentityLevel > profile.IdentityLevel {
		profile.IdentityLevel = cert.IdentityLevel
	}
	err = m.storage.UpdateProfile(conn, cert.OwnerID.Unwrap(), "IdentityLevel", profile.IdentityLevel)
	if err != nil {
		return fmt.Errorf("failed to updateProfileCertificates (Level): %v", err)
	}

	return nil
}

func (m *L1Processor) updateEntitiesByProfile(conn queryConn, userID common.Address) error {
	profile, err := m.storage.GetProfileByID(conn, userID)
	if err != nil {
		return fmt.Errorf("failed to getProfileInfo: %v", err)
	}

	if err := m.storage.UpdateOrders(conn, profile); err != nil {
		return fmt.Errorf("failed to updateOrders: %v", err)
	}

	if err = m.storage.UpdateDealsSupplier(conn, profile); err != nil {
		return fmt.Errorf("failed to updateDealsSupplier: %v", err)
	}

	err = m.storage.UpdateDealsConsumer(conn, profile)
	if err != nil {
		return fmt.Errorf("failed to updateDealsConsumer: %v", err)
	}

	return nil
}

func (m *L1Processor) updateProfileStats(conn queryConn, orderType sonm.OrderType, userID common.Address, update int) error {
	var field string
	if orderType == sonm.OrderType_ASK {
		field = "ActiveAsks"
	} else {
		field = "ActiveBids"
	}

	if err := m.storage.UpdateProfileStats(conn, userID, field, update); err != nil {
		return fmt.Errorf("failed to UpdateProfileStats: %v", err)
	}

	return nil
}

func (m *L1Processor) onCertificateUpdated(certID *big.Int) error {
	conn, err := newTxConn(m.db, m.logger)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer conn.Finish()

	cert, err := m.storage.GetCertificate(conn, certID)
	if err != nil {
		return fmt.Errorf("failed to GetCertificate: %v", err)
	}

	profileLevel, err := m.blockchain.ProfileRegistry().GetProfileLevel(m.ctx, cert.OwnerID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to GetProfileLevel: %v", err)
	}

	if err := m.storage.DeleteCertificate(conn, certID); err != nil {
		return fmt.Errorf("failed to DeleteCertificate: %v", err)
	}

	err = m.storage.UpdateProfile(conn, cert.OwnerID.Unwrap(), "IdentityLevel", profileLevel)
	if err != nil {
		return fmt.Errorf("failed to updateProfileCertificates (Level): %v", err)
	}

	if err := m.updateEntitiesByProfile(conn, cert.OwnerID.Unwrap()); err != nil {
		return fmt.Errorf("failed to updateEntitiesByProfile: %v", err)
	}

	return nil
}

// coldStart waits till last seen block number gets to `w.cfg.ColdStart.UpToBlock` and then tries to create indices.
func (m *L1Processor) coldStart() error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	targetBlock, err := m.blockchain.Events().GetLastBlock(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to GetLastBlock: %v", err)
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

func (m *L1Processor) maybeCreateIndices(targetBlock uint64) (targetBlockReached bool, err error) {
	lastEvent, err := m.getLastEvent()
	if err != nil {
		return false, err
	}

	m.logger.Info("current block (waiting to CreateIndices)", zap.Uint64("block_number", lastEvent.BlockNumber))
	if lastEvent.BlockNumber >= targetBlock {
		if err := m.storage.CreateIndices(m.db); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

func (m *L1Processor) getLastEvent() (*blockchain.Event, error) {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	return m.storage.GetLastEvent(conn)
}

func (m *L1Processor) updateLastEvent(event *blockchain.Event) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.UpdateLastEvent(conn, event); err != nil {
		return fmt.Errorf("failed to updateLastEvent: %v", err)
	}

	return nil
}

func (m *L1Processor) insertLastEvent(event *blockchain.Event) error {
	conn := newSimpleConn(m.db)
	defer conn.Finish()

	if err := m.storage.InsertLastEvent(conn, event); err != nil {
		return fmt.Errorf("failed to updateLastEvent: %v", err)
	}

	return nil
}

func (m *L1Processor) updateDealConditionEndTime(conn queryConn, dealID *sonm.BigInt, eventTS uint64) error {
	dealConditions, _, err := m.storage.GetDealConditions(conn, &sonm.DealConditionsRequest{DealID: dealID})
	if err != nil {
		return fmt.Errorf("failed to getDealConditions: %v", err)
	}

	dealCondition := dealConditions[0]
	if err := m.storage.UpdateDealConditionEndTime(conn, dealCondition.Id, eventTS); err != nil {
		return fmt.Errorf("failed to update DealCondition: %v", err)
	}

	return nil
}

func (m *L1Processor) saveLastEvent() {
	if err := m.updateLastEvent(m.lastEvent); err != nil {
		m.logger.Warn("failed to updateLastEvent", zap.Error(err))
	}
}

type eventsDispatcher struct {
	logger                    *zap.Logger
	NumBenchmarksUpdated      []*blockchain.Event
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
	CertificatesUpdated       []*blockchain.Event
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
	case *blockchain.NumBenchmarksUpdatedData:
		m.NumBenchmarksUpdated = append(m.NumBenchmarksUpdated, event)
	case *blockchain.ValidatorCreatedData:
		m.ValidatorsCreated = append(m.ValidatorsCreated, event)
	case *blockchain.ValidatorDeletedData:
		m.ValidatorsDeleted = append(m.ValidatorsDeleted, event)
	case *blockchain.CertificateCreatedData:
		m.CertificatesCreated = append(m.CertificatesCreated, event)
	case *blockchain.CertificateUpdatedData:
		m.CertificatesUpdated = append(m.CertificatesUpdated, event)
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
