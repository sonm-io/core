package hub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReservedOrder struct {
	OrderID          OrderID
	MinerID          string
	EthAddr          common.Address
	ReservedFrom     time.Time
	ReservedDuration time.Duration
}

func (o *ReservedOrder) IsExpired() bool {
	return time.Now().Sub(o.ReservedFrom) >= o.ReservedDuration
}

type askPlan struct {
	ID    string
	Order *structs.Order
}

type stateJSON struct {
	Acl              *workerACLStorage           `json:"acl"`
	Deals            map[DealID]*DealMeta        `json:"deals"`
	Tasks            map[string]*TaskInfo        `json:"tasks"`
	Miner            *MinerCtx                   `json:"miner"`
	Orders           map[OrderID]ReservedOrder   `json:"orders"`
	AskPlans         map[string]*askPlan         `json:"ask_plans"`
	DeviceProperties map[string]DeviceProperties `json:"device_properties"`
}

type state struct {
	mu      sync.Mutex
	ctx     context.Context
	eth     ETH
	cluster Cluster
	market  pb.MarketClient

	deals            map[DealID]*DealMeta
	tasks            map[string]*TaskInfo
	minerCtx         *MinerCtx
	orders           map[OrderID]ReservedOrder
	askPlans         map[string]*askPlan
	deviceProperties map[string]DeviceProperties
}

func newState(ctx context.Context, eth ETH, market pb.MarketClient, cluster Cluster, minerCtx *MinerCtx) (
	*state, error) {
	out := &state{
		ctx:     ctx,
		eth:     eth,
		cluster: cluster,
		market:  market,

		deals:            make(map[DealID]*DealMeta),
		tasks:            make(map[string]*TaskInfo),
		minerCtx:         minerCtx,
		orders:           make(map[OrderID]ReservedOrder, 0),
		askPlans:         make(map[string]*askPlan, 0),
		deviceProperties: make(map[string]DeviceProperties),
	}

	if err := out.init(minerCtx); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *state) Dump() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.dump()
}

func (s *state) dump() error {
	sJSON := &stateJSON{
		Deals:            s.deals,
		Tasks:            s.tasks,
		Miner:            s.minerCtx,
		Orders:           s.orders,
		AskPlans:         s.askPlans,
		DeviceProperties: s.deviceProperties,
	}

	return s.cluster.Synchronize(sJSON)
}

func (s *state) Load(other *stateJSON) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.load(other)
}

func (s *state) load(other *stateJSON) error {
	s.deals = other.Deals
	s.tasks = other.Tasks
	s.orders = other.Orders
	s.askPlans = other.AskPlans
	s.deviceProperties = other.DeviceProperties
	s.minerCtx.usageMapping = other.Miner.usageMapping
	s.restoreResourceUsage()

	return nil
}

func (s *state) init(minerCtx *MinerCtx) error {
	sJSON := &stateJSON{
		Deals:            make(map[DealID]*DealMeta),
		Tasks:            make(map[string]*TaskInfo),
		Miner:            minerCtx,
		Orders:           make(map[OrderID]ReservedOrder, 0),
		AskPlans:         make(map[string]*askPlan, 0),
		DeviceProperties: make(map[string]DeviceProperties),
	}

	if err := s.cluster.RegisterAndLoadEntity("state", sJSON); err != nil {
		return err
	}

	return s.load(sJSON)
}

