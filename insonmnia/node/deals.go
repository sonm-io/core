package node

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xconcurrency"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dealsAPI struct {
	remotes *remoteOptions
	log     *zap.SugaredLogger
}

func (d *dealsAPI) List(ctx context.Context, req *sonm.Count) (*sonm.DealsReply, error) {
	addr := sonm.NewEthAddress(crypto.PubkeyToAddress(d.remotes.key.PublicKey))
	filter := &sonm.DealsRequest{
		Status: sonm.DealStatus_DEAL_ACCEPTED,
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

	reply := &sonm.DealsReply{Deal: []*sonm.Deal{}}
	for _, deal := range dealsBySupplier.GetDeals() {
		reply.Deal = append(reply.Deal, deal.Deal)
	}

	for _, deal := range dealsByConsumer.GetDeals() {
		reply.Deal = append(reply.Deal, deal.Deal)
	}

	return reply, nil
}

func (d *dealsAPI) Status(ctx context.Context, id *sonm.BigInt) (*sonm.DealInfoReply, error) {
	deal, err := d.remotes.eth.Market().GetDealInfo(ctx, id.Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not get deal info from blockchain: %s", err)
	}

	reply := &sonm.DealInfoReply{Deal: deal}

	// try to extract extra info for deal if current user is consumer
	if deal.GetConsumerID().Unwrap().Big().Cmp(crypto.PubkeyToAddress(d.remotes.key.PublicKey).Big()) == 0 {
		dealID := deal.GetId().Unwrap().String()
		workerCtx, workerCtxCancel := context.WithTimeout(ctx, 10*time.Second)
		defer workerCtxCancel()

		worker, closer, err := d.remotes.getWorkerClientByEthAddr(workerCtx, deal.GetSupplierID().Unwrap().Hex())
		if err == nil {
			d.log.Debug("try to obtain deal info from the worker")
			defer closer.Close()

			info, err := worker.GetDealInfo(workerCtx, &sonm.ID{Id: dealID})
			if err == nil {
				return info, nil
			}
		}
	}

	return reply, nil
}

func (d *dealsAPI) Finish(ctx context.Context, req *sonm.DealFinishRequest) (*sonm.Empty, error) {
	if err := d.remotes.eth.Market().CloseDeal(ctx, d.remotes.key, req.GetId().Unwrap(), req.GetBlacklistType()); err != nil {
		return nil, fmt.Errorf("could not close deal in blockchain: %s", err)
	}

	return &sonm.Empty{}, nil
}

func (d *dealsAPI) FinishDeals(ctx context.Context, req *sonm.DealsFinishRequest) (*sonm.ErrorByID, error) {
	return d.finishDeals(ctx, req.GetDealInfo())
}

func (d *dealsAPI) finishDeals(ctx context.Context, deals []*sonm.DealFinishRequest) (*sonm.ErrorByID, error) {
	errs := sonm.NewTSErrorByID()
	xconcurrency.Run(purgeConcurrency, deals, func(elem interface{}) {
		info := elem.(*sonm.DealFinishRequest)
		if info.GetId().IsZero() {
			errs.Append(info.GetId(), errors.New("zero deal id specified"))
		} else {
			d.log.Debugw("closing deal", zap.String("id", info.GetId().Unwrap().String()))
			err := d.remotes.eth.Market().CloseDeal(ctx, d.remotes.key, info.GetId().Unwrap(), info.GetBlacklistType())
			errs.Append(info.GetId(), err)
		}
	})

	return errs.Unwrap(), nil
}

func (d *dealsAPI) PurgeDeals(ctx context.Context, req *sonm.DealsPurgeRequest) (*sonm.ErrorByID, error) {
	deals, err := d.remotes.dwh.GetDeals(ctx, &sonm.DealsRequest{
		Status:     sonm.DealStatus_DEAL_ACCEPTED,
		ConsumerID: sonm.NewEthAddress(crypto.PubkeyToAddress(d.remotes.key.PublicKey)),
		Limit:      dwh.MaxLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deals from DWH: %s", err)
	}

	dealInfo := make([]*sonm.DealFinishRequest, 0, len(deals.GetDeals()))
	for _, deal := range deals.GetDeals() {
		dealInfo = append(dealInfo, &sonm.DealFinishRequest{Id: deal.GetDeal().GetId(), BlacklistType: req.GetBlacklistType()})
	}
	return d.finishDeals(ctx, dealInfo)
}

func (d *dealsAPI) Open(ctx context.Context, req *sonm.OpenDealRequest) (*sonm.Deal, error) {
	ask, err := d.remotes.eth.Market().GetOrderInfo(ctx, req.GetAskID().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ask order: %s", err)
	}

	if !req.Force {
		d.remotes.log.Debug("checking worker availability")
		if available := d.remotes.isWorkerAvailable(ctx, ask.GetAuthorID().Unwrap()); !available {
			return nil, status.Errorf(codes.Unavailable,
				"failed to fetch status from %s, seems like worker is offline", ask.GetAuthorID().Unwrap().Hex())
		}
	} else {
		d.remotes.log.Info("forcing deal opening, worker availability checking skipped")
	}

	deal, err := d.remotes.eth.Market().OpenDeal(ctx, d.remotes.key, req.GetAskID().Unwrap(), req.GetBidID().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("could not open deal in blockchain: %s", err)
	}

	return deal, nil
}

func (d *dealsAPI) QuickBuy(ctx context.Context, req *sonm.QuickBuyRequest) (*sonm.DealInfoReply, error) {
	ask, err := d.remotes.eth.Market().GetOrderInfo(ctx, req.GetAskID().Unwrap())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ask order for duration lookup: %s", err)
	}

	var duration uint64
	if req.Duration == nil {
		duration = ask.Duration
	} else {
		duration = uint64(req.GetDuration().Unwrap().Seconds())
	}

	if !req.Force {
		d.remotes.log.Debug("checking worker availability")
		if available := d.remotes.isWorkerAvailable(ctx, ask.GetAuthorID().Unwrap()); !available {
			return nil, status.Errorf(codes.Unavailable,
				"failed to fetch status from %s, seems like worker is offline", ask.GetAuthorID().Unwrap().Hex())
		}
	} else {
		d.remotes.log.Info("forcing deal opening, worker availability checking skipped",
			zap.String("ask_id", req.AskID.Unwrap().String()))
	}

	deal, err := d.remotes.eth.Market().QuickBuy(ctx, d.remotes.key, req.GetAskID().Unwrap(), duration)
	if err != nil {
		return nil, err
	}

	supplierAddr, err := auth.ParseAddr(deal.GetSupplierID().Unwrap().Hex())
	if err != nil {
		d.log.Debugw("cannot create auth.Addr from supplier addr", zap.Error(err))
		return &sonm.DealInfoReply{Deal: deal}, nil
	}

	cli, closer, err := d.remotes.workerCreator(ctx, supplierAddr)
	if err != nil {
		d.log.Debugw("cannot create worker client", zap.Error(err))
		return &sonm.DealInfoReply{Deal: deal}, nil
	}
	defer closer.Close()

	workerDeal, err := cli.GetDealInfo(ctx, &sonm.ID{Id: deal.GetId().Unwrap().String()})
	if err != nil {
		d.log.Debugw("cannot get deal from worker", zap.Error(err))
		return &sonm.DealInfoReply{Deal: deal}, nil
	}

	return workerDeal, nil
}

func (d *dealsAPI) ChangeRequestsList(ctx context.Context, id *sonm.BigInt) (*sonm.DealChangeRequestsReply, error) {
	return d.remotes.dwh.GetDealChangeRequests(ctx, id)
}

func (d *dealsAPI) CreateChangeRequest(ctx context.Context, req *sonm.DealChangeRequest) (*sonm.BigInt, error) {
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
		return nil, fmt.Errorf("cannot create change request: %v", err)
	}

	return sonm.NewBigInt(id), nil
}

