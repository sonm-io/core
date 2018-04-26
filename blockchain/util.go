package blockchain

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/util"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	defaultGasLimit             = 360000
	gasLimitForPlaceOrderMethod = 1000000
)

var (
	// TODO: need to move this values into blockchain/merket/addresses.go
	DealOpenedTopic               = common.HexToHash("0xb9ffc65567b7238dd641372277b8c93ed03df73945932dd84fd3cbb33f3eddbf")
	DealUpdatedTopic              = common.HexToHash("0x0b27183934cfdbeb1fbbe288c2e163ed7aa8f458a954054970f78446bccb36e0")
	OrderPlacedTopic              = common.HexToHash("0xffa896d8919f0556f53ace1395617969a3b53ab5271a085e28ac0c4a3724e63d")
	OrderUpdatedTopic             = common.HexToHash("0xb8b459bc0688c37baf5f735d17f1711684bc14ab7db116f88bc18bf409b9309a")
	DealChangeRequestSentTopic    = common.HexToHash("0x7ff56b2eb3ce318aad93d0ba39a3e4a406992a136f9554f17f6bcc43509275d1")
	DealChangeRequestUpdatedTopic = common.HexToHash("0x4b92d35447745e95b7344414a41ae94984787d0ebcd2c12021169197bb59af39")
	BilledTopic                   = common.HexToHash("0x51f87cd83a2ce6c4ff7957861f7aba400dc3857d2325e0c94cc69f468874515c")
	WorkerAnnouncedTopic          = common.HexToHash("0xe398d33bf7e881cdfc9f34c743822904d4e45a0be0db740dd88cb132e4ce2ed9")
	WorkerConfirmedTopic          = common.HexToHash("0x4940ef08d5aed63b7d3d3db293d69d6ed1d624995b90e9e944839c8ea0ae450d")
	WorkerRemovedTopic            = common.HexToHash("0x7822736ed69a5fe0ad6dc2c6669e8053495d711118e5435b047f9b83deda4c37")
	AddedToBlacklistTopic         = common.HexToHash("0x708802ac7da0a63d9f6b2df693b53345ad263e42d74c245110e1ec1e03a1567e")
	RemovedFromBlacklistTopic     = common.HexToHash("0x576a9aef294e1b4baf3617fde4cbc80ba5344d5eb508222f29e558981704a457")
	ValidatorCreatedTopic         = common.HexToHash("0x02db26aafd16e8ecd93c4fa202917d50b1693f30b1594e57f7a432ede944eefc")
	ValidatorDeletedTopic         = common.HexToHash("0xa7a579573d398d7b67cd7450121bb250bbd060b29eabafdebc3ce0918658635c")
	CertificateCreatedTopic       = common.HexToHash("0xb9bb1df26fde5c1295a7ccd167330e5d6cb9df14fe4c3884669a64433cc9e760")
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

func extractAddress(log types.Log, pos int) (common.Address, error) {
	if len(log.Topics) < pos {
		return common.Address{}, errors.New("topic index out of range")
	}

	return common.HexToAddress(log.Topics[pos].Hex()), nil
}

func extractBig(log types.Log, pos int) (*big.Int, error) {
	if len(log.Topics) < pos {
		return nil, errors.New("topic index out of range")
	}

	return log.Topics[pos].Big(), nil
}

func parseTransactionLogs(ctx context.Context, client *ethclient.Client, tx *types.Transaction, topic common.Hash) (*big.Int, error) {
	txReceipt, err := client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	if txReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction failed")
	}

	for _, l := range txReceipt.Logs {
		if len(l.Topics) < 1 {
			return nil, errors.New("transaction topics is malformed")
		}

		receivedTopic := l.Topics[0]
		topicCmp := bytes.Compare(receivedTopic.Bytes(), topic.Bytes())
		if topicCmp == 0 && len(l.Topics) > 1 {
			return l.Topics[1].Big(), nil
		}
	}

	// TODO(sshaman1101): not so user-friendly message leaved for debugging, remove before releasing.
	return nil, fmt.Errorf("cannot find topic \"%s\"in transaction", topic.Hex())
}

func waitForTransactionResult(ctx context.Context, client *ethclient.Client, logParsePeriod time.Duration, tx *types.Transaction, topic common.Hash) (*big.Int, error) {
	tk := util.NewImmediateTicker(logParsePeriod)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			id, err := parseTransactionLogs(ctx, client, tx, topic)
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}

			return id, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func getTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit uint64, gasPrice int64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	opts.GasLimit = gasLimit
	opts.GasPrice = big.NewInt(gasPrice)
	return opts
}
