package node

import (
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (m *masterMgmtAPI) WorkersList(context.Context, *sonm.Empty) (*sonm.WorkerListReply, error) {
	ctxlog.G(m.ctx).Info("handling WorkersList request")
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *masterMgmtAPI) WorkerConfirm(context.Context, *sonm.ID) (*sonm.Empty, error) {
	ctxlog.G(m.ctx).Info("handling WorkersConfirm request")
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *masterMgmtAPI) WorkerRemove(context.Context, *sonm.ID) (*sonm.Empty, error) {
	ctxlog.G(m.ctx).Info("handling WorkersRemove request")
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
