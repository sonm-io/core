package commands

import (
	"time"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type NodeHubInteractor interface {
	Status() (*pb.HubStatusReply, error)

	WorkersList() (*pb.ListReply, error)
	WorkerStatus(id string) (*pb.InfoReply, error)

	GetRegisteredWorkers() (*pb.GetRegisteredWorkersReply, error)
	RegisterWorker(id string) (*pb.Empty, error)
	DeregisterWorker(id string) (*pb.Empty, error)

	DevicesList() (*pb.DevicesReply, error)
	GetDeviceProperties(id string) (*pb.GetDevicePropertiesReply, error)
	SetDeviceProperties(ID string, properties map[string]float64) (*pb.Empty, error)

	GetAskPlans() (*pb.SlotsReply, error)
	CreateAskPlan(req *pb.InsertSlotRequest) (*pb.ID, error)
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

func (it *hubInteractor) GetRegisteredWorkers() (*pb.GetRegisteredWorkersReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.GetRegisteredWorkers(ctx, &pb.Empty{})
}

func (it *hubInteractor) RegisterWorker(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.RegisterWorker(ctx, req)
}

func (it *hubInteractor) DeregisterWorker(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.DeregisterWorker(ctx, req)
}

func (it *hubInteractor) DevicesList() (*pb.DevicesReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.DeviceList(ctx, &pb.Empty{})
}

func (it *hubInteractor) GetDeviceProperties(id string) (*pb.GetDevicePropertiesReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.ID{Id: id}
	return it.hub.GetDeviceProperties(ctx, req)
}

func (it *hubInteractor) SetDeviceProperties(ID string, properties map[string]float64) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.SetDevicePropertiesRequest{
		ID:         ID,
		Properties: properties,
	}

	return it.hub.SetDeviceProperties(ctx, req)
}

func (it *hubInteractor) GetAskPlans() (*pb.SlotsReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.GetAskPlans(ctx, &pb.Empty{})
}

func (it *hubInteractor) CreateAskPlan(req *pb.InsertSlotRequest) (*pb.ID, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.CreateAskPlan(ctx, req)
}

func (it *hubInteractor) RemoveAskPlan(id string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.hub.RemoveAskPlan(ctx, &pb.ID{Id: id})
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
	cc, err := util.MakeWalletAuthenticatedClient(context.Background(), creds, addr)
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
	GetProcessing() (*pb.GetProcessingReply, error)
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
		Order: &pb.Order{
			Slot:      slot.Unwrap(),
			OrderType: orderType,
		},
		Count: count,
	}

	reply, err := it.market.GetOrders(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply.GetOrders(), nil
}

func (it *marketInteractor) GetProcessing() (*pb.GetProcessingReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.market.GetProcessing(ctx, &pb.Empty{})
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
	cc, err := util.MakeWalletAuthenticatedClient(context.Background(), creds, addr)
	if err != nil {
		return nil, err
	}

	market := pb.NewMarketClient(cc)

	return &marketInteractor{
		timeout: timeout,
		market:  market,
	}, nil

}

type DealsInteractor interface {
	List(from string, status pb.DealStatus) ([]*pb.Deal, error)
	Status(id string) (*pb.Deal, error)
	FinishDeal(id string) error
}

type dealsInteractor struct {
	timeout time.Duration
	deals   pb.DealManagementClient
}

func (it *dealsInteractor) List(from string, status pb.DealStatus) ([]*pb.Deal, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.DealListRequest{Owner: from, Status: status}
	reply, err := it.deals.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return reply.GetDeal(), nil
}

func (it *dealsInteractor) Status(id string) (*pb.Deal, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.deals.Status(ctx, &pb.ID{Id: id})
}

func (it *dealsInteractor) FinishDeal(id string) error {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	_, err := it.deals.Finish(ctx, &pb.ID{Id: id})
	return err
}

func NewDealsInteractor(addr string, timeout time.Duration) (DealsInteractor, error) {
	cc, err := util.MakeWalletAuthenticatedClient(context.Background(), creds, addr)
	if err != nil {
		return nil, err
	}

	return &dealsInteractor{
		timeout: timeout,
		deals:   pb.NewDealManagementClient(cc),
	}, nil
}

type TasksInteractor interface {
	List(hubAddr string) (*pb.TaskListReply, error)
	ImagePush(ctx context.Context) (pb.Hub_PushTaskClient, error)
	Start(req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error)
	Status(id, hub string) (*pb.TaskStatusReply, error)
	Logs(req *pb.TaskLogsRequest) (pb.TaskManagement_LogsClient, error)
	Stop(id, hub string) (*pb.Empty, error)
	ImagePull(dealID, taskID string) (pb.Hub_PullTaskClient, error)
}

type tasksInteractor struct {
	timeout time.Duration
	tasks   pb.TaskManagementClient
}

func (it *tasksInteractor) List(hubAddr string) (*pb.TaskListReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	req := &pb.TaskListRequest{HubID: hubAddr}
	return it.tasks.List(ctx, req)
}

func (it *tasksInteractor) ImagePush(ctx context.Context) (pb.Hub_PushTaskClient, error) {
	return it.tasks.PushTask(ctx)
}

func (it *tasksInteractor) Start(req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.tasks.Start(ctx, req)
}

func (it *tasksInteractor) Status(id, hub string) (*pb.TaskStatusReply, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.tasks.Status(ctx, &pb.TaskID{Id: id, HubAddr: hub})
}

func (it *tasksInteractor) Logs(req *pb.TaskLogsRequest) (pb.TaskManagement_LogsClient, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.tasks.Logs(ctx, req)
}

func (it *tasksInteractor) Stop(id, hub string) (*pb.Empty, error) {
	ctx, cancel := ctx(it.timeout)
	defer cancel()

	return it.tasks.Stop(ctx, &pb.TaskID{Id: id, HubAddr: hub})
}

func (it *tasksInteractor) ImagePull(dealID, taskID string) (pb.Hub_PullTaskClient, error) {
	ctx := context.Background()

	req := &pb.PullTaskRequest{
		DealId: dealID,
		TaskId: taskID,
	}

	return it.tasks.PullTask(ctx, req)
}

func NewTasksInteractor(addr string, timeout time.Duration, opts ...grpc.DialOption) (TasksInteractor, error) {
	cc, err := util.MakeWalletAuthenticatedClient(context.Background(), creds, addr, opts...)
	if err != nil {
		return nil, err
	}

	return &tasksInteractor{
		timeout: timeout,
		tasks:   pb.NewTaskManagementClient(cc),
	}, nil
}
