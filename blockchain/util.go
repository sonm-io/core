package blockchain

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/sonm-io/core/util"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func getCallOptions(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{
		Pending: false,
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

// WaitTxAndExtractLog await transaction mining and extract log contains for given topic
// this func is composition of FindLogByTopic and WaitTransactionReceipt
func WaitTxAndExtractLog(ctx context.Context, client CustomEthereumClient, confirmations int64, logParsePeriod time.Duration, tx *types.Transaction, topic common.Hash) (*types.Log, error) {
	receipt, err := WaitTransactionReceipt(ctx, client, confirmations, logParsePeriod, tx)
	if err != nil {
		return nil, err
	}

	txLog, err := FindLogByTopic(receipt, topic)
	if err != nil {
		return nil, err
	}

	return txLog, nil
}

// FindLogByTopic safety search log in transaction receipt
// return error if transaction failed
// return error if topic doesn't contain in receipt
func FindLogByTopic(txReceipt *Receipt, topic common.Hash) (*types.Log, error) {
	if txReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction failed")
	}

	for _, l := range txReceipt.Logs {
		if len(l.Topics) < 1 {
			return nil, errors.New("transaction topics is malformed")
		}
		receivedTopic := l.Topics[0]
		if bytes.Compare(receivedTopic.Bytes(), topic.Bytes()) == 0 {
			return l, nil
		}
	}

	// TODO(sshaman1101): not so user-friendly message leaved for debugging, remove before releasing.
	return nil, fmt.Errorf("cannot find topic \"%s\"in transaction receipt", topic.Hex())
}

// WaitTransactionReceipt await transaction with confirmations
// returns Receipt of completed transaction
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