func (s *state) RunMonitoring(ctx context.Context) error {
	timer := time.NewTicker(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			if err := s.monitoringStep(); err != nil {
				return err
			}
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *state) monitoringStep() error {
	if err := s.checkAcceptedDealsTS(); err != nil {
		log.G(s.ctx).Error("failed to check accepted deals", zap.Error(err))
		return err
	}

	if err := s.checkClosedDealsTS(); err != nil {
		log.G(s.ctx).Error("failed to check closed deals", zap.Error(err))
		return err
	}

	if err := s.checkAnnouncesTS(); err != nil {
		log.G(s.ctx).Error("failed to check announces", zap.Error(err))
		return err
	}

	if err := s.checkOrdersTS(); err != nil {
		log.G(s.ctx).Error("failed to check orders", zap.Error(err))
		return err
	}

	s.closeExpiredDealsTS()

	s.mu.Lock()
	if err := s.dump(); err != nil {
		log.G(s.ctx).Error("failed to dump state", zap.Error(err))
	}
	s.mu.Unlock()

	return nil
}

// Synchronized by `s.mu`.
func (s *state) checkAcceptedDealsTS() error {
	log.G(s.ctx).Info("checking accepted deals")
	acceptedDeals, err := s.eth.GetAcceptedDeals(s.ctx)
	if err != nil {
		log.G(s.ctx).Warn("failed to fetch accepted deals from the Blockchain", zap.Error(err))
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, acceptedDeal := range acceptedDeals {
		deal, ok := s.deals[DealID(acceptedDeal.Id)]
		if !ok {
			continue
		}

		// Update deal expiration time according to the contract.
		deal.EndTime = acceptedDeal.EndTime.Unix()
	}

	return nil
}

// WatchDealsClosed watches ETH for closed deals.
// Synchronized by `s.mu`.
func (s *state) checkClosedDealsTS() error {
	log.G(s.ctx).Debug("checking closed deals")
	closedDeals, err := s.eth.GetClosedDeals(s.ctx)
	if err != nil {
		log.G(s.ctx).Warn("failed to fetch closed deals from the Blockchain", zap.Error(err))
		return nil
	}

	s.mu.Lock()

	ordersToRepublish := map[DealID]OrderID{}
	for _, closedDeal := range closedDeals {
		dealID := DealID(closedDeal.Id)
		deal, ok := s.deals[dealID]
		if !ok {
			continue
		}

		orderID := OrderID(deal.Order.GetID())

		if err := s.releaseDeal(dealID); err != nil {
			log.G(s.ctx).Error("failed to release deal resources",
				zap.Stringer("dealID", dealID),
				zap.Stringer("orderID", orderID),
				zap.Error(err),
			)
			continue
		}

		s.minerCtx.Release(orderID)
	}

	s.mu.Unlock()

	for dealID, orderID := range ordersToRepublish {
		if err := s.publishOrder(orderID); err != nil {
			log.G(s.ctx).Error("failed to republish order on a market",
				zap.Stringer("dealID", dealID),
				zap.Stringer("orderID", orderID),
				zap.Error(err),
			)
		}
	}

	return nil
}

// Synchronized by `s.mu`.
func (s *state) checkAnnouncesTS() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.G(s.ctx).Debug("checking announces")
	var toUpdate = make([]string, 0)
	for _, plan := range s.askPlans {
		has := s.hasResources(plan.Order.GetSlot().GetResources())
		announced := plan.Order.Id != ""
		if has && !announced {
			log.S(s.ctx).Debugf("hub has enough resources for ask-plan %s, announcing", plan.ID)
			s.announcePlan(s.ctx, plan)
		}
		if !has && announced {
			log.S(s.ctx).Debugf("hub lacks resources for ask-plan %s, deannouncing", plan.ID)
			s.deannouncePlan(s.ctx, plan)
		}
		if has && announced {
			log.S(s.ctx).Debugf("hub has enough resources for ask-plan %s, will touch corresponding order %s",
				plan.ID, plan.Order.Id)
			toUpdate = append(toUpdate, plan.Order.Id)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.market.TouchOrders(s.ctx, &pb.TouchOrdersRequest{IDs: toUpdate})
		if err != nil {
			log.G(s.ctx).Warn("failed to touch orders on market, forcing renewing announces", zap.Error(err))
			for _, plan := range s.askPlans {
				if s.hasResources(plan.Order.GetSlot().GetResources()) {
					s.announcePlan(s.ctx, plan)
				} else {
					s.deannouncePlan(s.ctx, plan)
				}
			}
		}
	}

	return nil
}

// Synchronized by `s.mu`.
func (s *state) checkOrdersTS() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.G(s.ctx).Debug("checking orders")
	renewedOrders := make(map[OrderID]ReservedOrder, 0)
	for orderID, orderInfo := range s.orders {
		if orderInfo.IsExpired() {
			s.minerCtx.Release(orderID)
			log.G(s.ctx).Info("releasing order due to timeout", zap.Stringer("orderID", orderID), zap.String("minerID", orderInfo.MinerID))
		} else {
			renewedOrders[orderID] = orderInfo
		}
	}

	s.orders = renewedOrders

	return nil
}

func (s *state) closeExpiredDealsTS() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for dealID, dealMeta := range s.deals {
		if now.After(dealMeta.EndTime) {
			if err := s.eth.CloseDeal(dealID); err != nil {
				log.G(s.ctx).Error("failed to close deal using blockchain API",
					zap.Stringer("dealID", dealID),
					zap.Error(err),
				)
			}
		}
	}
}

