package hub

import (
	"context"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type AskPlan struct {
	Id    string
	Order *structs.Order
}

type AskPlansData map[string]*AskPlan

type AskPlans struct {
	Data   AskPlansData
	ctx    context.Context
	mu     sync.Mutex
	hub    *Hub
	market pb.MarketClient
}

func NewAskPlans(ctx context.Context, hub *Hub, market pb.MarketClient) *AskPlans {
	askPlans := AskPlans{
		Data:   make(map[string]*AskPlan),
		ctx:    ctx,
		hub:    hub,
		market: market,
	}
	return &askPlans
}

func (a *AskPlans) Run() error {
	a.hub.cfg.Market
	ticker := util.NewImmediateTicker(time.Second)
	for {
		select {
		case <-a.ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.checkAnnounces(); err != nil {
				return err
			}
		}
	}
}

func (a *AskPlans) Add(order *structs.Order) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := uuid.New()
	a.Data[id] = &AskPlan{
		Id:    id,
		Order: order,
	}
	return id, nil
}

func (a *AskPlans) DumpSlots() map[string]*pb.Slot {
	result := make(map[string]*pb.Slot)
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, plan := range a.Data {
		result[id] = plan.Order.Slot
	}
	return result
}

func (a *AskPlans) Dump() AskPlansData {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Data
}

func (a *AskPlans) RestoreFrom(data AskPlansData) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Data = data
}

func (a *AskPlans) Remove(planId string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	askPlan, ok := a.Data[planId]
	if !ok {
		return errSlotNotExists
	}
	a.deannouncePlan(askPlan)
	delete(a.Data, planId)
	a.sync()
	return nil
}

func (a *AskPlans) HasOrder(orderId string) bool {
	panic("unimplemented")
}

func (a *AskPlans) checkAnnounces() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	changed := false
	toUpdate := make([]string, 0)
	for _, plan := range a.Data {

		has := a.hub.HasResources(plan.Order.GetSlot().GetResources())
		announced := plan.Order.Id != ""
		if has && !announced {
			changed = true
			a.announcePlan(plan)
		}
		if !has && announced {
			changed = true
			a.deannouncePlan(plan)
		}
		if has && announced {
			toUpdate = append(toUpdate, plan.Order.Id)
		}
	}
	if len(toUpdate) > 0 {
		_, err := a.market.TouchOrders(a.ctx, &pb.TouchOrdersRequest{IDs: toUpdate})
		if err != nil {
			log.G(a.ctx).Warn("failed to touch orders on market", zap.Error(err))
		}
	}
	if changed {
		a.sync()
	}
	return nil
}

//TODO: do we need to signal about error?
func (a *AskPlans) announcePlan(plan *AskPlan) {
	createdOrder, err := a.market.CreateOrder(a.ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(a.ctx).Warnf("failed to announce ask plan with id{} on market - {}", plan.Id, zap.Error(err))
		return
	}
	wrappedOrder, err := structs.NewOrder(createdOrder)
	if err != nil {
		log.S(a.ctx).Warnf("invalid order received from market - {}", plan.Id, zap.Error(err))
		return
	}
	plan.Order = wrappedOrder
}

func (a *AskPlans) deannouncePlan(plan *AskPlan) {
	_, err := a.market.CancelOrder(a.ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(a.ctx).Warnf("failed to deannounce order {} (ask plan - {}) on market - {}", plan.Order.Id, plan.Id, zap.Error(err))
	} else {
		plan.Order.Id = ""
	}
}

func (a *AskPlans) sync() {
	if err := a.hub.SynchronizeAskPlans(a.Data); err != nil {
		log.G(a.ctx).Warn("failed to sync ask plans to cluster", zap.Error(err))
	}
}
