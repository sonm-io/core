package commands

import (
	"time"

	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type CliInteractor interface {
	HubPing() (*pb.PingReply, error)
	HubStatus() error

	// MinerList()
	// MinerStatus()

	// TaskList()
	// TaskStart()
	// TaskStatus()
	// TaskStop()
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

func (it *grpcInteractor) HubPing() (*pb.PingReply, error) {
	ctx, cancel := it.ctx(gctx)
	defer cancel()
	return pb.NewHubClient(it.cc).Ping(ctx, &pb.PingRequest{})
}

func (it *grpcInteractor) HubStatus() error {
	return nil
}

func NewGrpcInteractor(addr string, to time.Duration) (CliInteractor, error) {
	i := &grpcInteractor{timeout: to}
	err := i.call(addr)
	if err != nil {
		return nil, err
	}

	return i, nil
}
