package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Receipt extends transaction receipt
// with block number on which this transaction
// was mined.
type Receipt struct {
	*types.Receipt
	BlockNumber int64
}

// UnmarshalJSON calls parent unmarshaller for Receipt and
// also unmarshall block number and associate it with the struct.
func (r *Receipt) UnmarshalJSON(input []byte) error {
	// call parent unmarshal
	err := r.Receipt.UnmarshalJSON(input)
	if err != nil {
		return err
	}

	// define temporary struct to unmarshall block number
	type blockNumber struct {
		BlockNumber string `json:"BlockNumber"`
	}

	var dec blockNumber
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	if dec.BlockNumber == "" {
		return fmt.Errorf("unmarshaled block number is empty")
	}

	v, err := strconv.ParseInt(dec.BlockNumber, 16, 64)
	if err != nil {
		return err
	}

	r.BlockNumber = v
	return nil
}

// EthereumClientBackend release all methods to execute interaction with Ethereum Blockchain
type EthereumClientBackend interface {
	bind.ContractBackend
	ethereum.ChainReader
	ethereum.TransactionReader
}

// CustomEthereumClient extends EthereumClientBackend
type CustomEthereumClient interface {
	EthereumClientBackend
	// GetLastBlock returns number of last mined block
	GetLastBlock(ctx context.Context) (*big.Int, error)
	// GetTransactionReceipt returns receipt of mined transaction or notFound if tx not mined
	GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*Receipt, error)
}

type CustomClient struct {
	*ethclient.Client
	c *rpc.Client
}

func NewClient(endpoint string) (CustomEthereumClient, error) {
	tmp := &CustomClient{}
	var err error

	tmp.c, err = rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	tmp.Client = ethclient.NewClient(tmp.c)
	if err != nil {
		return nil, err
	}

	return tmp, nil
}

func (cc *CustomClient) GetLastBlock(ctx context.Context) (*big.Int, error) {
	var result string
	err := cc.c.CallContext(ctx, &result, "eth_blockNumber")
	if err != nil {
		return nil, err
	}
	return common.HexToHash(result).Big(), nil
}

func (cc *CustomClient) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*Receipt, error) {
	result := &Receipt{}
	result.Receipt = &types.Receipt{}
	err := cc.c.CallContext(ctx, &result, "eth_getTransactionReceipt", txHash)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ethereum.NotFound
	}
	return result, nil
}
