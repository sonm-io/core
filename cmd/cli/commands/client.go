package commands

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc"
)

// newClientConn provides a single point for gPRC's ClientConn configuration.
//
// Note that `timeoutFlag`, `nodeAddressFlag` and `creds` are set implicitly because it is global for all CLI-related stuff.
func newClientConn(ctx context.Context) (*grpc.ClientConn, error) {
	var addr string

	fullEndpoint, err := auth.ParseAddr(nodeAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to parse node address: %v", err)
	}

	if insecureFlag {
		addr, err = fullEndpoint.Addr()
		if err != nil {
			return nil, fmt.Errorf("failed to parse node addr: %v", err)
		}
	} else {
		addr = fullEndpoint.String()
	}

	return xgrpc.NewClient(ctx, addr, creds)
}

func newWorkerManagementClient(ctx context.Context) (sonm.WorkerManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewWorkerManagementClient(cc), nil
}

func newMasterManagementClient(ctx context.Context) (sonm.MasterManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewMasterManagementClient(cc), nil
}

func newMarketClient(ctx context.Context) (sonm.MarketClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewMarketClient(cc), nil
}

func newDealsClient(ctx context.Context) (sonm.DealManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewDealManagementClient(cc), nil
}

func newTaskClient(ctx context.Context) (sonm.WorkerClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewWorkerClient(cc), nil
}

func newTokenManagementClient(ctx context.Context) (sonm.TokenManagementClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewTokenManagementClient(cc), nil
}

func newBlacklistClient(ctx context.Context) (sonm.BlacklistClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewBlacklistClient(cc), nil
}

func newProfilesClient(ctx context.Context) (sonm.ProfilesClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewProfilesClient(cc), nil
}

func newDWHClient(ctx context.Context) (sonm.DWHClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewDWHClient(cc), nil
}

func newDevicesClient(ctx context.Context) (sonm.DevicesStorageClient, error) {
	cc, err := newClientConn(ctx)
	if err != nil {
		return nil, err
	}

	return sonm.NewDevicesStorageClient(cc), nil
}
