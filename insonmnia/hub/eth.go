package hub

import (
	"context"
	"crypto/ecdsa"

	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
)

type ETH interface {
	// WaitForDealCreated waits for deal created on Buyer-side
	WaitForDealCreated(deal *pb.Deal) (bool, error)

	// ApproveDeal approves deal on Hub-side
	ApproveDeal(deal *pb.Deal) error

	// CheckDealExists checks whether a given deal exists.
	CheckDealExists(deal *pb.Deal) (bool, error)
}

type eth struct {
	key *ecdsa.PrivateKey
	bc  *blockchain.API
	ctx context.Context
}

func (e *eth) WaitForDealCreated(deal *pb.Deal) (bool, error) {
	// TODO(sshaman1101): implement eth calls
	return true, nil
}

func (e *eth) ApproveDeal(deal *pb.Deal) error {
	// TODO(sshaman1101): implement eth calls
	return nil
}

func (e *eth) CheckDealExists(deal *pb.Deal) (bool, error) {
	// TODO(sshaman1101): implement eth calls
	return true, nil
}

// NewETH constructs a new Ethereum client.
func NewETH(ctx context.Context, key *ecdsa.PrivateKey) (ETH, error) {
	bcAPI, err := blockchain.NewBlockchainAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	return &eth{
		ctx: ctx,
		key: key,
		bc:  bcAPI,
	}, nil
}
