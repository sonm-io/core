package commands

import (
	"context"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
)

// newClientConn provides a single point for gPRC's ClientConn configuration.
//
// Note that `timeoutFlag`, `nodeAddressFlag` and `creds` are set implicitly because it is global for all CLI-related stuff.
func newClientConn(ctx context.Context) (*grpc.ClientConn, error) {
	return xgrpc.NewClient(ctx, nodeAddress(), creds)
}

func newWorkerManagementClient(ctx context.Context) (pb.WorkerManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewWorkerManagementClient(cc), nil
}

func newMasterManagementClient(ctx context.Context) (pb.MasterManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewMasterManagementClient(cc), nil
}

func newMarketClient(ctx context.Context) (pb.MarketClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewMarketClient(cc), nil
}

func newDealsClient(ctx context.Context) (pb.DealManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewDealManagementClient(cc), nil
}

func newTaskClient(ctx context.Context) (pb.TaskManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewTaskManagementClient(cc), nil
}

func newTokenManagementClient(ctx context.Context) (pb.TokenManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewTokenManagementClient(cc), nil
}

func newBlacklistClient(ctx context.Context) (pb.BlacklistClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewBlacklistClient(cc), nil
}

func newProfilesClient(ctx context.Context) (pb.ProfilesClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewProfilesClient(cc), nil
}
