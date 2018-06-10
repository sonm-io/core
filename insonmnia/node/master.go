package node

import (
	"fmt"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
)

type masterMgmtAPI struct {
	ctx     context.Context
	remotes *remoteOptions
}

// TODO(sshaman1101): DWH is required to implement this service
func newMasterManagementAPI(opts *remoteOptions) sonm.MasterManagementServer {
	return &masterMgmtAPI{
		ctx:     opts.ctx,
		remotes: opts,
	}
}

func (m *masterMgmtAPI) WorkersList(ctx context.Context, address *sonm.EthAddress) (*sonm.WorkerListReply, error) {
	ctxlog.G(m.ctx).Info("handling WorkersList request")
	// TODO: pagination
	reply, err := m.remotes.dwh.GetWorkers(ctx, &sonm.WorkersRequest{MasterID: address})
	if err != nil {
		return nil, fmt.Errorf("could not get dependant worker list from DWH: %s", err)
	}
	return &sonm.WorkerListReply{
		Workers: reply.Workers,
	}, nil
}

func (m *masterMgmtAPI) WorkerConfirm(ctx context.Context, address *sonm.EthAddress) (*sonm.Empty, error) {
	ctxlog.G(m.ctx).Info("handling WorkersConfirm request")
	err := m.remotes.eth.Market().ConfirmWorker(ctx, m.remotes.key, address.Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not confirm dependant worker in blockchain: %s", err)
	}
	return &sonm.Empty{}, nil
}

func (m *masterMgmtAPI) WorkerRemove(ctx context.Context, request *sonm.WorkerRemoveRequest) (*sonm.Empty, error) {
	ctxlog.G(m.ctx).Info("handling WorkersRemove request")
	err := m.remotes.eth.Market().RemoveWorker(ctx, m.remotes.key, request.GetMaster().Unwrap(), request.GetWorker().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not remove dependant worker from blockchain: %s", err)
	}
	return &sonm.Empty{}, nil
}
