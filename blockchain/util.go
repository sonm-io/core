package blockchain

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	CertsAddress    = "0x1"
	defaultGasLimit = 360000
)

var (
	DealOpenedTopic          = common.HexToHash("0x1")
	DealUpdatedTopic         = common.HexToHash("0x2")
	OrderPlacedTopic         = common.HexToHash("0x3")
	OrderUpdatedTopic        = common.HexToHash("0x4")
	DealChangeRequestSent    = common.HexToHash("0x5")
	DealChangeRequestUpdated = common.HexToHash("0x6")
	WorkerAnnouncedTopic     = common.HexToHash("0x7")
	WorkerConfirmedTopic     = common.HexToHash("0x8")
	WorkerRemovedTopic       = common.HexToHash("0x9")
)

func initEthClient(endpoint string) (*ethclient.Client, error) {
	ethClient, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	return ethClient, nil
}

func getCallOptions(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{
		Pending: true,
		Context: ctx,
	}
}
