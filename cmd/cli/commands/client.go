package commands

import (
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// newUnixSocketClientConn provides a single point for gPRC's ClientConn configuration.
//
// Note that `timeoutFlag` and `nodeAddressFlag` are set implicitly because it is global for all CLI-related stuff.
func newUnixSocketClientConn(ctx context.Context) (*grpc.ClientConn, error) {
	return xgrpc.NewUnencryptedUnixSocketClient(ctx, nodeAddressFlag, timeoutFlag)
}

func newHubManagementClient(ctx context.Context) (pb.HubManagementClient, error) {
	cc, err := newUnixSocketClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewHubManagementClient(cc), nil
}

func newMarketClient(ctx context.Context) (pb.MarketClient, error) {
	cc, err := newUnixSocketClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewMarketClient(cc), nil
}

func newDealsClient(ctx context.Context) (pb.DealManagementClient, error) {
	cc, err := newUnixSocketClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewDealManagementClient(cc), nil
}

func newTaskClient(ctx context.Context) (pb.TaskManagementClient, error) {
	cc, err := newUnixSocketClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return pb.NewTaskManagementClient(cc), nil
}
