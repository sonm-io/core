package node

import (
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type dealsAPI struct {
	ctx     context.Context
	remotes *remoteOptions
}

func (d *dealsAPI) List(ctx context.Context, req *pb.DealListRequest) (*pb.DealListReply, error) {
	IDs, err := d.remotes.eth.GetDeals(req.Owner)
	if err != nil {
		return nil, err
	}

	deals := make([]*pb.Deal, 0, len(IDs))
	for _, id := range IDs {
		deal, err := d.remotes.eth.GetDealInfo(id)
		if err != nil {
			return nil, err
		}

		// filter by status
		if req.Status != pb.DealStatus_ANY_STATUS && req.Status != pb.DealStatus(deal.Status) {
			continue
		}

		deals = append(deals, deal)
	}

	return &pb.DealListReply{Deal: deals}, nil
}

func (d *dealsAPI) Status(ctx context.Context, id *pb.ID) (*pb.Deal, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	return d.remotes.eth.GetDealInfo(bigID)
}

func (d *dealsAPI) Finish(ctx context.Context, id *pb.ID) (*pb.Empty, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	_, err = d.remotes.eth.CloseDeal(d.remotes.key, bigID)
	if err != nil {
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
