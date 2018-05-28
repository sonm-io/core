package blockchain

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sonm-io/core/blockchain/source/api"
)

// ContractBinding contains options for all bindings
type ContractBinding struct {
	address  common.Address
	client   bind.ContractBackend
	gasPrice int64
}

type OracleUSD struct {
	bind     ContractBinding
	contract *api.OracleUSD
}

func NewOracleUSD(address common.Address, client bind.ContractBackend, gasPrice int64) *OracleUSD {
	return &OracleUSD{
		bind: ContractBinding{
			address:  address,
			gasPrice: gasPrice,
			client:   client,
		},
	}
}

func (a *OracleUSD) SetCurrentPrice(ctx context.Context, key *ecdsa.PrivateKey, price *big.Int) (*types.Transaction, error) {
	if err := a.bindContract(); err != nil {
		return nil, err
	}
	opts := getTxOpts(ctx, key, defaultGasLimitForSidechain, a.bind.gasPrice)
	return a.contract.SetCurrentPrice(opts, price)
}

func (a *OracleUSD) GetCurrentPrice(ctx context.Context) (*big.Int, error) {
	if err := a.bindContract(); err != nil {
		return nil, err
	}
	return a.contract.GetCurrentPrice(getCallOptions(ctx))
}

func (a *OracleUSD) bindContract() error {
	if a.contract == nil {
		contract, err := api.NewOracleUSD(a.bind.address, a.bind.client)
		if err != nil {
			return err
		}
		a.contract = contract
	}
	return nil
}
