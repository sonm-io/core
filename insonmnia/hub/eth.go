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
	WaitForDealCreated(request *structs.DealRequest) (*pb.Deal, error)

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

func (e *eth) WaitForDealCreated(request *structs.DealRequest) (*pb.Deal, error) {
	// e.findDeals blocks until order will be found or timeout will reached
	return e.findDeals(e.ctx, request.BidId, request.SpecHash)
}

func (e *eth) findDeals(ctx context.Context, addr, hash string) (*pb.Deal, error) {
	// TODO(sshaman1101): make if configurable?
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	tk := time.NewTicker(3 * time.Second)
	defer tk.Stop()

	if deal := e.findDealOnce(addr, hash); deal != nil {
		return deal, nil
	}

	for {
		select {
		case <-tk.C:
			if deal := e.findDealOnce(addr, hash); deal != nil {
				return deal, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (e *eth) findDealOnce(addr, hash string) *pb.Deal {
	// get deals opened by our client
	IDs, err := e.bc.GetDeals(addr)
	if err != nil {
		return nil
	}

	for _, id := range IDs {
		// then get extended info
		deal, err := e.bc.GetDealInfo(id)
		if err != nil {
			continue
		}

		// then check for status
		// and check if task hash is equal with request's one
		if deal.GetStatus() == pb.DealStatus_PENDING && deal.GetSpecificationHash() == hash {
			return deal
		}
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
