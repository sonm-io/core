package commands

import (
	"time"

	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CliInteractor interface {
	HubPing(context.Context) (*pb.PingReply, error)
	HubStatus() error

	MinerList(context.Context) (*pb.ListReply, error)
	MinerStatus(minerID string, appCtx context.Context) (*pb.InfoReply, error)

	TaskList(appCtx context.Context, minerID string) (*pb.StatusMapReply, error)
	TaskStart(appCtx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error)
	TaskStatus(appCtx context.Context, taskID string) (*pb.TaskStatusReply, error)
	TaskStop(appCtx context.Context, taskID string) (*pb.StopTaskReply, error)
}

type grpcInteractor struct {
	cc      *grpc.ClientConn
	timeout time.Duration
	hub     pb.HubClient
}

func (it *grpcInteractor) call(addr string) error {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
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
	return pb.NewHubClient(it.cc).Ping(ctx, &pb.PingRequest{})
}

func (it *grpcInteractor) HubStatus() error {
	return nil
}

func (it *grpcInteractor) MinerList(appCtx context.Context) (*pb.ListReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).List(ctx, &pb.ListRequest{})
}

func (it *grpcInteractor) MinerStatus(minerID string, appCtx context.Context) (*pb.InfoReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = pb.HubInfoRequest{Miner: minerID}
	return pb.NewHubClient(it.cc).Info(ctx, &req)
}

func (it *grpcInteractor) TaskList(appCtx context.Context, minerID string) (*pb.StatusMapReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	req := &pb.HubStatusMapRequest{Miner: minerID}
	return pb.NewHubClient(it.cc).MinerStatus(ctx, req)
}

func (it *grpcInteractor) TaskStart(appCtx context.Context, req *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).StartTask(ctx, req)
}

func (it *grpcInteractor) TaskStatus(appCtx context.Context, taskID string) (*pb.TaskStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.TaskStatusRequest{Id: taskID}
	return pb.NewHubClient(it.cc).TaskStatus(ctx, req)
}

func (it *grpcInteractor) TaskStop(appCtx context.Context, taskID string) (*pb.StopTaskReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.StopTaskRequest{Id: taskID}
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