func (s *state) GetTaskInfo(dealID, taskID string) (*TaskInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, ok := s.deals[DealID(dealID)]
	if !ok {
		return nil, errDealNotFound
	}

	for _, task := range tasks.Tasks {
		if task.ID == taskID {
			return task, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "task not found")
}

func (s *state) GetDealMeta(dealID DealID) (*DealMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getDealMeta(dealID)
}

func (s *state) getDealMeta(dealID DealID) (*DealMeta, error) {
	meta, ok := s.deals[dealID]
	if !ok {
		return nil, errDealNotFound
	}

	return meta, nil
}

func (s *state) PopDealHistory(dealID DealID) ([]*TaskInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.popDealHistory(dealID)
}

func (s *state) popDealHistory(dealID DealID) ([]*TaskInfo, error) {
	tasks, ok := s.deals[dealID]
	if !ok {
		return nil, errDealNotFound
	}

	delete(s.deals, dealID)
	dealsGauge.Dec()

	return tasks.Tasks, nil
}

func (s *state) GetTaskList(ctx context.Context) (*pb.TaskListReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reply := &pb.TaskListReply{Info: map[string]*pb.TaskListReply_TaskInfo{}}

	// iter over known tasks
	for taskID := range s.tasks {
		info := &pb.TaskListReply_TaskInfo{Tasks: map[string]*pb.TaskStatusReply{
			taskID: {Status: pb.TaskStatusReply_UNKNOWN},
		}}

		taskDetails, err := s.minerCtx.miner.TaskDetails(ctx, &pb.ID{Id: taskID})
		if err == nil {
			info.Tasks[taskID] = taskDetails
		}

		reply.Info["self"] = info
	}

	return reply, nil
}

func (s *state) HasOrder(orderID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, plan := range s.askPlans {
		if plan.Order.Id == orderID {
			return true
		}
	}

	return false
}

func (s *state) ReserveOrder(orderID OrderID, minerID string, ethAddr common.Address, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.orderExists(orderID) {
		return fmt.Errorf("order already reserved")
	}

	s.orders[orderID] = ReservedOrder{
		OrderID:          orderID,
		MinerID:          minerID,
		EthAddr:          ethAddr,
		ReservedFrom:     now,
		ReservedDuration: duration,
	}

	return nil
}

func (s *state) OrderExists(orderID OrderID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.orderExists(orderID)
}

func (s *state) orderExists(orderID OrderID) bool {
	_, ok := s.orders[orderID]

	return ok
}

// Commit commits the specified reserved order, removing it from the shelter.
// Note, that this method does not release resources from the miner's tracker,
// because using it means that the resource's lifetime was prolonged by
// accepting a deal.
func (s *state) CommitOrder(orderID OrderID) (ReservedOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return ReservedOrder{}, fmt.Errorf("order not found")
	}

	delete(s.orders, orderID)

	return order, nil
}

func (s *state) PollCommitOrder(orderID OrderID, ethAddr common.Address) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found")
	}

	if auth.EqualAddresses(order.EthAddr, ethAddr) {
		return nil
	} else {
		return fmt.Errorf("order %s cannot be commited by %s", orderID, ethAddr)
	}
}

func (s *state) SetDealMeta(dealMeta *DealMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deals[dealMeta.ID] = dealMeta
}

func (s *state) IsTaskFinished(taskID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.isTaskFinished(taskID)
}

func (s *state) isTaskFinished(taskID string) bool {
	_, ok := s.tasks[taskID]
	return !ok
}

func (s *state) GetDeviceProperties(id string) (*pb.GetDevicePropertiesReply, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	properties := s.deviceProperties[id]
	return &pb.GetDevicePropertiesReply{Properties: properties}, nil
}

func (s *state) SetDeviceProperties(id string, properties map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deviceProperties[id] = DeviceProperties(properties)
}

func (s *state) DumpSlots() map[string]*pb.Slot {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]*pb.Slot)
	for id, plan := range s.askPlans {
		result[id] = plan.Order.Slot
	}

	return result
}

func (s *state) AddSlot(ctx context.Context, order *structs.Order) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var (
		id   = uuid.New()
		plan = askPlan{ID: id, Order: order}
	)
	s.askPlans[id] = &plan
	if s.hasResources(plan.Order.GetSlot().GetResources()) {
		s.announcePlan(ctx, &plan)
	}

	return id, nil
}

func (s *state) RemoveSlot(ctx context.Context, planID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	askPlan, ok := s.askPlans[planID]
	if !ok {
		return errSlotNotExists
	}

	if askPlan.Order.Id != "" {
		s.deannouncePlan(ctx, askPlan)
	}
	delete(s.askPlans, planID)

	return nil
}

func (s *state) GetTaskByID(taskID string) (*TaskInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getTaskByID(taskID)
}

func (s *state) getTaskByID(taskID string) (*TaskInfo, bool) {
	taskInfo, ok := s.tasks[taskID]
	return taskInfo, ok
}

func (s *state) GetTaskStatus(taskID string) (*pb.TaskStatusReply, error) {
	req := &pb.ID{Id: taskID}
	return s.minerCtx.miner.TaskDetails(s.ctx, req)
}

