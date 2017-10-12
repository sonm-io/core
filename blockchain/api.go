package blockchain

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sonm-io/core/blockchain/utils"
	"github.com/sonm-io/core/blockchain/tsc/api"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"fmt"
)

const ethEndpoint string = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

var ethClient *ethclient.Client = nil

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
}

func NewBlockchainAPI(key *ecdsa.PrivateKey) (bch *BlockchainAPI, err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return nil, err
		}
	}

	txOpts := bind.NewKeyedTransactor(key)

	txOpts.GasPrice = big.NewInt(9999999999999999)
	txOpts.GasLimit = big.NewInt(5000000)

	bch = &BlockchainAPI{
		prv:    key,
		txOpts: txOpts,
	}
	return bch, nil
}


// todo: repair struct casting
func GetDealInfo(id *big.Int) (deal *struct {
	SpecHach *big.Int
	Client   common.Address
	Hub      common.Address
	Price    *big.Int
	Status   *big.Int
}, err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return nil, err
		}
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	//dealsContract.GetDealInfo(&bind.CallOpts{Pending: true}, id)
	*deal, err = dealsContract.GetDealInfo(&bind.CallOpts{Pending: true}, id)
	if err != nil {
		fmt.Println(err)
	}
	return deal, nil
}

func GetHubDeals(address string) (deals []*big.Int, err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return nil, err
		}
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	deals, err = dealsContract.GetDealByHubAddress(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func GetClientDeals(address string) (deals []*big.Int, err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return nil, err
		}
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	deals, err = dealsContract.GetDealByClient(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return deals, nil
}

func GetDealAmount() (amount *big.Int, err error) {
	if ethClient == nil {
		ethClient, err = utils.InitEthClient(ethEndpoint)
		if err != nil {
			return nil, err
		}
	}

	dealsContract, err := api.NewDeals(common.HexToAddress(tsc.DealsAddress), ethClient)
	if err != nil {
		return nil, err
	}
	res, err := dealsContract.GetDealAmount(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}
