package node

import (
	"crypto/ecdsa"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type dealsAPI struct {
	key *ecdsa.PrivateKey
	bc  blockchain.Blockchainer
	ctx context.Context
}

func (d *dealsAPI) List(ctx context.Context, req *pb.DealListRequest) (*pb.DealListReply, error) {
	log.G(d.ctx).Info("handling Deals_List request", zap.Any("req", req))
	IDs, err := d.bc.GetDeals(req.Owner)
	if err != nil {
		return nil, err
	}

	deals := make([]*pb.Deal, 0, len(IDs))
	for _, id := range IDs {
		deal, err := d.bc.GetDealInfo(id)
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
	log.G(d.ctx).Info("handling Deals_Status request", zap.String("id", id.Id))
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	return d.bc.GetDealInfo(bigID)
}

func (d *dealsAPI) Finish(ctx context.Context, id *pb.ID) (*pb.Empty, error) {
	log.G(d.ctx).Info("handling Deals_Finish request", zap.String("id", id.Id))
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	_, err = d.bc.CloseDeal(d.key, bigID)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func newDealsAPI(opts *remoteOptions) (pb.DealManagementServer, error) {
	api, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	return &dealsAPI{key: opts.key, bc: api, ctx: opts.ctx}, nil
}
