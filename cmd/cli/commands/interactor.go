package commands

import (
	"time"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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
	cc      *grpc.ClientConn
	timeout time.Duration
	hub     pb.HubClient
}

func (it *grpcInteractor) call(addr string) error {
	cc, err := grpc.Dial(addr, grpc.WithInsecure(),
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	if err != nil {
		return err
	}

	it.cc = cc
	return nil
}

func (it *grpcInteractor) ctx(appCtx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(appCtx, it.timeout)
}

func (it *grpcInteractor) HubPing(appCtx context.Context) (*pb.PingReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).Ping(ctx, &pb.Empty{})
}

func (it *grpcInteractor) HubStatus(appCtx context.Context) (*pb.HubStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).Status(ctx, &pb.Empty{})
}

func (it *grpcInteractor) MinerList(appCtx context.Context) (*pb.ListReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).List(ctx, &pb.Empty{})
}

func (it *grpcInteractor) MinerStatus(minerID string, appCtx context.Context) (*pb.InfoReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = pb.HubInfoRequest{Miner: minerID}
	return pb.NewHubClient(it.cc).Info(ctx, &req)
}

func (it *grpcInteractor) MinerGetProperties(ctx context.Context, ID string) (*pb.GetMinerPropertiesReply, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()

	req := pb.ID{Id: ID}
	return pb.NewHubClient(it.cc).GetMinerProperties(c, &req)
}

func (it *grpcInteractor) MinerSetProperties(ctx context.Context, ID string, properties map[string]string) (*pb.Empty, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()

	req := pb.SetMinerPropertiesRequest{
		ID:         ID,
		Properties: properties,
	}
	return pb.NewHubClient(it.cc).SetMinerProperties(c, &req)
}

func (it *grpcInteractor) MinerShowSlots(ctx context.Context, ID string) (*pb.GetSlotsReply, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()
	return pb.NewHubClient(it.cc).GetSlots(c, &pb.ID{Id: ID})
}

func (it *grpcInteractor) MinerAddSlot(ctx context.Context, ID string, slot *structs.Slot) (*pb.Empty, error) {
	c, cancel := it.ctx(ctx)
	defer cancel()
	return pb.NewHubClient(it.cc).AddSlot(c, &pb.AddSlotRequest{ID: ID, Slot: slot.Unwrap()})
}

func (it *grpcInteractor) TaskList(appCtx context.Context, minerID string) (*pb.StatusMapReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	req := &pb.ID{Id: minerID}
	return pb.NewHubClient(it.cc).MinerStatus(ctx, req)
}

func (it *grpcInteractor) TaskLogs(appCtx context.Context, req *pb.TaskLogsRequest) (pb.Hub_TaskLogsClient, error) {
	return pb.NewHubClient(it.cc).TaskLogs(appCtx, req)
}

func (it *grpcInteractor) TaskStart(appCtx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).StartTask(ctx, req)
}

func (it *grpcInteractor) TaskStatus(appCtx context.Context, taskID string) (*pb.TaskStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.ID{Id: taskID}
	return pb.NewHubClient(it.cc).TaskStatus(ctx, req)
}

func (it *grpcInteractor) TaskStop(appCtx context.Context, taskID string) (*pb.Empty, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.ID{Id: taskID}
	return pb.NewHubClient(it.cc).StopTask(ctx, req)
}

func NewGrpcInteractor(addr string, to time.Duration) (CliInteractor, error) {
	i := &grpcInteractor{timeout: to}
	err := i.call(addr)
	if err != nil {
		return nil, err
	}

	return i, nil
}
