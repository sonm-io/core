package commands

import (
	"time"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type CliInteractor interface {
	HubPing(context.Context) (*pb.PingReply, error)
	HubStatus(context.Context) (*pb.HubStatusReply, error)

	MinerList(context.Context) (*pb.ListReply, error)
	MinerStatus(minerID string, appCtx context.Context) (*pb.InfoReply, error)
	MinerGetProperties(ctx context.Context, ID string) (*pb.GetMinerPropertiesReply, error)
	MinerSetProperties(ctx context.Context, ID string, properties map[string]string) (*pb.Empty, error)
	MinerShowSlots(ctx context.Context, ID string) (*pb.GetSlotsReply, error)
	MinerAddSlot(ctx context.Context, ID string, slot *structs.Slot) (*pb.Empty, error)

	TaskList(appCtx context.Context, minerID string) (*pb.StatusMapReply, error)
	TaskLogs(appCtx context.Context, req *pb.TaskLogsRequest) (pb.Hub_TaskLogsClient, error)
	TaskStart(appCtx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error)
	TaskStatus(appCtx context.Context, taskID string) (*pb.TaskStatusReply, error)
	TaskStop(appCtx context.Context, taskID string) (*pb.Empty, error)
}

type grpcInteractor struct {
	hub     pb.HubClient
	timeout time.Duration
}

func (it *grpcInteractor) call(addr string) error {

	return nil
}

func (it *grpcInteractor) ctx(appCtx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(appCtx, it.timeout)
}

func (it *grpcInteractor) HubPing(appCtx context.Context) (*pb.PingReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return it.hub.Ping(ctx, &pb.Empty{})
}

func (it *grpcInteractor) HubStatus(appCtx context.Context) (*pb.HubStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return it.hub.Status(ctx, &pb.Empty{})
}

func (it *grpcInteractor) MinerList(appCtx context.Context) (*pb.ListReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return it.hub.List(ctx, &pb.Empty{})
}

func (it *grpcInteractor) MinerStatus(minerID string, appCtx context.Context) (*pb.InfoReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = pb.ID{Id: minerID}
	return it.hub.Info(ctx, &req)
}

func (it *grpcInteractor) MinerGetProperties(ctx context.Context, ID string) (*pb.GetMinerPropertiesReply, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()

	req := pb.ID{Id: ID}
	return it.hub.GetMinerProperties(c, &req)
}

func (it *grpcInteractor) MinerSetProperties(ctx context.Context, ID string, properties map[string]string) (*pb.Empty, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()

	req := pb.SetMinerPropertiesRequest{
		ID:         ID,
		Properties: properties,
	}
	return it.hub.SetMinerProperties(c, &req)
}

func (it *grpcInteractor) MinerShowSlots(ctx context.Context, ID string) (*pb.GetSlotsReply, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()
	return it.hub.GetSlots(c, &pb.ID{Id: ID})
}

func (it *grpcInteractor) MinerAddSlot(ctx context.Context, ID string, slot *structs.Slot) (*pb.Empty, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()
	return it.hub.AddSlot(c, &pb.AddSlotRequest{ID: ID, Slot: slot.Unwrap()})
}

func (it *grpcInteractor) TaskList(appCtx context.Context, minerID string) (*pb.StatusMapReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	req := &pb.ID{Id: minerID}
	return it.hub.MinerStatus(ctx, req)
}

func (it *grpcInteractor) TaskLogs(appCtx context.Context, req *pb.TaskLogsRequest) (pb.Hub_TaskLogsClient, error) {
	return it.hub.TaskLogs(appCtx, req)
}

func (it *grpcInteractor) TaskStart(appCtx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return it.hub.StartTask(ctx, req)
}

func (it *grpcInteractor) TaskStatus(appCtx context.Context, taskID string) (*pb.TaskStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.ID{Id: taskID}
	return it.hub.TaskStatus(ctx, req)
}

func (it *grpcInteractor) TaskStop(appCtx context.Context, taskID string) (*pb.Empty, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.ID{Id: taskID}
	return it.hub.StopTask(ctx, req)
}

func NewGrpcInteractor(addr string, to time.Duration) (CliInteractor, error) {
	cc, err := util.MakeGrpcClient(addr, nil)
	if err != nil {
		return nil, err
	}

	return &grpcInteractor{
		hub:     pb.NewHubClient(cc),
		timeout: to,
	}, nil
}

type NodeHubInteractor interface {
	Status() (*pb.HubStatusReply, error)

	WorkersList() (*pb.ListReply, error)
	WorkerStatus(id string) (*pb.InfoReply, error)

	GetRegistredWorkers() (*pb.GetRegistredWorkersReply, error)
	RegisterWorker(id string) (*pb.Empty, error)
	UnregisterWorker(id string) (*pb.Empty, error)

	GetWorkerProperties(id string) (*pb.GetMinerPropertiesReply, error)
	SetWorkerProperties(req *pb.SetMinerPropertiesRequest) (*pb.Empty, error)

	GetAskPlans() (*pb.GetAllSlotsReply, error)
	CreateAskPlan(id string, slot *structs.Slot) (*pb.Empty, error)
	RemoveAskPlan(id string) (*pb.Empty, error)

	TaskList() (*pb.TaskListReply, error)
	TaskStatus(id string) (*pb.TaskStatusReply, error)
}

type hubInteractor struct {
	timeout time.Duration
	hub     pb.HubManagementClient
}

func (it *hubInteractor) Status() (*pb.HubStatusReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.Status(ctx, &pb.Empty{})
}

func (it *hubInteractor) WorkersList() (*pb.ListReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.WorkersList(ctx, &pb.Empty{})
}

func (it *hubInteractor) WorkerStatus(id string) (*pb.InfoReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.WorkerStatus(ctx, req)
}

func (it *hubInteractor) GetRegistredWorkers() (*pb.GetRegistredWorkersReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.GetRegistredWorkers(ctx, &pb.Empty{})
}

func (it *hubInteractor) RegisterWorker(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.RegisterWorker(ctx, req)
}

func (it *hubInteractor) UnregisterWorker(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.UnregisterWorker(ctx, req)
}

func (it *hubInteractor) GetWorkerProperties(id string) (*pb.GetMinerPropertiesReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.GetWorkerProperties(ctx, req)
}

func (it *hubInteractor) SetWorkerProperties(req *pb.SetMinerPropertiesRequest) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.SetWorkerProperties(ctx, req)
}

