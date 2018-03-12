package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sonm-io/core/blockchain/tsc"
	token_api "github.com/sonm-io/core/blockchain/tsc/api"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

const defaultEthEndpoint = "https://rinkeby.infura.io/00iTrs5PIy0uGODwcsrb"

const defaultGasPrice = 20 * 1000000000

// Dealer interface describe ethereum-backed deals into the SONM network.
//
// Entities to operate with:
// client - a person who wanna buy resources
// hub - a person who wanna sell their resources
type Dealer interface {
	// OpenDeal is function to open new deal in blockchain from given address,
	// it have effect to change blockchain state, key is mandatory param
	// other params caused by SONM office's agreement
	// It could be called by client
	// return transaction, not deal id
	OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, deal *pb.Deal) (*types.Transaction, error)
	// OpenDealPending creates deal and waits for transaction to be committed on blockchain.
	// wait is duration to wait for transaction commit, recommended value is 180 seconds.
	OpenDealPending(ctx context.Context, key *ecdsa.PrivateKey, deal *pb.Deal, wait time.Duration) (*big.Int, error)

	// AcceptDeal accepting deal by hub, causes that hub accept to sell its resources
	// It could be called by hub
	AcceptDeal(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)
	// AcceptDealPending accept deal and waits for transaction to be committed on blockchain.
	// wait is duration to wait for transaction commit, recommended value is 180 seconds.
	AcceptDealPending(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int, wait time.Duration) error

	// CloseDeal closing deal by given id
	// It could be called by client
	CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error)
	// CloseDealPending close deal and waits for transaction to be committed on blockchain.
	// wait is duration to wait for transaction commit, recommended value is 180 seconds.
	CloseDealPending(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int, wait time.Duration) error

	// GetDeals is returns ids by given address
	GetDeals(ctx context.Context, address string) ([]*big.Int, error)
	// GetDealInfo is returns deal info by given id
	GetDealInfo(ctx context.Context, id *big.Int) (*pb.Deal, error)
	// GetDealAmount return global deal counter
	GetDealAmount(ctx context.Context) (*big.Int, error)
	// GetOpenedDeal returns only opened deals by given hub/client addresses
	GetOpenedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error)
	// GetAcceptedDeal returns only accepted deals by given hub/client addresses
	GetAcceptedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error)
	// GetClosedDeal returns only closed deals by given hub/client addresses
	GetClosedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error)
}

// Tokener is go implementation of ERC20-compatibility token with full functionality high-level interface
// standart description with placed: https://github.com/ethereum/EIPs/blob/master/EIPS/eip-20-token-standard.md
type Tokener interface {
	// Approve - add allowance from caller to other contract to spend tokens
	Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// Transfer token from caller
	Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error)
	// TransferFrom fallback function for contracts to transfer you allowance
	TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error)

	// BalanceOf returns balance of given address
	BalanceOf(ctx context.Context, address string) (*big.Int, error)
	// AllowanceOf returns allowance of given address to spender account
	AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error)
	// TotalSupply - all amount of emitted token
	TotalSupply(ctx context.Context) (*big.Int, error)
	// GetTokens - send 100 SNMT token for message caller
	// this function added for MVP purposes and has been deleted later
	GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error)
}

// Blockchainer interface describes operations with deals and tokens
type Blockchainer interface {
	Dealer
	Tokener
	// GetTxOpts return transaction options that used to perform operations into Ethereum blockchain
	GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts
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

func (bch *api) GetTxOpts(ctx context.Context, key *ecdsa.PrivateKey, gasLimit int64) *bind.TransactOpts {
	opts := bind.NewKeyedTransactor(key)
	opts.Context = ctx
	opts.GasLimit = big.NewInt(gasLimit)
	opts.GasPrice = big.NewInt(bch.gasPrice)
	return opts
}

func getCallOptions(ctx context.Context) *bind.CallOpts {
	return &bind.CallOpts{
		Pending: true,
		Context: ctx,
	}
}

type api struct {
	client   *ethclient.Client
	gasPrice int64

	dealsContract *token_api.Deals
	tokenContract *token_api.SNMTToken
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

	tokenContract, err := token_api.NewSNMTToken(common.HexToAddress(tsc.SNMTAddress), client)
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

var DealOpenedTopic = common.HexToHash("0x873cb35202fef184c9f8ee23c04e36dc38f3e26fb285224ca574a837be976848")
var DealAcceptedTopic = common.HexToHash("0x3a38edea6028913403c74ce8433c90eca94f4ca074d318d8cb77be5290ba4f15")
var DealClosedTopic = common.HexToHash("0x72615f99a62a6cc2f8452d5c0c9cbc5683995297e1d988f09bb1471d4eefb890")

func (bch *api) OpenDeal(ctx context.Context, key *ecdsa.PrivateKey, deal *pb.Deal) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 360000)

	bigSpec, err := util.ParseBigInt(deal.SpecificationHash)
	if err != nil {
		return nil, err
	}

	tx, err := bch.dealsContract.OpenDeal(
		opts,
		common.HexToAddress(deal.GetSupplierID()),
		common.HexToAddress(deal.GetBuyerID()),
		bigSpec,
		deal.Price.Unwrap(),
		big.NewInt(int64(deal.GetWorkTime())),
	)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) checkTransactionResult(ctx context.Context, tx *types.Transaction) (*big.Int, error) {
	txReceipt, err := bch.client.TransactionReceipt(ctx, tx.Hash())
	if err != nil {
		return nil, err
	}

	if txReceipt.Status != types.ReceiptStatusSuccessful {
		return nil, errors.New("transaction failed")
	}

	for _, l := range txReceipt.Logs {
		if len(l.Topics) < 1 {
			return nil, errors.New("transaction topics is malformed")
		}

		nameTopic := l.Topics[0]
		topicMatched := (nameTopic == DealOpenedTopic) || (nameTopic == DealAcceptedTopic) || (nameTopic == DealClosedTopic)
		if topicMatched && len(l.Topics) > 3 {
			return l.Topics[3].Big(), nil
		}
	}

	return nil, errors.New("cannot find the DealOpened topic in transaction")
}

