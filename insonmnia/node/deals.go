package node

import (
	"github.com/sonm-io/core/insonmnia/dwh"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type dealsAPI struct {
	ctx     context.Context
	remotes *remoteOptions
}

func (d *dealsAPI) List(ctx context.Context, req *pb.DealListRequest) (*pb.DealListReply, error) {
	// TODO(sshaman1101): better filters, need to discuss first
	filters := dwh.DealsFilter{
		Author: req.Owner.Unwrap(),
		Status: pb.MarketDealStatus(req.Status),
	}

	deals, err := d.remotes.dwh.GetDeals(ctx, filters)
	if err != nil {
		return nil, err
	}

	return &pb.DealListReply{Deal: deals}, nil
}

func (d *dealsAPI) Status(ctx context.Context, id *pb.ID) (*pb.DealStatusReply, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	deal, err := d.remotes.eth.GetDealInfo(ctx, bigID)
	if err != nil {
		return nil, err
	}

	reply := &pb.DealStatusReply{
		// TODO(sshaman1101): need to find a way to extract deal details from related Worker.
		Deal: deal,
	}

	return reply, nil
}

func (d *dealsAPI) Finish(ctx context.Context, id *pb.ID) (*pb.Empty, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	if _, err = d.remotes.eth.CloseDeal(ctx, d.remotes.key, bigID, false); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func newDealsAPI(opts *remoteOptions) (pb.DealManagementServer, error) {
	return &dealsAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}, nil
}
