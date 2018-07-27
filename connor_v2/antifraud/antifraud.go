package antifraud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AntiFraud interface {
	Run(ctx context.Context) error
	DealOpened(deal *sonm.Deal) error
	//DealClosed(deal *sonm.Deal)
	TrackTask(ctx context.Context, deal *sonm.Deal, taskID string) error
	//TaskDead(dealID *sonm.BigInt, taskID string)
	FinishDeal(deal *sonm.Deal) error
}

type LogProcessor interface {
	Run(ctx context.Context) error
	TaskQuality() (accurate bool, quality float64)
}

type dealMeta struct {
	deal         *sonm.Deal
	logProcessor LogProcessor
	// todo:  count different causes of failures
}

func lifeTime(deal *sonm.Deal) time.Duration {
	return time.Now().Sub(deal.GetStartTime().Unix())
}

type antiFraud struct {
	mu                sync.RWMutex
	cfg               Config
	meta              map[string]*dealMeta
	blacklistWatchers map[common.Address]*blacklistWatcher
	nodeConnection    *grpc.ClientConn
	deals             sonm.DealManagementClient
	log               *zap.Logger
}

func NewAntiFraud(log *zap.Logger, nodeConnection *grpc.ClientConn) AntiFraud {
	return &antiFraud{
		meta:              make(map[string]*dealMeta),
		blacklistWatchers: map[common.Address]*blacklistWatcher{},
		nodeConnection:    nodeConnection,
		deals:             sonm.NewDealManagementClient(nodeConnection),
		log:               log,
	}
}

// This blocks until context is cancelled or unrecoverable error met
func (m *antiFraud) Run(ctx context.Context) error {
	m.log.Info("starting antifraud")
	//TODO: cfg
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			m.checkDeals(ctx)
		}
	}
}

//TODO: async
func (m *antiFraud) checkDeals(ctx context.Context) error {
	m.log.Debug("checking deals")
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, dealMeta := range m.meta {
		if dealMeta.logProcessor == nil {
			m.log.Debug("skipping deal without task")
			continue
		}

		accurate, quality := dealMeta.logProcessor.TaskQuality()
		if !accurate {
			m.log.Debug("skipping inaccurate quality", zap.Float64("value", quality))
			continue
		}

		// TODO(sshaman1101): config
		requiredTaskQuality := 0.9
		if quality < requiredTaskQuality {
			m.log.Warn("task quality is less that required, closing deal",
				zap.Float64("calculated", quality), zap.Float64("required", requiredTaskQuality))

			watcher, ok := m.blacklistWatchers[dealMeta.deal.SupplierID.Unwrap()]
			if !ok {
				m.log.Warn("cannot obtain blacklist watcher for deal, skipping")
				continue
			}

			watcher.Failure()
			m.finishDeal(dealMeta.deal, sonm.BlacklistType_BLACKLIST_WORKER)
		} else {
			m.log.Debug("task quality is fit into required required value", zap.Float64("quality", quality))
			m.blacklistWatchers[dealMeta.deal.SupplierID.Unwrap()].Success()
		}
	}

	//TODO: save this in DB, load on start
	for _, watcher := range m.blacklistWatchers {
		watcher.TryUnblacklist(ctx)
	}
	return nil
}

func (m *antiFraud) TrackTask(ctx context.Context, deal *sonm.Deal, taskID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	meta, ok := m.meta[deal.Id.Unwrap().String()]
	if !ok {
		return fmt.Errorf("could not register spawned task %s, no deal with id %s", taskID, deal.Id.Unwrap().String())
	}

	meta.logProcessor = NewLogProcessor(m.cfg.LogProcessorConfig, m.log, m.nodeConnection, deal, taskID)
	return meta.logProcessor.Run(ctx)
}

func (m *antiFraud) DealOpened(deal *sonm.Deal) error {
	m.log.Info("registering deal", zap.String("deal_id", deal.GetId().Unwrap().String()))

	meta := &dealMeta{
		deal: deal,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.meta[deal.GetId().Unwrap().String()] = meta
	if _, ok := m.blacklistWatchers[deal.GetSupplierID().Unwrap()]; !ok {
		//TODO: constructor
		m.blacklistWatchers[deal.GetSupplierID().Unwrap()] = &blacklistWatcher{
			address:     deal.GetSupplierID().Unwrap(),
			currentStep: minStep,
			client:      sonm.NewBlacklistClient(m.nodeConnection),
		}
	}
	return nil
}

func (m *antiFraud) FinishDeal(deal *sonm.Deal) error {
	return m.finishDeal(deal, sonm.BlacklistType_BLACKLIST_NOBODY)
}

func (m *antiFraud) finishDeal(deal *sonm.Deal, blacklistType sonm.BlacklistType) error {
	m.log.Info("finishing deal", zap.String("deal_id", deal.GetId().Unwrap().String()),
		zap.Duration("lifetime", lifeTime(deal)))

	//TODO: do we need timeout here? at least we may make it longer and perform in async way
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := m.deals.Finish(ctx, &sonm.DealFinishRequest{
		Id:            deal.GetId(),
		BlacklistType: blacklistType,
	})

	return err
}