func (bch *api) OpenDealPending(ctx context.Context, key *ecdsa.PrivateKey, deal *pb.Deal, wait time.Duration) (*big.Int, error) {
	tx, err := bch.OpenDeal(ctx, key, deal)
	if err != nil {
		return nil, err
	}

	id, err := bch.checkTransactionResult(ctx, tx)
	if err != nil {
		// if transaction status is NOT FOUND, then just wait for next tick
		// and try to find it again.
		if err != ethereum.NotFound {
			return nil, err
		}
	} else {
		return id, err
	}

	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			id, err := bch.checkTransactionResult(ctx, tx)
			if err != nil {
				if err == ethereum.NotFound {
					break
				}
				return nil, err
			}

			return id, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (bch *api) AcceptDeal(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 90000)

	tx, err := bch.dealsContract.AcceptDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) AcceptDealPending(ctx context.Context, key *ecdsa.PrivateKey, dealId *big.Int, wait time.Duration) error {
	tx, err := bch.AcceptDeal(ctx, key, dealId)
	if err != nil {
		return err
	}

	id, err := bch.checkTransactionResult(ctx, tx)
	if err != nil {
		// if transaction status is NOT FOUND, then just wait for next tick
		// and try to find it again.
		if err != ethereum.NotFound {
			return err
		}
	} else {
		if id.Cmp(dealId) != 0 {
			return errors.New("given transaction is malformed, dealId is mismatch")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			id, err := bch.checkTransactionResult(ctx, tx)
			if err != nil {
				// if transaction status is NOT FOUND, then just wait for next tick
				// and try to find it again.
				if err != ethereum.NotFound {
					return err
				}
			} else {
				if id != dealId {
					return errors.New("given transaction is malformed, dealId is mismatch")
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (bch *api) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, id *big.Int) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 300000)

	tx, err := bch.dealsContract.CloseDeal(opts, id)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) CloseDealPending(ctx context.Context, key *ecdsa.PrivateKey, dealId *big.Int, wait time.Duration) error {
	tx, err := bch.CloseDeal(ctx, key, dealId)
	if err != nil {
		return err
	}

	id, err := bch.checkTransactionResult(ctx, tx)
	if err != nil {
		// if transaction status is NOT FOUND, then just wait for next tick
		// and try to find it again.
		if err != ethereum.NotFound {
			return err
		}
	} else {
		if id.Cmp(dealId) != 0 {
			return errors.New("given transaction is malformed, dealId is mismatch")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	tk := time.NewTicker(1 * time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			id, err := bch.checkTransactionResult(ctx, tx)
			if err != nil {
				// if transaction status is NOT FOUND, then just wait for next tick
				// and try to find it again.
				if err != ethereum.NotFound {
					return err
				}
			} else {
				if id.Cmp(dealId) != 0 {
					return errors.New("given transaction is malformed, dealId is mismatch")
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (bch *api) GetOpenedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error) {
	var topics [][]common.Hash

	// precompile EventName topics
	var eventTopic = []common.Hash{DealOpenedTopic, DealAcceptedTopic, DealClosedTopic}
	topics = append(topics, eventTopic)

	// add filter topic by hub address
	// filtering by client address implemented below
	if hubAddr != "" {
		var addrTopic = []common.Hash{common.HexToHash(common.HexToAddress(hubAddr).String())}
		topics = append(topics, addrTopic)
	}

	logs, err := bch.client.FilterLogs(ctx, ethereum.FilterQuery{
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
		if clientAddr != "" {
			if l.Topics[2] != common.HexToHash(clientAddr) {
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

func (bch *api) GetAcceptedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error) {
	var topics [][]common.Hash

	// precompile EventName topics
	var eventTopic = []common.Hash{DealAcceptedTopic, DealClosedTopic}
	topics = append(topics, eventTopic)

	// add filter topic by hub address
	// filtering by client address implemented below
	if hubAddr != "" {
		var addrTopic = []common.Hash{common.HexToHash(common.HexToAddress(hubAddr).String())}
		topics = append(topics, addrTopic)
	}

	logs, err := bch.client.FilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(tsc.DealsAddress)},
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

	// shifts ids of `DealAccepted` event to new slice
	idsOpened := make([]common.Hash, 0)
	// ids of `DealClosed` to other map
	mb := make(map[string]bool)

	for _, l := range logs {
		// filtering by client address
		if clientAddr != "" {
			if l.Topics[2] != common.HexToHash(clientAddr) {
				continue
			}
		}

		idTopic := l.Topics[3]

		switch l.Topics[0] {
		case DealAcceptedTopic:
			idsOpened = append(idsOpened, idTopic)
			break
		case DealClosedTopic:
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

func (bch *api) GetClosedDeal(ctx context.Context, hubAddr string, clientAddr string) ([]*big.Int, error) {
	var topics [][]common.Hash

	// precompile EventName topics
	var eventTopic = []common.Hash{DealClosedTopic}
	topics = append(topics, eventTopic)

	// add filter topic by hub address
	// filtering by client address implemented below
	if hubAddr != "" {
		var addrTopic = []common.Hash{common.HexToHash(common.HexToAddress(hubAddr).String())}
		topics = append(topics, addrTopic)
	}

	logs, err := bch.client.FilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(tsc.DealsAddress)},
		Topics:    topics,
	})
	if err != nil {
		return nil, err
	}

	var out []*big.Int

	for _, l := range logs {
		// filtering by client address
		if clientAddr != "" {
			if l.Topics[2] != common.HexToHash(clientAddr) {
				continue
			}
		}
		out = append(out, l.Topics[3].Big())
	}

	return out, nil
}

func (bch *api) GetDeals(ctx context.Context, address string) ([]*big.Int, error) {
	clientDeals, err := bch.dealsContract.GetDeals(getCallOptions(ctx), common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return clientDeals, nil
}

func (bch *api) GetDealInfo(ctx context.Context, id *big.Int) (*pb.Deal, error) {
	deal, err := bch.dealsContract.GetDealInfo(getCallOptions(ctx), id)
	if err != nil {
		return nil, err
	}

	dealInfo := pb.Deal{
		Id:                id.String(),
		BuyerID:           deal.Client.String(),
		SupplierID:        deal.Hub.String(),
		SpecificationHash: deal.SpecHach.String(),
		Price:             pb.NewBigInt(deal.Price),
		Status:            pb.DealStatus(deal.Status.Int64()),
		StartTime:         &pb.Timestamp{Seconds: deal.StartTime.Int64()},
		WorkTime:          deal.WorkTime.Uint64(),
		EndTime:           &pb.Timestamp{Seconds: deal.EndTIme.Int64()},
	}
	return &dealInfo, nil
}

func (bch *api) GetDealAmount(ctx context.Context) (*big.Int, error) {
	res, err := bch.dealsContract.GetDealsAmount(getCallOptions(ctx))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ----------------
// Tokener appearance
// ----------------

func (bch *api) BalanceOf(ctx context.Context, address string) (*big.Int, error) {
	balance, err := bch.tokenContract.BalanceOf(getCallOptions(ctx), common.HexToAddress(address))
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (bch *api) AllowanceOf(ctx context.Context, from string, to string) (*big.Int, error) {
	allowance, err := bch.tokenContract.Allowance(getCallOptions(ctx), common.HexToAddress(from), common.HexToAddress(to))
	if err != nil {
		return nil, err
	}
	return allowance, nil
}

func (bch *api) Approve(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 50000)

	tx, err := bch.tokenContract.Approve(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) Transfer(ctx context.Context, key *ecdsa.PrivateKey, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 50000)

	tx, err := bch.tokenContract.Transfer(opts, common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) TransferFrom(ctx context.Context, key *ecdsa.PrivateKey, from string, to string, amount *big.Int) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 50000)

	tx, err := bch.tokenContract.TransferFrom(opts, common.HexToAddress(from), common.HexToAddress(to), amount)
	if err != nil {
		return nil, err
	}
	return tx, err
}

func (bch *api) TotalSupply(ctx context.Context) (*big.Int, error) {
	supply, err := bch.tokenContract.TotalSupply(getCallOptions(ctx))
	if err != nil {
		return nil, err
	}
	return supply, nil
}

func (bch *api) GetTokens(ctx context.Context, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	opts := bch.GetTxOpts(ctx, key, 50000)

	tx, err := bch.tokenContract.GetTokens(opts)
	if err != nil {
		return nil, err
	}
	return tx, err
}
