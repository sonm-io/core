package commands

import (
	"time"

	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CliInteractor interface {
	HubPing(context.Context) (*pb.PingReply, error)
	HubStatus(context.Context) (*pb.H_StatusReply, error)

	MinerList(context.Context) (*pb.H_MinerListReply, error)
	MinerStatus(minerID string, appCtx context.Context) (*pb.MinerStatusReply, error)

	TaskList(appCtx context.Context, minerID string) (*pb.TaskDetailsMapReply, error)
	TaskLogs(appCtx context.Context, req *pb.TaskLogsRequest) (pb.Hub_TaskLogsClient, error)
	TaskStart(appCtx context.Context, req *pb.H_StartTaskRequest) (*pb.H_StartTaskReply, error)
	TaskStatus(appCtx context.Context, taskID string) (*pb.TaskDetailsReply, error)
	TaskStop(appCtx context.Context, taskID string) (*pb.EmptyReply, error)
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
	return pb.NewHubClient(it.cc).Ping(ctx, &pb.EmptyRequest{})
}

func (it *grpcInteractor) HubStatus(appCtx context.Context) (*pb.H_StatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).Status(ctx, &pb.EmptyRequest{})
}

func (it *grpcInteractor) MinerList(appCtx context.Context) (*pb.H_MinerListReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).MinerList(ctx, &pb.EmptyRequest{})
}

func (it *grpcInteractor) MinerStatus(minerID string, appCtx context.Context) (*pb.MinerStatusReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = pb.H_MinerStatusRequest{Miner: minerID}
	return pb.NewHubClient(it.cc).MinerStatus(ctx, &req)
}

func (it *grpcInteractor) TaskList(appCtx context.Context, minerID string) (*pb.TaskDetailsMapReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	req := &pb.H_TaskMapRequest{Miner: minerID}
	return pb.NewHubClient(it.cc).TaskList(ctx, req)
}

func (it *grpcInteractor) TaskLogs(appCtx context.Context, req *pb.TaskLogsRequest) (pb.Hub_TaskLogsClient, error) {
	return pb.NewHubClient(it.cc).TaskLogs(appCtx, req)
}

func (it *grpcInteractor) TaskStart(appCtx context.Context, req *pb.H_StartTaskRequest) (*pb.H_StartTaskReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()
	return pb.NewHubClient(it.cc).TaskStart(ctx, req)
}

func (it *grpcInteractor) TaskStatus(appCtx context.Context, taskID string) (*pb.TaskDetailsReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.TaskDetailsRequest{Id: taskID}
	return pb.NewHubClient(it.cc).TaskStatus(ctx, req)
}

func (it *grpcInteractor) TaskStop(appCtx context.Context, taskID string) (*pb.EmptyReply, error) {
	ctx, cancel := it.ctx(appCtx)
	defer cancel()

	var req = &pb.TaskStopRequest{Id: taskID}
	return pb.NewHubClient(it.cc).TaskStop(ctx, req)
}

func NewGrpcInteractor(addr string, to time.Duration) (CliInteractor, error) {
	i := &grpcInteractor{timeout: to}
	err := i.call(addr)
	if err != nil {
		return nil, err
	}

	return i, nil
}
