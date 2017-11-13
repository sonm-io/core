package blockchain

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/blockchain/tsc/api"
)

const defaultEthEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

const defaultGasPrice = 20 * 1000000000

// Dealer - interface above SONM deals
// client - who wanna buy
// hub - who wanna selling its resources
// WARN: this may change at future, by any proposal
type Dealer interface {
	// OpenDeal is function to open new deal in blockchain from given address,
	// it have effect to change blockchain state, key is mandatory param
	// other params caused by SONM office's agreement
	// It could be called by client
	// return transaction, not deal id
	OpenDeal(key *ecdsa.PrivateKey, hub string, client string, specificationHash, price, workTime *big.Int) (*types.Transaction, error)

	// AcceptDeal accepting deal by hub, causes that hub accept to sell its resources
	// It could be called by hub
	AcceptDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)
	// CloseDeal closing deal by given id
	// It could be called by client
	CloseDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)

	// GetHubDeals is returns ids by given hub address
	GetHubDeals(address string) ([]*big.Int, error)
	// GetHubDeals is returns ids by given client address
	GetClientDeals(address string) ([]*big.Int, error)
	// GetDealInfo is returns deal info by given id
	GetDealInfo(id *big.Int) (*DealInfo, error)
	// GetDealAmount return global deal counter
	GetDealAmount() (*big.Int, error)
}

// Token is go implementation of ERC20-compatibility token with full functionality high-level interface
// standart description with placed: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20-token-standard.md
type Token interface {
	// Approve - add allowance from caller to other contract to spend tokens
	Approve(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// Transfer token from caller
	Transfer(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// TransferFrom fallback function for contracts to transfer you allowance
	TransferFrom(key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error)

	// BalanceOf returns balance of given address
	BalanceOf(address string) (*big.Int, error)
	// AllowanceOf returns allowance of given address to spender account
	AllowanceOf(from string, to string) (*big.Int, error)
	// TotalSupply - all amount of emitted token
	TotalSupply() (*big.Int, error)
}

func initEthClient(ethEndpoint *string) (*ethclient.Client, error) {
	var endpoint string
	if ethEndpoint == nil {
		endpoint = defaultEthEndpoint
	} else {
		endpoint = *ethEndpoint
	}
	ethClient, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return ethClient, nil
}

func (bch *API) getTxOpts(key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.GasLimit = big.NewInt(gasLimit)
	opts.GasPrice = big.NewInt(bch.gasPrice)
	return opts
}

type DealInfo struct {
	SpecificationHash *big.Int
	Client            string
	Hub               string
	Price             *big.Int
	Status            *big.Int
	StartTime         *big.Int
	WorkTime          *big.Int
	EndTime           *big.Int
}

type API struct {
	client   *ethclient.Client
	gasPrice int64

	dealsContract *api.Deals
	tokenContract *api.TSCToken
}

// NewBlockchainAPI - is
func NewBlockchainAPI(ethEndpoint *string, gasPrice *int64) (*API, error) {
	client, err := initEthClient(ethEndpoint)
	if err != nil {
		return nil, err
	}

	var gp int64
	if gasPrice == nil {
		gp = defaultGasPrice
	} else {
		gp = *gasPrice
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), client)
	if err != nil {
		return nil, err
	}

	tokenContract, err := api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), client)
	if err != nil {
		return nil, err
	}

	bch := &API{
		client:        client,
		gasPrice:      gp,
		dealsContract: dealsContract,
		tokenContract: tokenContract,
	}
	return bch, nil
}

// ----------------
// Deals appearance
// ----------------

func (bch *API) OpenDeal(key *ecdsa.PrivateKey, hub string, client string, specificationHash, price, workTime *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 305000)

	tx, err := bch.dealsContract.OpenDeal(opts, common.HexToAddress(hub), common.HexToAddress(client), specificationHash, price, workTime)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) AcceptDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 90000)

	tx, err := bch.dealsContract.AcceptDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) CloseDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 90000)

	tx, err := bch.dealsContract.CloseDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) GetDeals(address string) ([]*big.Int, error) {
	clientDeals, err := bch.dealsContract.GetDeals(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return clientDeals, nil
}

func (bch *API) GetDealInfo(id *big.Int) (*DealInfo, error) {
	deal, err := bch.dealsContract.GetDealInfo(&bind.CallOpts{Pending: true}, id)
	if err != nil {
		return nil, err
	}
	dealInfo := DealInfo{
		SpecificationHash: deal.SpecHach,
		Client:            deal.Client.String(),
		Hub:               deal.Hub.String(),
		Price:             deal.Price,
		Status:            deal.Status,
		StartTime:         deal.StartTime,
		WorkTime:          deal.WorkTime,
		EndTime:           deal.EndTIme,
	}
	return &dealInfo, nil
}

func (bch *API) GetDealAmount() (*big.Int, error) {
	res, err := bch.dealsContract.GetDealsAmount(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ----------------
// Token appearance
// ----------------

func (bch *API) BalanceOf(address string) (*big.Int, error) {
	balance, err := bch.tokenContract.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (bch *API) AllowanceOf(from string, to string) (*big.Int, error) {
	allowance, err := bch.tokenContract.Allowance(&bind.CallOpts{Pending: true}, common.HexToAddress(from), common.HexToAddress(to))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (bch *API) Approve(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.Approve(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) Transfer(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) TransferFrom(key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *API) TotalSupply() (*big.Int, error) {
	supply, err := bch.tokenContract.TotalSupply(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return supply, nil
}
