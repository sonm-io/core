package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Receipt struct {
	*types.Receipt
	BlockNumber string
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