func (d *dealsAPI) ApproveChangeRequest(ctx context.Context, id *sonm.BigInt) (*sonm.Empty, error) {
	req, err := d.remotes.eth.Market().GetDealChangeRequestInfo(ctx, id.Unwrap())
	if err != nil {
		return nil, fmt.Errorf("cannot get change request by id: %v", err)
	}

	matchingRequest := &sonm.DealChangeRequest{
		DealID:      req.GetDealID(),
		Duration:    req.GetDuration(),
		RequestType: invertOrderType(req.RequestType),
		Price:       req.GetPrice(),
	}

	_, err = d.remotes.eth.Market().CreateChangeRequest(ctx, d.remotes.key, matchingRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot approve change request: %v", err)
	}

	return &sonm.Empty{}, nil
}

func (d *dealsAPI) CancelChangeRequest(ctx context.Context, id *sonm.BigInt) (*sonm.Empty, error) {
	if err := d.remotes.eth.Market().CancelChangeRequest(ctx, d.remotes.key, id.Unwrap()); err != nil {
		return nil, fmt.Errorf("could not cancel change request: %v", err)
	}

	return &sonm.Empty{}, nil
}

func invertOrderType(s sonm.OrderType) sonm.OrderType {
	if s == sonm.OrderType_ASK {
		return sonm.OrderType_BID
	} else {
		return sonm.OrderType_ASK
	}
}

func newDealsAPI(opts *remoteOptions) sonm.DealManagementServer {
	return &dealsAPI{
		remotes: opts,
		log:     opts.log,
	}
}
