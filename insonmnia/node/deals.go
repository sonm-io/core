package node

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type dealsAPI struct {
	ctx     context.Context
	remotes *remoteOptions
}

func (d *dealsAPI) List(ctx context.Context, req *pb.Count) (*pb.DealsReply, error) {
	addr := pb.NewEthAddress(crypto.PubkeyToAddress(d.remotes.key.PublicKey))
	filter := &pb.DealsRequest{
		Status: pb.DealStatus_DEAL_ACCEPTED,
		Limit:  req.GetCount(),
	}

	filter.SupplierID = addr
	dealsBySupplier, err := d.remotes.dwh.GetDeals(ctx, filter)
	if err != nil {
		return nil, err
	}

	filter.SupplierID = nil
	filter.ConsumerID = addr
	dealsByConsumer, err := d.remotes.dwh.GetDeals(ctx, filter)
	if err != nil {
		return nil, err
	}

	reply := &pb.DealsReply{Deal: []*pb.Deal{}}
	for _, deal := range dealsBySupplier.GetDeals() {
		reply.Deal = append(reply.Deal, deal.Deal)
	}

	for _, deal := range dealsByConsumer.GetDeals() {
		reply.Deal = append(reply.Deal, deal.Deal)
	}

	return reply, nil
}

func (d *dealsAPI) Status(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	deal, err := d.remotes.eth.Market().GetDealInfo(ctx, bigID)
	if err != nil {
		return nil, err
	}

	reply := &pb.DealInfoReply{Deal: deal}

	// try to extract extra info for deal
	dealID := deal.GetId().Unwrap().String()
	hub, closer, err := d.remotes.getHubClientByEthAddr(ctx, deal.GetSupplierID().Unwrap().Hex())
	if err == nil {
		ctxlog.G(d.remotes.ctx).Debug("try to obtain deal info from the worker")
		defer closer.Close()
		info, err := hub.GetDealInfo(ctx, &pb.ID{Id: dealID})
		if err == nil {
			return info, nil
		}

	}

	return reply, nil
}

func (d *dealsAPI) Finish(ctx context.Context, id *pb.ID) (*pb.Empty, error) {
	bigID, err := util.ParseBigInt(id.Id)
	if err != nil {
		return nil, err
	}

	if err = <-d.remotes.eth.Market().CloseDeal(ctx, d.remotes.key, bigID, false); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (d *dealsAPI) Open(ctx context.Context, req *pb.OpenDealRequest) (*pb.Deal, error) {
	dealOrErr := <-d.remotes.eth.Market().OpenDeal(ctx, d.remotes.key, req.GetAskID().Unwrap(), req.GetBidID().Unwrap())
	if dealOrErr.Err != nil {
		return nil, dealOrErr.Err
	}

	return dealOrErr.Deal, nil
}

func newDealsAPI(opts *remoteOptions) (pb.DealManagementServer, error) {
	return &dealsAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}, nil
}
