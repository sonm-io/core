package hub

import (
	"context"
	"sync"

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

type AskPlansData map[string]AskPlan

type AskPlans struct {
	Data   AskPlansData
	mu     sync.Mutex
	hub    *Hub
	market pb.MarketClient
}

func NewAskPlans(hub *Hub, market pb.MarketClient) *AskPlans {
	return &AskPlans{
		Data:   make(map[string]AskPlan),
		hub:    hub,
		market: market,
	}
}

func (a *AskPlans) Run(ctx context.Context) error {
	ticker := util.NewImmediateTicker(a.hub.cfg.Market.UpdatePeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.checkAnnounces(ctx)
		}
	}
}

func (a *AskPlans) Add(ctx context.Context, order *structs.Order) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	id := uuid.New()
	plan := AskPlan{
		Id:    id,
		Order: order,
	}
	a.Data[id] = plan
	if a.hub.HasResources(plan.Order.GetSlot().GetResources()) {
		a.announcePlan(ctx, &plan)
	}
	a.sync(ctx)
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

func (a *AskPlans) Remove(ctx context.Context, planId string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	askPlan, ok := a.Data[planId]
	if !ok {
		return errSlotNotExists
	}
	if askPlan.Order.Id != "" {
		a.deannouncePlan(ctx, &askPlan)
	}
	delete(a.Data, planId)
	a.sync(ctx)
	return nil
}

func (a *AskPlans) HasOrder(orderId string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	//TODO: not very efficient, maybe we can hold another index by market orderId, but now it looks like overkill
	for _, plan := range a.Data {
		if plan.Order.Id == orderId {
			return true
		}
	}
	return false
}

func (a *AskPlans) forceRenewAnnounces(ctx context.Context) {
	for _, plan := range a.Data {
		if a.hub.HasResources(plan.Order.GetSlot().GetResources()) {
			a.announcePlan(ctx, &plan)
		} else {
			a.deannouncePlan(ctx, &plan)
		}
	}
}

func (a *AskPlans) checkAnnounces(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	changed := false
	toUpdate := make([]string, 0)
	for _, plan := range a.Data {

		has := a.hub.HasResources(plan.Order.GetSlot().GetResources())
		announced := plan.Order.Id != ""
		if has && !announced {
			changed = true
			a.announcePlan(ctx, &plan)
		}
		if !has && announced {
			changed = true
			a.deannouncePlan(ctx, &plan)
		}
		if has && announced {
			toUpdate = append(toUpdate, plan.Order.Id)
		}
	}
	if len(toUpdate) > 0 {
		_, err := a.market.TouchOrders(ctx, &pb.TouchOrdersRequest{IDs: toUpdate})
		if err != nil {
			log.G(ctx).Warn("failed to touch orders on market, forcing renewing announces", zap.Error(err))
			a.forceRenewAnnounces(ctx)
		}
	}
	if changed {
		a.sync(ctx)
	}
}

//TODO: do we need to signal about error?
func (a *AskPlans) announcePlan(ctx context.Context, plan *AskPlan) {
	createdOrder, err := a.market.CreateOrder(ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(ctx).Warnf("failed to announce ask plan with id{} on market - {}", plan.Id, zap.Error(err))
		return
	}
	wrappedOrder, err := structs.NewOrder(createdOrder)
	if err != nil {
		log.S(ctx).Warnf("invalid order received from market - {}", plan.Id, zap.Error(err))
		return
	}
	plan.Order = wrappedOrder
}

func (a *AskPlans) deannouncePlan(ctx context.Context, plan *AskPlan) {
	_, err := a.market.CancelOrder(ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(ctx).Warnf("failed to deannounce order {} (ask plan - {}) on market - {}", plan.Order.Id, plan.Id, zap.Error(err))
	} else {
		plan.Order.Id = ""
	}
}

func (a *AskPlans) sync(ctx context.Context) {
	if err := a.hub.SynchronizeAskPlans(a.Data); err != nil {
		log.G(ctx).Warn("failed to sync ask plans to cluster", zap.Error(err))
	}
}
