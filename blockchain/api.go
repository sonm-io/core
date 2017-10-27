package blockchain

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/blockchain/tsc/api"
	"github.com/sonm-io/core/blockchain/utils"
	"math/big"
)

const ethEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

var ethClient *ethclient.Client = nil

func initEthClient() (err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

type Deal struct {
	SpecHach *big.Int
	Client   common.Address
	Hub      common.Address
	Price    *big.Int
	Status   *big.Int
}

type BlockchainAPI struct {
	prv    *ecdsa.PrivateKey
	txOpts *bind.TransactOpts
	client *ethclient.Client
}

func NewBlockchainAPI(key *ecdsa.PrivateKey, gasPrice *big.Int) (bch *BlockchainAPI, err error) {
	err = initEthClient()
	if err != nil {
		return nil, err
	}

	txOpts := bind.NewKeyedTransactor(key)

	if gasPrice == nil {
		gasPrice = big.NewInt(20 * 1000000000)
	}

	txOpts.GasPrice = gasPrice
	txOpts.GasLimit = big.NewInt(300000)

	bch = &BlockchainAPI{
		prv:    key,
		txOpts: txOpts,
		client: ethClient,
	}
	return bch, nil
}

// ----------------
// Deals appearance
// ----------------

func (bch *BlockchainAPI) OpenDeal(hub string, client string, specificationHash *big.Int, price *big.Int, workTime *big.Int) (*types.Transaction, error) {
	deals, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(305000)

	tx, err := deals.OpenDeal(opts, common.HexToAddress(hub), common.HexToAddress(client), specificationHash, price, workTime)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *BlockchainAPI) AcceptDeal(id big.Int) (*types.Transaction, error) {
	deals, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(90000)

	tx, err := deals.AcceptDeal(opts, &id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *BlockchainAPI) CloseDeal(id *big.Int) (*types.Transaction, error) {
	deals, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(90000)

	tx, err := deals.CloseDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func GetHubDeals(address string) (deals []*big.Int, err error) {
	initEthClient()

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	deals, err = dealsContract.GetDealsByHubAddress(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func GetClientDeals(address string) (deals []*big.Int, err error) {
	err = initEthClient()
	if err != nil {
		return nil, err
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	deals, err = dealsContract.GetDealsByClient(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func GetDealAmount() (amount *big.Int, err error) {
	err = initEthClient()
	if err != nil {
		return nil, err
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	res, err := dealsContract.GetDealsAmount(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ----------------
// Token appearance
// ----------------

func BalanceOf(address string) (balance *big.Int, err error) {
	err = initEthClient()
	if err != nil {
		return nil, err
	}

	token, err := api.NewTSCToken(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	balance, err = token.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func AllowanceOf(from string, to string) (allowance *big.Int, err error) {
	err = initEthClient()
	if err != nil {
		return nil, err
	}

	token, err := api.NewTSCToken(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	allowance, err = token.Allowance(&bind.CallOpts{Pending: true}, common.HexToAddress(from), common.HexToAddress(to))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (bch *BlockchainAPI) Approve(to string, amount *big.Int) (*types.Transaction, error) {
	token, err := api.NewTSCToken(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(50000)

	tx, err := token.Approve(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *BlockchainAPI) Transfer(to string, amount *big.Int) (*types.Transaction, error) {
	token, err := api.NewTSCToken(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(50000)

	tx, err := token.Transfer(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *BlockchainAPI) TransferFrom(from string, to string, amount *big.Int) (*types.Transaction, error) {
	token, err := api.NewTSCToken(common.HexToAddress(tsc.DealsAddress), bch.client)
	if err != nil {
		return nil, err
	}
	opts := bch.txOpts
	opts.GasLimit = big.NewInt(50000)

	tx, err := token.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}
