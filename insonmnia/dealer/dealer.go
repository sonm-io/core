package dealer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

// Dealer interface describes method to open a deal within two orders.
type Dealer interface {
	Deal(ctx context.Context, bid, ask *sonm.Order) (*big.Int, error)
}

type dealer struct {
	key     *ecdsa.PrivateKey
	timeout time.Duration
	bc      blockchain.Blockchainer
	hub     sonm.HubClient
}

// NewDealer returns `Dealer` implementation which can open deal
// for BID and matched ASK orders.
func NewDealer(key *ecdsa.PrivateKey, hub sonm.HubClient, bc blockchain.Blockchainer, timeout time.Duration) Dealer {
	return &dealer{
		key:     key,
		timeout: timeout,
		bc:      bc,
		hub:     hub,
	}
}

func (d *dealer) Deal(ctx context.Context, bid, ask *sonm.Order) (*big.Int, error) {
	log.G(ctx).Info("creating deal on ethereum", zap.String("bidID", bid.GetId()))
	id, err := d.openDeal(ctx, bid)
	if err != nil {
		log.G(ctx).Warn("cannot open deal", zap.Error(err), zap.String("bidID", bid.GetId()))
		return nil, err
	}

	log.G(ctx).Info("approving deal on hub",
		zap.String("dealID", id.String()),
		zap.String("bidID", bid.GetId()),
		zap.String("hubID", ask.GetSupplierID()))

	err = d.approveOnHub(ctx, id, bid, ask)
	if err != nil {
		log.G(ctx).Warn("cannot approve deal on hub, closing deal",
			zap.Error(err), zap.String("bidID", bid.GetId()))

		err = d.closeUnapproved(ctx, id)
		if err != nil {
			log.G(ctx).Warn("cannot close unapproved deal", zap.Error(err))
			return nil, err

		}
		log.G(ctx).Debug("unapproved deal closed")
		return nil, fmt.Errorf("deal id=%s does not approved by hub", id.String())
	}

	return id, nil
}

func (d *dealer) openDeal(ctx context.Context, bid *sonm.Order) (*big.Int, error) {
	dealRequest := &sonm.Deal{
		WorkTime:          bid.GetSlot().GetDuration(),
		BuyerID:           util.PubKeyToAddr(d.key.PublicKey).Hex(),
		SupplierID:        bid.GetSupplierID(),
		Price:             sonm.NewBigInt(structs.CalculateTotalPrice(bid)),
		Status:            sonm.DealStatus_PENDING,
		SpecificationHash: structs.CalculateSpecHash(bid),
	}

	return d.bc.OpenDealPending(ctx, d.key, dealRequest, d.timeout)

}

func (d *dealer) approveOnHub(ctx context.Context, id *big.Int, bid, ask *sonm.Order) error {
	approveRequest := &sonm.ApproveDealRequest{
		DealID: sonm.NewBigInt(id),
		AskID:  ask.GetId(),
		BidID:  bid.GetId(),
	}

	_, err := d.hub.ApproveDeal(ctx, approveRequest)
	return err
}

func (d *dealer) closeUnapproved(ctx context.Context, id *big.Int) error {
	return d.bc.CloseDealPending(ctx, d.key, id, d.timeout)
}
