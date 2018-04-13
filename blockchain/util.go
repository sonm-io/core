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
	// TODO: need to move this values into blockchain/merket/addresses.go
	DealOpenedTopic          = common.HexToHash("0xb9ffc65567b7238dd641372277b8c93ed03df73945932dd84fd3cbb33f3eddbf")
	DealUpdatedTopic         = common.HexToHash("0x0b27183934cfdbeb1fbbe288c2e163ed7aa8f458a954054970f78446bccb36e0")
	OrderPlacedTopic         = common.HexToHash("0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d")
	OrderUpdatedTopic        = common.HexToHash("0xb8b459bc0688c37baf5f735d17f1711684bc14ab7db116f88bc18bf409b9309a")
	DealChangeRequestSent    = common.HexToHash("0x7ff56b2eb3ce318aad93d0ba39a3e4a406992a136f9554f17f6bcc43509275d1")
	DealChangeRequestUpdated = common.HexToHash("0x4b92d35447745e95b7344414a41ae94984787d0ebcd2c12021169197bb59af39")
	WorkerAnnouncedTopic     = common.HexToHash("0xe398d33bf7e881cdfc9f34c743822904d4e45a0be0db740dd88cb132e4ce2ed9")
	WorkerConfirmedTopic     = common.HexToHash("0x4940ef08d5aed63b7d3d3db293d69d6ed1d624995b90e9e944839c8ea0ae450d")
	WorkerRemovedTopic       = common.HexToHash("0x7822736ed69a5fe0ad6dc2c6669e8053495d711118e5435b047f9b83deda4c37")
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
