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

// Receipt extends transaction receipt
// with block number on which this transaction
// was mined.
type Receipt struct {
	*types.Receipt
	BlockNumber      int64
	BlockHash        common.Hash
	From             common.Address
	To               common.Address
	TransactionIndex uint64
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
	type extendedReceipt struct {
		BlockNumber      string `json:"blockNumber"`
		BlockHash        string `json:"blockHash"`
		From             string `json:"from"`
		To               string `json:"to"`
		TransactionIndex string `json:"transactionIndex"`
	}

	var dec extendedReceipt
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}

	if dec.BlockNumber == "" {
		return fmt.Errorf("unmarshaled block number is empty")
	}
	v := big.NewInt(0).SetBytes(common.FromHex(dec.BlockNumber))
	r.BlockNumber = v.Int64()

	if dec.BlockHash == "" {
		return fmt.Errorf("unmarshaled BlockHash is empty")
	}
	r.BlockHash = common.HexToHash(dec.BlockHash)

	if dec.From == "" {
		return fmt.Errorf("unmarshaled from field is empty")
	}
	r.From = common.HexToAddress(dec.From)

	if dec.To == "" {
		dec.To = "0x0"
	}
	r.To = common.HexToAddress(dec.To)

	txIndexInt := big.NewInt(0).SetBytes(common.Hex2Bytes(dec.TransactionIndex))
	if err != nil {
		return err
	}
	if !txIndexInt.IsUint64() {
		return fmt.Errorf("transaction index overflows uint64")
	}
	r.TransactionIndex = txIndexInt.Uint64()
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
	// GetEthereumBalanceAt returns amount of ethereum on account at given block
	GetEthereumBalanceAt(ctx context.Context, address common.Address, block *big.Int) (*big.Int, error)
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

func (cc *CustomClient) GetEthereumBalanceAt(ctx context.Context, address common.Address, block *big.Int) (*big.Int, error) {
	return cc.Client.BalanceAt(ctx, address, block)
}
