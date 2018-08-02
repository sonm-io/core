package node

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type masterMgmtAPI struct {
	remotes *remoteOptions
	log     *zap.SugaredLogger
}

// TODO(sshaman1101): DWH is required to implement this service
func newMasterManagementAPI(opts *remoteOptions) sonm.MasterManagementServer {
	return &masterMgmtAPI{
		remotes: opts,
		log:     opts.log,
	}
}

func (m *masterMgmtAPI) WorkersList(ctx context.Context, address *sonm.EthAddress) (*sonm.WorkerListReply, error) {
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
	err := m.remotes.eth.Market().ConfirmWorker(ctx, m.remotes.key, address.Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not confirm dependant worker in blockchain: %s", err)
	}
	return &sonm.Empty{}, nil
}

func (m *masterMgmtAPI) WorkerRemove(ctx context.Context, request *sonm.WorkerRemoveRequest) (*sonm.Empty, error) {
	err := m.remotes.eth.Market().RemoveWorker(ctx, m.remotes.key, request.GetMaster().Unwrap(), request.GetWorker().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not remove dependant worker from blockchain: %s", err)
	}
	return &sonm.Empty{}, nil
}
