package salesman

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/proto"
)

func NewDumEthAPI() *dumbEthAPI {
	return &dumbEthAPI{
		Deals:  map[string]*sonm.Deal{},
		Orders: map[string]*sonm.Order{},
	}
}

type dumbEthAPI struct {
	mu     sync.RWMutex
	Deals  map[string]*sonm.Deal
	Orders map[string]*sonm.Order
}

func (m *dumbEthAPI) MarketAddress() common.Address {
	return common.HexToAddress("0x0000000000000000000000000000000000000001")
}

func (m *dumbEthAPI) CloseDeal(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int, blacklistType sonm.BlacklistType) error {
	id := dealID.String()
	m.mu.Lock()
	defer m.mu.Unlock()
	deal, ok := m.Deals[id]
	if !ok {
		return fmt.Errorf("deal with id %s is not found", id)
	}
	if deal.Status != sonm.DealStatus_DEAL_ACCEPTED {
		return fmt.Errorf("deal with id %s is not active", id)
	}
	deal.Status = sonm.DealStatus_DEAL_CLOSED
	return nil
}

func (m *dumbEthAPI) GetDealInfo(ctx context.Context, dealID *big.Int) (*sonm.Deal, error) {
	id := dealID.String()
	m.mu.RLock()
	defer m.mu.RUnlock()
	deal, ok := m.Deals[id]
	if !ok {
		return nil, fmt.Errorf("deal with id %s is not found", id)
	}
	return deepcopy.Copy(deal).(*sonm.Deal), nil
}

func (m *dumbEthAPI) PlaceOrder(ctx context.Context, key *ecdsa.PrivateKey, order *sonm.Order) (*sonm.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := len(m.Orders)
	m.Orders[fmt.Sprint(id)] = order
	order.Id = sonm.NewBigIntFromInt(int64(id))
	order.AuthorID = sonm.NewEthAddress(crypto.PubkeyToAddress(key.PublicKey))
	return order, nil

}

func (m *dumbEthAPI) CancelOrder(ctx context.Context, key *ecdsa.PrivateKey, orderID *big.Int) error {
	id := orderID.String()
	m.mu.Lock()
	defer m.mu.Unlock()
	order, ok := m.Orders[id]
	if !ok {
		return fmt.Errorf("order with id %s is not found", id)
	}
	if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		return fmt.Errorf("order with id %s is not active", id)
	}
	order.OrderStatus = sonm.OrderStatus_ORDER_INACTIVE
	return nil
}

func (m *dumbEthAPI) GetOrderInfo(ctx context.Context, orderID *big.Int) (*sonm.Order, error) {
	id := orderID.String()
	m.mu.RLock()
	defer m.mu.RUnlock()
	order, ok := m.Orders[id]
	if !ok {
		return nil, fmt.Errorf("order with id %s is not found", id)
	}
	return deepcopy.Copy(order).(*sonm.Order), nil
}

func (m *dumbEthAPI) Bill(ctx context.Context, key *ecdsa.PrivateKey, dealID *big.Int) error {
	return nil
}

func (m *dumbEthAPI) CreateDealByOrder(ctx context.Context, order *sonm.Order) (*sonm.Deal, error) {
	id := order.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	askOrder, ok := m.Orders[id]
	if !ok {
		return nil, fmt.Errorf("order with id %s is not found", id)
	}
	if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		return nil, fmt.Errorf("order with id %s is not active", id)
	}
	bidOrder := deepcopy.Copy(askOrder).(*sonm.Order)
	bidOrder.OrderType = sonm.OrderType_BID
	orderID := len(m.Orders)
	dealID := len(m.Deals)

	deal := &sonm.Deal{
		Id:         sonm.NewBigIntFromInt(int64(dealID)),
		Benchmarks: deepcopy.Copy(askOrder.Benchmarks).(*sonm.Benchmarks),
		SupplierID: askOrder.AuthorID,
		ConsumerID: bidOrder.AuthorID,
		MasterID:   askOrder.AuthorID,
		AskID:      askOrder.GetId(),
		BidID:      askOrder.GetId(),
		Duration:   askOrder.Duration,
		Price:      askOrder.Price,
		StartTime:  sonm.CurrentTimestamp(),
		Status:     sonm.DealStatus_DEAL_ACCEPTED,
		LastBillTS: sonm.CurrentTimestamp(),
	}
	m.Orders[fmt.Sprint(orderID)] = bidOrder
	m.Deals[fmt.Sprint(dealID)] = deal
	return deepcopy.Copy(deal).(*sonm.Deal), nil
}
