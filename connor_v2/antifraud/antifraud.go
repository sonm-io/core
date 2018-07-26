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
	RegisterDeal(deal *sonm.Deal)
	// RegisterTask(dealID *sonm.BigInt, taskID string)
	FinishDeal(dealID *sonm.BigInt) error
}

type dealMeta struct {
	ID        *sonm.BigInt
	Supplier  common.Address
	CreatedAt time.Time
	// todo:  count different causes of failures
}

func (m *dealMeta) lifeTime() time.Duration {
	return time.Now().Sub(m.CreatedAt)
}

func (m *dealMeta) blacklistType() sonm.BlacklistType {
	return sonm.BlacklistType_BLACKLIST_NOBODY
}

type antiFraud struct {
	mu    sync.Mutex
	meta  map[string]*dealMeta
	deals sonm.DealManagementClient
	log   *zap.Logger
}

func NewAntiFraud(log *zap.Logger, cc *grpc.ClientConn) AntiFraud {
	return &antiFraud{
		log:   log,
		deals: sonm.NewDealManagementClient(cc),
		meta:  make(map[string]*dealMeta),
	}
}

func (af *antiFraud) RegisterDeal(deal *sonm.Deal) {
	af.mu.Lock()
	defer af.mu.Unlock()

	af.log.Info("registering deal", zap.String("deal_id", deal.GetId().Unwrap().String()))

	meta := &dealMeta{
		ID:        deal.GetId(),
		Supplier:  deal.GetSupplierID().Unwrap(),
		CreatedAt: time.Now(),
	}

	af.meta[deal.GetId().Unwrap().String()] = meta
}

func (af *antiFraud) FinishDeal(dealID *sonm.BigInt) error {
	af.mu.Lock()
	defer af.mu.Unlock()

	id := dealID.Unwrap().String()
	deal, ok := af.meta[id]
	if !ok {
		return fmt.Errorf("deal with id = %s is not registred into AntiFraud module", id)
	}

	af.log.Info("finishing deal", zap.String("deal_id", deal.ID.Unwrap().String()),
		zap.Duration("lifetime", deal.lifeTime()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := af.deals.Finish(ctx, &sonm.DealFinishRequest{
		Id:            deal.ID,
		BlacklistType: deal.blacklistType(),
	})

	return err
}
