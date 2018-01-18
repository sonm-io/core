package node

import (
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"

	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
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

func (d *dealsAPI) Status(ctx context.Context, id *pb.ID) (*pb.DealStatusReply, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	deal, err := d.remotes.eth.GetDealInfo(bigID)
	if err != nil {
		return nil, err
	}

	reply := &pb.DealStatusReply{
		Deal: deal,
	}

	if deal.GetStatus() == pb.DealStatus_ACCEPTED || deal.GetStatus() == pb.DealStatus_PENDING {
		hubClient, closr, err := getHubClientByEthAddr(ctx, d.remotes, deal.GetSupplierID())
		if err == nil {
			defer closr.Close()
			dealInfo, err := hubClient.GetDealInfo(ctx, id)
			if err == nil {
				reply.Info = dealInfo
			} else {
				log.G(ctx).Info("cannot get deal details from hub", zap.Error(err))
			}
		} else {
			log.G(ctx).Info("cannot resolve hub address", zap.Error(err))
		}
	}

	return reply, nil
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
