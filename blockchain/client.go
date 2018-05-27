package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Receipt extends go-ethereum/core/types/Receipt struct with BlockNumber
type Receipt struct {
	*types.Receipt
	BlockNumber string
}

// UnmarshalJSON implement json/Unmarshaler interface (stdlib)
// As far as ethereum Receipt has it's own unmarshaller blockNumber of our custom Receipt was never been unmarshalled.
// Introduced UnmarshallJSON method overrides this behaviour.
func (r *Receipt) UnmarshalJSON(input []byte) error {
	// call parent unmarshal
	err := r.Receipt.UnmarshalJSON(input)
	if err != nil {
		return err
	}

	// define pure structure for clearly casting
	type rec struct {
		BlockNumber string `json:"BlockNumber"`
	}
	var dec rec
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	if dec.BlockNumber == "" {
		return fmt.Errorf("unmarshaled block number is empty")
	}

	// assign decoded values
	r.BlockNumber = dec.BlockNumber
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
