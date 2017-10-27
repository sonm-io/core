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

type BlockchainAPI struct {
	prv    *ecdsa.PrivateKey
	txOpts *bind.TransactOpts
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
	}
	return bch, nil
}

// ----------------
// Deals appearance
// ----------------

func connectDeals() (*api.Deals, error) {
	err := initEthClient()
	if err != nil {
		return nil, err
	}
	return api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
}

func (bch *BlockchainAPI) OpenDeal(hub string, client string, specificationHash *big.Int, price *big.Int, workTime *big.Int) (*types.Transaction, error) {
	deals, err := connectDeals()
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
	deals, err := connectDeals()
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
	deals, err := connectDeals()
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

func GetHubDeals(address string) ([]*big.Int, error) {
	deals, err := connectDeals()
	if err != nil {
		return nil, err
	}

	hubDeals, err := deals.GetDealsByHubAddress(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return hubDeals, nil

}

func GetClientDeals(address string) ([]*big.Int, error) {
	deals, err := connectDeals()
	if err != nil {
		return nil, err
	}

	clientDeals, err := deals.GetDealsByClient(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return clientDeals, nil
}

func GetDealInfo(id *big.Int) (*DealInfo, error) {
	deals, err := connectDeals()
	if err != nil {
		return nil, err
	}

	deal, err := deals.GetDealInfo(&bind.CallOpts{Pending: true}, id)
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

func GetDealAmount() (amount *big.Int, err error) {
	deals, err := connectDeals()
	if err != nil {
		return nil, err
	}

	res, err := deals.GetDealsAmount(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ----------------
// Token appearance
// ----------------

func connectToken() (*api.TSCToken, error) {
	err := initEthClient()
	if err != nil {
		return nil, err
	}
	return api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), ethClient)
}

func BalanceOf(address string) (balance *big.Int, err error) {
	token, err := connectToken()
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
	token, err := connectToken()
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
	token, err := connectToken()
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
	token, err := connectToken()
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
	token, err := connectToken()
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
