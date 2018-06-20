package node

import (
	"context"

	"github.com/sonm-io/core/proto"
)

type profileAPI struct {
	remotes *remoteOptions
}

func newProfileAPI(opts *remoteOptions) sonm.ProfilesServer {
	return &profileAPI{remotes: opts}
}

func (p *profileAPI) List(ctx context.Context, req *sonm.ProfilesRequest) (*sonm.ProfilesReply, error) {
	return p.remotes.dwh.GetProfiles(ctx, req)
}

func (p *profileAPI) Status(ctx context.Context, addr *sonm.EthID) (*sonm.Profile, error) {
	return p.remotes.dwh.GetProfileInfo(ctx, addr)
}

func (p *profileAPI) RemoveAttribute(ctx context.Context, id *sonm.BigInt) (*sonm.Empty, error) {
	if err := p.remotes.eth.ProfileRegistry().RemoveCertificate(ctx, p.remotes.key, id.Unwrap()); err != nil {
		return nil, err
	}

	return &sonm.Empty{}, nil
}
