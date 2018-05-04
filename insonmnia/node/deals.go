package node

import (
	"github.com/ethereum/go-ethereum/crypto"
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

	reply := &pb.DealInfoReply{
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

	if err = <-d.remotes.eth.Market().CloseDeal(ctx, d.remotes.key, bigID, false); err != nil {
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
