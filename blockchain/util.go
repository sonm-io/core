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
	defaultGasLimit             = 500000
	defaultGasLimitForSidechain = 2000000
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

func extractAddress(topics []common.Hash, pos int) (common.Address, error) {
	if len(topics) < pos {
		return common.Address{}, errors.New("topic index out of range")
	}

	return common.HexToAddress(topics[pos].Hex()), nil
}

func extractBig(topics []common.Hash, pos int) (*big.Int, error) {
	if len(topics) < pos {
		return nil, errors.New("topic index out of range")
	}

	return topics[pos].Big(), nil
}

func findLogByTopic(ctx context.Context, client *ethclient.Client, tx *types.Transaction, topic common.Hash) (*types.Log, error) {
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
		if topicCmp == 0 {
			return l, nil
		}
	}

	// TODO(sshaman1101): not so user-friendly message leaved for debugging, remove before releasing.
	return nil, fmt.Errorf("cannot find topic \"%s\"in transaction", topic.Hex())
}

func waitForTransactionResult(ctx context.Context, client *ethclient.Client, logParsePeriod time.Duration, tx *types.Transaction, topic common.Hash) (*types.Log, error) {
	tk := util.NewImmediateTicker(logParsePeriod)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			log, err := findLogByTopic(ctx, client, tx, topic)
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}

			return log, err
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
