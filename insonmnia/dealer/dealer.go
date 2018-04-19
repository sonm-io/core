package dealer

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Dealer interface describes method to open a deal within two orders.
type Dealer interface {
	Deal(ctx context.Context, bid, ask *sonm.MarketOrder) (*big.Int, error)
}

type dealer struct {
	key     *ecdsa.PrivateKey
	timeout time.Duration
	bc      blockchain.API
	hub     sonm.HubClient
}

// NewDealer returns `Dealer` implementation which can open deal
// for BID and matched ASK orders.
func NewDealer(key *ecdsa.PrivateKey, hub sonm.HubClient, bc blockchain.API, timeout time.Duration) Dealer {
	return &dealer{
		key:     key,
		timeout: timeout,
		bc:      bc,
		hub:     hub,
	}
}

func (d *dealer) Deal(ctx context.Context, bid, ask *sonm.MarketOrder) (*big.Int, error) {
	log.G(ctx).Info("creating deal on ethereum", zap.String("bidID", bid.GetId()))
	id, err := d.openDeal(ctx, bid)
	if err != nil {
		log.G(ctx).Warn("cannot open deal", zap.Error(err), zap.String("bidID", bid.GetId()))
		return nil, err
	}

	log.G(ctx).Info("created deal on eth",
		zap.String("dealID", id.String()),
		zap.String("bidID", bid.GetId()),
		zap.String("hubID", ask.GetAuthor()))

	return id, nil
}

func (d *dealer) openDeal(ctx context.Context, bid *sonm.MarketOrder) (*big.Int, error) {
	//TODO: use MarketDeal
	//dealRequest := &sonm.Deal{
	//	WorkTime:          bid.GetDuration(),
	//	BuyerID:           util.PubKeyToAddr(d.key.PublicKey).Hex(),
	//	SupplierID:        bid.GetCounterparty(),
	//	Price:             sonm.NewBigInt(structs.CalculateTotalPrice(bid)),
	//	Status:            sonm.DealStatus_PENDING,
	//	SpecificationHash: structs.CalculateSpecHash(bid),
	//}
	//
	//return d.bc.OpenDealPending(ctx, d.key, dealRequest, d.timeout)
	return nil, errors.New("unimplemented")
}
