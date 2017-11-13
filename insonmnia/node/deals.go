package node

import (
	"crypto/ecdsa"
	"math/big"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type dealsAPI struct {
	key *ecdsa.PrivateKey
	bc  *blockchain.API
}

func (d *dealsAPI) List(ctx context.Context, req *pb.DealListRequest) (*pb.DealListReply, error) {
	log.G(ctx).Info("handling Deals_List request")
	addr := util.PubKeyToAddr(d.key.PublicKey)
	IDs, err := d.bc.GetDeals(addr)
	if err != nil {
		return nil, err
	}

	// TODO(sshaman1101): add filter by status

	deals := make([]*pb.Deal, 0, len(IDs))
	for _, id := range IDs {
		info, err := d.bc.GetDealInfo(id)
		if err != nil {
			return nil, err
		}

		deals = append(deals, &pb.Deal{
			Price:       info.Price.Int64(),
			Status:      pb.DealStatus(info.Status.Int64()),
			BuyerID:     info.Client,
			SuppclierID: info.Hub,
			StartTime:   &pb.Timestamp{Seconds: info.StartTime.Int64()},
			EndTime:     &pb.Timestamp{Seconds: info.EndTime.Int64()},
		})
	}

	return &pb.DealListReply{Deal: deals}, nil
}

func (d *dealsAPI) Status(ctx context.Context, id *pb.IntID) (*pb.Deal, error) {
	log.G(ctx).Info("handling Deals_Status request", zap.Int64("id", id.Id))
	bigID := big.NewInt(id.Id)

	info, err := d.bc.GetDealInfo(bigID)
	if err != nil {
		return nil, err
	}

	deal := &pb.Deal{
		Price:       info.Price.Int64(),
		Status:      pb.DealStatus(info.Status.Int64()),
		BuyerID:     info.Client,
		SuppclierID: info.Hub,
		StartTime:   &pb.Timestamp{Seconds: info.StartTime.Int64()},
		EndTime:     &pb.Timestamp{Seconds: info.EndTime.Int64()},
	}

	return deal, nil
}

func (d *dealsAPI) Finish(ctx context.Context, id *pb.IntID) (*pb.Empty, error) {
	log.G(ctx).Info("handling Deals_Finish request", zap.Int64("id", id.Id))

	bigID := big.NewInt(id.Id)

	_, err := d.bc.CloseDeal(d.key, bigID)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func newDealsAPI(key *ecdsa.PrivateKey) (pb.DealManagementServer, error) {
	api, err := blockchain.NewBlockchainAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	return &dealsAPI{key: key, bc: api}, nil
}