func (s *state) SaveTask(dealID DealID, info *TaskInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	taskIDs, ok := s.deals[dealID]
	if !ok {
		return errDealNotFound
	}

	s.tasks[info.ID] = info

	taskIDs.Tasks = append(taskIDs.Tasks, info)
	s.deals[dealID] = taskIDs

	return nil
}

func (s *state) StopTask(ctx context.Context, taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stopTask(ctx, taskID)
}

func (s *state) stopTask(ctx context.Context, taskID string) error {
	task, ok := s.getTaskByID(taskID)
	if !ok {
		return errors.New("no such task")
	}

	_, err := s.minerCtx.miner.Stop(ctx, &pb.ID{Id: task.ID})
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to stop the task %s", task.ID)
	}

	s.deleteTask(task.ID)
	tasksGauge.Dec()

	return nil
}

func (s *state) deleteTask(taskID string) {
	taskInfo, ok := s.tasks[taskID]
	if ok {
		delete(s.tasks, taskID)
	}

	// Commit end time if such task exists in the history, if not - do nothing,
	// something terrible happened, but we just pretend nothing happened.
	taskHistory, ok := s.deals[taskInfo.DealId]
	if ok {
		for _, dealTaskInfo := range taskHistory.Tasks {
			if dealTaskInfo.ID == taskID {
				now := time.Now()
				dealTaskInfo.EndTime = &now
			}
		}
	}
}

// NOTE: `tasksMu` must be held.
func (s *state) restoreResourceUsage() {
	log.G(s.ctx).Debug("synchronizing resource usage")

	for _, dealInfo := range s.deals {

		// It's okay to ignore `AlreadyConsumed` errors here.
		s.minerCtx.Consume(OrderID(dealInfo.Order.GetID()), &dealInfo.Usage)
	}

	for _, orderID := range s.minerCtx.Orders() {
		orderExists := s.orderExists(orderID)
		for _, dealInfo := range s.deals {
			if orderExists {
				break
			}
			if orderID == OrderID(dealInfo.Order.GetID()) {
				orderExists = true
			}
		}

		if !orderExists {
			s.minerCtx.Release(orderID)
		}
	}
}

// releaseDeal closes the specified deal freeing all associated resources.
func (s *state) releaseDeal(dealID DealID) error {
	tasks, err := s.popDealHistory(dealID)
	if err != nil {
		return err
	}

	log.S(s.ctx).Infof("stopping at max %d tasks due to deal closing", len(tasks))
	for _, task := range tasks {
		if s.isTaskFinished(task.ID) {
			continue
		}

		if err := s.stopTask(s.ctx, task.ID); err != nil {
			log.G(s.ctx).Error("failed to stop task",
				zap.Stringer("dealID", dealID),
				zap.String("taskID", task.ID),
				zap.Error(err),
			)
		} else {
			tasksGauge.Dec()
		}
	}

	return nil
}

func (s *state) publishOrder(orderID OrderID) error {
	balance, err := s.eth.Balance()
	if err != nil {
		return err
	}

	if balance.Cmp(orderPublishThresholdETH) <= 0 {
		return fmt.Errorf("insufficient balance (%s <= %s)", balance.String(), orderPublishThresholdETH.String())
	}

	_, err = s.market.CreateOrder(s.ctx, &pb.Order{Id: orderID.String(), OrderType: pb.OrderType_ASK})

	return err
}

func (s *state) hasResources(resources *structs.Resources) bool {
	usage := resource.NewResources(
		int(resources.GetCpuCores()),
		int64(resources.GetMemoryInBytes()),
		resources.GetGPUCount(),
	)
	if err := s.minerCtx.PollConsume(&usage); err == nil {
		return true
	}
	return false
}

// TODO: do we need to signal about error?
func (s *state) announcePlan(ctx context.Context, plan *askPlan) {
	createdOrder, err := s.market.CreateOrder(ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(ctx).Warnf("failed to announce ask plan with id %s on market - %s", plan.ID, zap.Error(err))
		return
	}

	wrappedOrder, err := structs.NewOrder(createdOrder)
	if err != nil {
		log.S(ctx).Warnf("invalid order received from market - %s", plan.ID, zap.Error(err))
		return
	}

	plan.Order = wrappedOrder
}

func (s *state) deannouncePlan(ctx context.Context, plan *askPlan) {
	_, err := s.market.CancelOrder(ctx, plan.Order.Unwrap())
	if err != nil {
		log.S(ctx).Warnf("failed to deannounce order %s (ask plan - %s) on market - %s",
			plan.Order.Id, plan.ID, zap.Error(err))
	} else {
		plan.Order.Id = ""
	}
}
