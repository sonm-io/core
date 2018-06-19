package node

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
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

	filter.ConsumerID = addr
	dealsBySupplier, err := d.remotes.dwh.GetDeals(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not get deals from DWH: %s", err)
	}

	filter.ConsumerID = nil
	filter.MasterID = addr
	dealsByConsumer, err := d.remotes.dwh.GetDeals(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("could not get deals from DWH: %s", err)
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

func (d *dealsAPI) Status(ctx context.Context, id *pb.BigInt) (*pb.DealInfoReply, error) {
	deal, err := d.remotes.eth.Market().GetDealInfo(ctx, id.Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not get deal info from blockchain: %s", err)
	}

	reply := &pb.DealInfoReply{Deal: deal}

	// try to extract extra info for deal
	dealID := deal.GetId().Unwrap().String()
	workerCtx, workerCtxCancel := context.WithTimeout(ctx, 10*time.Second)
	defer workerCtxCancel()

	worker, closer, err := d.remotes.getWorkerClientByEthAddr(workerCtx, deal.GetSupplierID().Unwrap().Hex())
	if err == nil {
		ctxlog.G(d.remotes.ctx).Debug("try to obtain deal info from the worker")
		defer closer.Close()

		info, err := worker.GetDealInfo(workerCtx, &pb.ID{Id: dealID})
		if err == nil {
			return info, nil
		}
	}

	return reply, nil
}

func (d *dealsAPI) Finish(ctx context.Context, req *pb.DealFinishRequest) (*pb.Empty, error) {
	if err := d.remotes.eth.Market().CloseDeal(ctx, d.remotes.key, req.GetId().Unwrap(), req.GetBlacklistType()); err != nil {
		return nil, fmt.Errorf("could not close deal in blockchain: %s", err)
	}

	return &pb.Empty{}, nil
}

func (d *dealsAPI) Open(ctx context.Context, req *pb.OpenDealRequest) (*pb.Deal, error) {
	deal, err := d.remotes.eth.Market().OpenDeal(ctx, d.remotes.key, req.GetAskID().Unwrap(), req.GetBidID().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not open deal in blockchain: %s", err)
	}

	return deal, nil
}

func (d *dealsAPI) QuickBuy(ctx context.Context, req *pb.QuickBuyRequest) (*pb.Deal, error) {
	var duration uint64
	if req.Duration == nil {
		ask, err := d.remotes.eth.Market().GetOrderInfo(ctx, req.GetAskId().Unwrap())
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ask order for duration lookup: %s", err)
		}
		duration = ask.Duration
	} else {
		duration = uint64(req.GetDuration().Unwrap().Seconds())
	}
	return d.remotes.eth.Market().QuickBuy(ctx, d.remotes.key, req.GetAskId().Unwrap(), duration)
}

func (d *dealsAPI) ChangeRequestsList(ctx context.Context, id *pb.BigInt) (*pb.DealChangeRequestsReply, error) {
	return d.remotes.dwh.GetDealChangeRequests(ctx, id)
}

func (d *dealsAPI) CreateChangeRequest(ctx context.Context, req *pb.DealChangeRequest) (*pb.BigInt, error) {
	deal, err := d.remotes.eth.Market().GetDealInfo(ctx, req.GetDealID().Unwrap())
	if err != nil {
		return nil, err
	}

	myAddr := crypto.PubkeyToAddress(d.remotes.key.PublicKey)
	iamConsumer := deal.GetConsumerID().Unwrap().Big().Cmp(myAddr.Big()) == 0
	iamMaster := deal.GetMasterID().Unwrap().Big().Cmp(myAddr.Big()) == 0

	if !(iamConsumer || iamMaster) {
		return nil, errors.New("deal is not related to current user")
	}

	if req.Duration == 0 {
		req.Duration = deal.GetDuration()
	}

	if req.Price == nil {
		req.Price = deal.GetPrice()
	}

	id, err := d.remotes.eth.Market().CreateChangeRequest(ctx, d.remotes.key, req)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot create change request")
	}

	return pb.NewBigInt(id), nil
}

func (d *dealsAPI) ApproveChangeRequest(ctx context.Context, id *pb.BigInt) (*pb.Empty, error) {
	req, err := d.remotes.eth.Market().GetDealChangeRequestInfo(ctx, id.Unwrap())
	if err != nil {
		return nil, errors.WithMessage(err, "cannot get change request by id")
	}

	matchingRequest := &pb.DealChangeRequest{
		DealID:      req.GetDealID(),
		Duration:    req.GetDuration(),
		RequestType: invertOrderType(req.RequestType),
		Price:       req.GetPrice(),
	}

	_, err = d.remotes.eth.Market().CreateChangeRequest(ctx, d.remotes.key, matchingRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot approve change request")
	}

	return &pb.Empty{}, nil
}

func (d *dealsAPI) CancelChangeRequest(ctx context.Context, id *pb.BigInt) (*pb.Empty, error) {
	if err := d.remotes.eth.Market().CancelChangeRequest(ctx, d.remotes.key, id.Unwrap()); err != nil {
		return nil, fmt.Errorf("could not cancel change request: %v", err)
	}

	return &pb.Empty{}, nil
}

func invertOrderType(s pb.OrderType) pb.OrderType {
	if s == pb.OrderType_ASK {
		return pb.OrderType_BID
	} else {
		return pb.OrderType_ASK
	}
}

func newDealsAPI(opts *remoteOptions) (pb.DealManagementServer, error) {
	return &dealsAPI{
		remotes: opts,
		ctx:     opts.ctx,
	}, nil
}
