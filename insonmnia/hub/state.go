package hub

import (
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/resource"
	"go.uber.org/zap"
)

type hubState struct {
	Tasks    map[string]*TaskInfo
	Devices  map[string]DeviceProperties
	AskPlans AskPlansData
	Acl      []string
	Deals    map[DealID]*DealMeta
	Orders   map[OrderID]ReservedOrder
	Miners   map[string]map[OrderID]resource.Resources
}

func (h *Hub) dumpState() error {
	state := h.getState()

	if err := h.cluster.Synchronize(state); err != nil {
		return err
	}

	return nil
}

func (h *Hub) loadState(state *hubState) error {
	h.orderShelter.mu.Lock()
	defer h.orderShelter.mu.Unlock()

	h.askPlans.mu.Lock()
	defer h.askPlans.mu.Unlock()

	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	h.associatedHubsMu.Lock()
	defer h.associatedHubsMu.Unlock()

	h.devicePropertiesMu.Lock()
	defer h.devicePropertiesMu.Unlock()

	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	h.tasks = state.Tasks
	h.deviceProperties = state.Devices
	h.askPlans.Load(state.AskPlans)
	h.acl.Load(state.Acl)
	h.deals = state.Deals
	h.restoreResourceUsage()
	h.orderShelter.Load(state.Orders)

	for minerID, usageMapping := range state.Miners {
		minerCtx, ok := h.miners[minerID]
		if !ok {
			log.G(h.ctx).Error("found unknown miner, skipping", zap.String("miner_id", minerID))
			if h.cluster.IsLeader() {
				// TODO: maybe we should close deals and release resources if this happens.
			}
		} else {
			minerCtx.mu.Lock()
			minerCtx.usageMapping = usageMapping
			minerCtx.mu.Unlock()
		}
	}

	return nil
}

func (h *Hub) getState() *hubState {
	h.orderShelter.mu.Lock()
	defer h.orderShelter.mu.Unlock()

	h.askPlans.mu.Lock()
	defer h.askPlans.mu.Unlock()

	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	h.associatedHubsMu.Lock()
	defer h.associatedHubsMu.Unlock()

	h.devicePropertiesMu.Lock()
	defer h.devicePropertiesMu.Unlock()

	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	state := &hubState{
		Tasks:    h.tasks,
		Devices:  h.deviceProperties,
		AskPlans: h.askPlans.Dump(),
		Acl:      h.acl.Dump(),
		Deals:    h.deals,
		Orders:   h.orderShelter.Dump(),
		Miners:   make(map[string]map[OrderID]resource.Resources),
	}

	for minerID, minerCtx := range h.miners {
		minerCtx.mu.Lock()
		state.Miners[minerID] = minerCtx.usageMapping
		minerCtx.mu.Unlock()
	}

	return state
}