func (it *hubInteractor) GetAskPlans() (*pb.GetAllSlotsReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.GetAskPlans(ctx, &pb.Empty{})
}

func (it *hubInteractor) CreateAskPlan(id string, slot *structs.Slot) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.AddSlotRequest{
		ID:   id,
		Slot: slot.Unwrap(),
	}

	return it.hub.CreateAskPlan(ctx, req)
}

func (it *hubInteractor) RemoveAskPlan(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.RemoveAskPlan(ctx, req)
}

func (it *hubInteractor) TaskList() (*pb.TaskListReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.TaskList(ctx, &pb.Empty{})
}

func (it *hubInteractor) TaskStatus(id string) (*pb.TaskStatusReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.TaskStatus(ctx, req)
}

func NewHubInteractor(addr string, timeout time.Duration) (NodeHubInteractor, error) {
	cc, err := util.MakeGrpcClient(addr, nil)
	if err != nil {
		return nil, err
	}

	return &hubInteractor{
		timeout: timeout,
		hub:     pb.NewHubManagementClient(cc),
	}, nil
}

type NodeMarketInteractor interface {
	GetOrders(slot *structs.Slot, orderType pb.OrderType, count uint64) ([]*pb.Order, error)
	GetOrderByID(id string) (*pb.Order, error)
	CreateOrder(order *pb.Order) (*pb.Order, error)
	CancelOrder(id string) error
}

type marketInteractor struct {
	timeout time.Duration
	market  pb.MarketClient
}

func (it *marketInteractor) GetOrders(slot *structs.Slot, orderType pb.OrderType, count uint64) ([]*pb.Order, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.GetOrdersRequest{
		Slot:      slot.Unwrap(),
		OrderType: orderType,
		Count:     count,
	}

	reply, err := it.market.GetOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply.GetOrders(), nil
}

func (it *marketInteractor) GetOrderByID(id string) (*pb.Order, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.market.GetOrderByID(ctx, &pb.ID{Id: id})
}

func (it *marketInteractor) CreateOrder(order *pb.Order) (*pb.Order, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.market.CreateOrder(ctx, order)
}

func (it *marketInteractor) CancelOrder(id string) error {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	_, err := it.market.CancelOrder(ctx, &pb.Order{Id: id})
	return err
}

func ctx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

func NewMarketInteractor(addr string, timeout time.Duration) (NodeMarketInteractor, error) {
	cc, err := util.MakeGrpcClient(addr, nil)
	if err != nil {
		return nil, err
	}

	market := pb.NewMarketClient(cc)

	return &marketInteractor{
		timeout: timeout,
		market:  market,
	}, nil

}
