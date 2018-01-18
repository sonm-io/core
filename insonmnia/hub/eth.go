package hub

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type ETH interface {
	// VerifyBuyerBalance verifies that the buyer specified under the given
	// order has enough balance to have a deal.
	VerifyBuyerBalance(bidOrder *structs.Order) error
	// VerifyBuyerAllowance verifies that the buyer specified under the given
	// order has enough allowance to have a deal.
	VerifyBuyerAllowance(bidOrder *structs.Order) error
	// GetAcceptedDeals returns all currently accepted deals.
	GetAcceptedDeals(ctx context.Context) ([]*pb.Deal, error)
	// GetClosedDeals returns all currently closed deals.
	// Warning: use with caution: this may return significantly large amount
	// of data.
	GetClosedDeals(ctx context.Context) ([]*pb.Deal, error)
	// WaitForDealCreated waits for deal created on Buyer-side
	WaitForDealCreated(dealID DealID, buyerID common.Address) (*pb.Deal, error)
	// WaitForDealClosed blocks the current execution context until the
	// specified deal is closed.
	WaitForDealClosed(ctx context.Context, dealID DealID, buyerID string) error
	// AcceptDeal approves deal on Hub-side.
	AcceptDeal(id string) error
	// CloseDeal closes the specified deal on Hub-side.
	CloseDeal(id DealID) error
	// GetDeal checks whether a given deal exists.
	GetDeal(id string) (*pb.Deal, error)
}

const defaultDealWaitTimeout = 900 * time.Second

type eth struct {
	key     *ecdsa.PrivateKey
	bc      blockchain.Blockchainer
	ctx     context.Context
	timeout time.Duration
}

func (e *eth) hubAddress() string {
	return crypto.PubkeyToAddress(e.key.PublicKey).Hex()
}

func (e *eth) VerifyBuyerBalance(bidOrder *structs.Order) error {
	balance, err := e.bc.BalanceOf(bidOrder.ByuerID)
	if err != nil {
		return err
	}
	if balance.Cmp(bidOrder.GetTotalPrice()) == -1 {
		return fmt.Errorf("not enough balance to have a deal: %v when required at least %v", balance, bidOrder.GetTotalPrice())
	}

	return nil
}

func (e *eth) VerifyBuyerAllowance(bidOrder *structs.Order) error {
	allowance, err := e.bc.AllowanceOf(bidOrder.ByuerID, tsc.DealsAddress)
	if err != nil {
		return err
	}
	if allowance.Cmp(bidOrder.GetTotalPrice()) == -1 {
		return fmt.Errorf("not enough allowance to have a deal: %v when required at least %v", allowance, bidOrder.GetTotalPrice())
	}

	return nil
}

func (e *eth) GetAcceptedDeals(ctx context.Context) ([]*pb.Deal, error) {
	return e.getTemplateDeals(ctx, e.bc.GetAcceptedDeal)
}

func (e *eth) GetClosedDeals(ctx context.Context) ([]*pb.Deal, error) {
	return e.getTemplateDeals(ctx, e.bc.GetClosedDeal)
}

func (e *eth) getTemplateDeals(ctx context.Context, fn func(string, string) ([]*big.Int, error)) ([]*pb.Deal, error) {
	ids, err := fn(e.hubAddress(), "")
	if err != nil {
		return nil, err
	}

	deals := make([]*pb.Deal, 0, len(ids))
	for _, id := range ids {
		deal, err := e.bc.GetDealInfo(id)
		if err != nil {
			return nil, err
		}

		deals = append(deals, deal)
	}

	return deals, nil
}

func (e *eth) WaitForDealCreated(dealID DealID, buyerID common.Address) (*pb.Deal, error) {
	log.G(e.ctx).Debug("waiting for deal created", zap.Stringer("dealID", dealID))

	id, err := util.ParseBigInt(dealID.String())
	if err != nil {
		return nil, err
	}

	return e.findDeals(e.ctx, id, buyerID)
}

func (e *eth) WaitForDealClosed(ctx context.Context, dealID DealID, buyerID string) error {
	log.G(ctx).Debug("waiting for deal closed", zap.Stringer("dealID", dealID))

	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.G(ctx).Debug("checking whether deal is closed")

			ids, err := e.bc.GetClosedDeal(crypto.PubkeyToAddress(e.key.PublicKey).Hex(), buyerID)
			if err != nil {
				return err
			}

			log.G(ctx).Info("found some closed deals", zap.Int("count", len(ids)))

			for _, id := range ids {
				dealInfo, err := e.bc.GetDealInfo(id)
				if err != nil {
					continue
				}

				if dealInfo.GetId() == dealID.String() && dealInfo.GetStatus() == pb.DealStatus_CLOSED {
					return nil
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (e *eth) findDeals(ctx context.Context, dealID *big.Int, buyerID common.Address) (*pb.Deal, error) {
	ctx, cancel := context.WithTimeout(e.ctx, e.timeout)
	defer cancel()

	tk := time.NewTicker(3 * time.Second)
	defer tk.Stop()

	if deal := e.findDealOnce(buyerID, dealID); deal != nil {
		return deal, nil
	}

	for {
		select {
		case <-tk.C:
			if deal := e.findDealOnce(buyerID, dealID); deal != nil {
				return deal, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (e *eth) findDealOnce(buyerID common.Address, dealID *big.Int) *pb.Deal {
	deal, err := e.bc.GetDealInfo(dealID)
	if err != nil {
		return nil
	}

	log.G(e.ctx).Info("deal found",
		zap.String("dealID", deal.GetId()),
		zap.String("supplierID", deal.GetSupplierID()),
		zap.String("buyerID", deal.GetBuyerID()))

	me := util.PubKeyToAddr(e.key.PublicKey)
	dealSupplier := common.HexToAddress(deal.GetSupplierID())
	dealBuyer := common.HexToAddress(deal.GetBuyerID())

	if auth.EqualAddresses(buyerID, dealBuyer) && auth.EqualAddresses(me, dealSupplier) {
		return deal
	}

	return nil
}

func (e *eth) AcceptDeal(id string) error {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return err
	}

	_, err = e.bc.AcceptDeal(e.key, bigID)
	return err
}

func (e *eth) CloseDeal(id DealID) error {
	bigID, err := util.ParseBigInt(string(id))
	if err != nil {
		return err
	}

	_, err = e.bc.CloseDeal(e.key, bigID)
	return err
}

func (e *eth) GetDeal(id string) (*pb.Deal, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return nil, err
	}

	deal, err := e.bc.GetDealInfo(bigID)
	if err != nil {
		return nil, err
	}

	// NOTE: May GetSupplierID return common.Address?
	idOK := deal.GetSupplierID() == crypto.PubkeyToAddress(e.key.PublicKey).Hex()
	statusOK := deal.GetStatus() == pb.DealStatus_ACCEPTED
	dealOK := idOK && statusOK

	if dealOK {
		return deal, nil
	} else {
		return nil, errDealNotFound
	}
}

// NewETH constructs a new Ethereum client.
func NewETH(ctx context.Context, key *ecdsa.PrivateKey, bcr blockchain.Blockchainer, timeout time.Duration) (ETH, error) {
	var err error
	if bcr == nil {
		bcr, err = blockchain.NewAPI(nil, nil)
		if err != nil {
			return nil, err
		}
	}

	return &eth{
		ctx:     ctx,
		key:     key,
		bc:      bcr,
		timeout: timeout,
	}, nil
}
