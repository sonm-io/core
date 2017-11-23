package blockchain

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sonm-io/core/blockchain/tsc"
	token_api "github.com/sonm-io/core/blockchain/tsc/api"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

const defaultEthEndpoint = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

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
	OpenDeal(key *ecdsa.PrivateKey, deal *pb.Deal) (*types.Transaction, error)

	// AcceptDeal accepting deal by hub, causes that hub accept to sell its resources
	// It could be called by hub
	AcceptDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)
	// CloseDeal closing deal by given id
	// It could be called by client
	CloseDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)

	// GetDeals is returns ids by given address
	GetDeals(address string) ([]*big.Int, error)
	// GetDealInfo is returns deal info by given id
	GetDealInfo(id *big.Int) (*pb.Deal, error)
	// GetDealAmount return global deal counter
	GetDealAmount() (*big.Int, error)
	// GetOpenedDeal returns only opened deal by given hub/client addresses
	GetOpenedDeal(hubAddr *string, clientAddr *string) ([]*big.Int, error)
}

// Tokener is go implementation of ERC20-compatibility token with full functionality high-level interface
// standart description with placed: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20-token-standard.md
type Tokener interface {
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

// Blockchainer interface describes operations with deals and tokens
type Blockchainer interface {
	Dealer
	Tokener
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

func (bch *api) getTxOpts(key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.GasLimit = big.NewInt(gasLimit)
	opts.GasPrice = big.NewInt(bch.gasPrice)
	return opts
}

type api struct {
	client   *ethclient.Client
	gasPrice int64

	dealsContract *token_api.Deals
	tokenContract *token_api.TSCToken
}

// NewAPI builds new Blockchain instance with given endpoint and gas price
func NewAPI(ethEndpoint *string, gasPrice *int64) (Blockchainer, error) {
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

	dealsContract, err := token_api.NewDeals(common.HexToAddress(tsc.DealsAddress), client)
	if err != nil {
		return nil, err
	}

	tokenContract, err := token_api.NewTSCToken(common.HexToAddress(tsc.TSCAddress), client)
	if err != nil {
		return nil, err
	}

	bch := &api{
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

var DealOpenedTopic common.Hash = common.HexToHash("0x873cb35202fef184c9f8ee23c04e36dc38f3e26fb285224ca574a837be976848")
var DealAcceptedTopic common.Hash = common.HexToHash("0x3a38edea6028913403c74ce8433c90eca94f4ca074d318d8cb77be5290ba4f15")
var DealClosedTopic common.Hash = common.HexToHash("0x72615f99a62a6cc2f8452d5c0c9cbc5683995297e1d988f09bb1471d4eefb890")

func (bch *api) GetOpenedDeal(hubAddr *string, clientAddr *string) ([]*big.Int, error) {
	var topics [][]common.Hash

	// precompile EventName topics
	var eventTopic = []common.Hash{DealOpenedTopic, DealAcceptedTopic, DealClosedTopic}
	topics = append(topics, eventTopic)

	// add filter topic by hub address
	// filtering by client address implemented below
	if hubAddr != nil {
		var addrTopic = []common.Hash{common.HexToHash(common.HexToAddress(*hubAddr).String())}
		topics = append(topics, addrTopic)
	}

	logs, err := bch.client.FilterLogs(context.Background(), ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(tsc.DealsAddress)},
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

	// shifts ids of `DealOpened` event to new slice
	idsOpened := make([]common.Hash, 0)
	// ids of `DealAccepted` & `DealClosed` to other map
	mb := make(map[string]bool)

	for _, l := range logs {
		// filtering by client address
		if clientAddr != nil {
			if l.Topics[2] != common.HexToHash(*clientAddr) {
				continue
			}
		}

		idTopic := l.Topics[3]

		switch l.Topics[0] {
		case DealOpenedTopic:
			idsOpened = append(idsOpened, idTopic)
			break
		case DealAcceptedTopic, DealClosedTopic:
			mb[idTopic.String()] = true
			break
		}
	}

	// shift ids of opened deals by accepted and closed deals
	var out []*big.Int
	for _, item := range idsOpened {
		if _, ok := mb[item.String()]; !ok {
			if err != nil {
				continue
			}
			out = append(out, item.Big())
		}
	}

	return out, nil
}

func (bch *api) OpenDeal(key *ecdsa.PrivateKey, deal *pb.Deal) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 305000)

	bigSpec, err := util.ParseBigInt(deal.SpecificationHash)
	if err != nil {
		return nil, err
	}

	bigPrice, err := util.ParseBigInt(deal.Price)
	if err != nil {
		return nil, err
	}

	tx, err := bch.dealsContract.OpenDeal(
		opts,
		common.HexToAddress(deal.GetSupplierID()),
		common.HexToAddress(deal.GetBuyerID()),
		bigSpec,
		bigPrice,
		big.NewInt(int64(deal.GetWorkTime())),
	)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) AcceptDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 90000)

	tx, err := bch.dealsContract.AcceptDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) CloseDeal(key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 90000)

	tx, err := bch.dealsContract.CloseDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) GetDeals(address string) ([]*big.Int, error) {
	clientDeals, err := bch.dealsContract.GetDeals(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return clientDeals, nil
}

func (bch *api) GetDealInfo(id *big.Int) (*pb.Deal, error) {
	deal, err := bch.dealsContract.GetDealInfo(&bind.CallOpts{Pending: true}, id)
	if err != nil {
		return nil, err
	}
	dealInfo := pb.Deal{
		BuyerID:           deal.Client.String(),
		SupplierID:        deal.Hub.String(),
		SpecificationHash: deal.SpecHach.String(),
		Price:             deal.Price.String(),
		Status:            pb.DealStatus(deal.Status.Int64()),
		StartTime:         &pb.Timestamp{Seconds: deal.StartTime.Int64()},
		WorkTime:          deal.WorkTime.Uint64(),
		EndTime:           &pb.Timestamp{Seconds: deal.EndTIme.Int64()},
	}
	return &dealInfo, nil
}

func (bch *api) GetDealAmount() (*big.Int, error) {
	res, err := bch.dealsContract.GetDealsAmount(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ----------------
// Tokener appearance
// ----------------

func (bch *api) BalanceOf(address string) (*big.Int, error) {
	balance, err := bch.tokenContract.BalanceOf(&bind.CallOpts{Pending: true}, common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (bch *api) AllowanceOf(from string, to string) (*big.Int, error) {
	allowance, err := bch.tokenContract.Allowance(&bind.CallOpts{Pending: true}, common.HexToAddress(from), common.HexToAddress(to))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (bch *api) Approve(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.Approve(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) Transfer(key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) TransferFrom(key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.getTxOpts(key, 50000)

	tx, err := bch.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) TotalSupply() (*big.Int, error) {
	supply, err := bch.tokenContract.TotalSupply(&bind.CallOpts{Pending: true})
	if err != nil {
		return nil, err
	}
	return supply, nil
}
