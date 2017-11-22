package hub

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

type ETH interface {
	// WaitForDealCreated waits for deal created on Buyer-side
	WaitForDealCreated(request *structs.DealRequest) (bool, error)

	// AcceptDeal approves deal on Hub-side
	AcceptDeal(id string) error

	// CheckDealExists checks whether a given deal exists.
	CheckDealExists(id string) (bool, error)
}

type eth struct {
	key *ecdsa.PrivateKey
	bc  blockchain.Blockchainer
	ctx context.Context
}

func (e *eth) WaitForDealCreated(request *structs.DealRequest) (bool, error) {
	// e.findDeals blocks until order will be found or timeout will reached
	return e.findDeals(e.ctx, request.BidId, request.SpecHash)
}

func (e *eth) findDeals(ctx context.Context, addr, hash string) (bool, error) {
	// TODO(sshaman1101): make if configurable?
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	tk := time.NewTicker(3 * time.Second)
	defer tk.Stop()

	if found := e.findDealOnce(addr, hash); found {
		return true, nil
	}

	for {
		select {
		case <-tk.C:
			if found := e.findDealOnce(addr, hash); found {
				return true, nil
			}
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
}

func (e *eth) findDealOnce(addr, hash string) bool {
	// get deals opened by our client
	IDs, err := e.bc.GetDeals(addr)
	if err != nil {
		return false
	}

	for _, id := range IDs {
		// then get extended info
		deal, err := e.bc.GetDealInfo(id)
		if err != nil {
			continue
		}

		// then check for status
		// and check if task hash is equal with request's one
		if deal.GetStatus() == pb.DealStatus_PENDING {
			if deal.GetSpecificationHash() == hash {
				return true
			}
		}
	}

	return false
}

func (e *eth) AcceptDeal(id string) error {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return err
	}

	_, err = e.bc.AcceptDeal(e.key, bigID)
	return err
}

func (e *eth) CheckDealExists(id string) (bool, error) {
	bigID, err := util.ParseBigInt(id)
	if err != nil {
		return false, err
	}

	deal, err := e.bc.GetDealInfo(bigID)
	if err != nil {
		return false, err
	}

	idOK := deal.GetSupplierID() == util.PubKeyToAddr(e.key.PublicKey)
	statusOK := deal.GetStatus() == pb.DealStatus_ACCEPTED
	dealOK := idOK && statusOK

	return dealOK, nil
}

// NewETH constructs a new Ethereum client.
func NewETH(ctx context.Context, key *ecdsa.PrivateKey) (ETH, error) {
	bcAPI, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	return &eth{
		ctx: ctx,
		key: key,
		bc:  bcAPI,
	}, nil
}
