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
)

const (
	defaultGasLimit             = 500000
	defaultGasLimitForSidechain = 2000000
)

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

func FindLogByTopic(txReceipt *Receipt, topic common.Hash) (*types.Log, error) {
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

func waitForTransactionResult(ctx context.Context, client EthereumClientBackend, logParsePeriod time.Duration, tx *types.Transaction, topic common.Hash) (*types.Log, error) {
	tk := util.NewImmediateTicker(logParsePeriod)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			var err error
			tmpRec := &Receipt{}
			tmpRec.Receipt, err = client.TransactionReceipt(ctx, tx.Hash())
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}
			logs, err := FindLogByTopic(tmpRec, topic)
			if err != nil {
				return nil, err
			}

			return logs, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func WaitTransactionReceipt(ctx context.Context, client CustomEthereumClient, confirmations int64, logParsePeriod time.Duration, tx *types.Transaction) (*Receipt, error) {
	tk := util.NewImmediateTicker(logParsePeriod)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			lastBlock, err := client.GetLastBlock(ctx)
			if err != nil {
				return nil, err
			}

			txReceipt, err := client.GetTransactionReceipt(ctx, tx.Hash())
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}

			if lastBlock.Int64() < txReceipt.BlockNumber+confirmations {
				break
			}

			return txReceipt, nil
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
