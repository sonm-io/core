package marketplace

import (
	"errors"
	"net"
	"sync"

	"sort"

	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	defaultResultsCount = 100
)

var (
	errOrderNotFound     = errors.New("Order cannot be found")
	errPriceIsZero       = errors.New("Order price cannot be less or equal than zero")
	errOrderIsNil        = errors.New("Order cannot be nil")
	errSlotIsNil         = errors.New("Order slot cannot be nil")
	errResourcesIsNil    = errors.New("Slot resources cannot be nil")
	errStartTimeAfterEnd = errors.New("Start time is after end time")
	errStartTimeRequired = errors.New("Start time is required")
	errEndTimeRequired   = errors.New("End time is required")
	errSearchParamsIsNil = errors.New("Search params cannot be nil")
)

// searchParams holds all fields that using to search on the market
// Preferring to extend this structure instead of increasing amount
// of params that accepting by OrderStorage.GetOrders() function
type searchParams struct {
	slot      *structs.Slot
	orderType pb.OrderType
	count     uint64
}

type OrderStorage interface {
	GetOrders(c *searchParams) ([]*structs.Order, error)
	GetOrderByID(id string) (*structs.Order, error)
	CreateOrder(order *structs.Order) (*structs.Order, error)
	DeleteOrder(id string) error
}

type inMemOrderStorage struct {
	sync.RWMutex
	db map[string]*structs.Order
}

func (in *inMemOrderStorage) generateID() string {
	return uuid.New()
}

func (in *inMemOrderStorage) GetOrders(c *searchParams) ([]*structs.Order, error) {
	if c == nil {
		return nil, errSearchParamsIsNil
	}

	if c.slot == nil {
		return nil, errSlotIsNil
	}

	in.RLock()
	defer in.RUnlock()

	orders := []*structs.Order{}
	for _, order := range in.db {
		if uint64(len(orders)) >= c.count {
			break
		}

		if compareOrderAndSlot(c.slot, order, c.orderType) {
			orders = append(orders, order)
		}
	}

	sort.Sort(structs.ByPrice(orders))
	return orders, nil
}

func compareOrderAndSlot(slot *structs.Slot, order *structs.Order, typ pb.OrderType) bool {
	if typ != pb.OrderType_ANY && typ != order.GetType() {
		return false
	}

	os, _ := structs.NewSlot(order.Unwrap().Slot)
	return slot.Compare(os, order.GetType())
}

func (in *inMemOrderStorage) GetOrderByID(id string) (*structs.Order, error) {
	in.RLock()
	defer in.RUnlock()

	ord, ok := in.db[id]
	if !ok {
		return nil, errOrderNotFound
	}

	return ord, nil
}

func (in *inMemOrderStorage) CreateOrder(o *structs.Order) (*structs.Order, error) {
	id := in.generateID()
	o.SetID(id)

	in.Lock()
	defer in.Unlock()

	in.db[id] = o
	return o, nil
}

func (in *inMemOrderStorage) DeleteOrder(id string) error {
	in.Lock()
	defer in.Unlock()

	_, ok := in.db[id]
	if !ok {
		return errOrderNotFound
	}

	delete(in.db, id)
	return nil
}

func NewInMemoryStorage() OrderStorage {
	return &inMemOrderStorage{
		db: make(map[string]*structs.Order),
	}
}

type Marketplace struct {
	db   OrderStorage
	addr string
}

func (m *Marketplace) GetOrders(_ context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	slot, err := structs.NewSlot(req.Slot)
	if err != nil {
		return nil, err
	}

	resultCount := req.GetCount()
	if resultCount == 0 {
		resultCount = defaultResultsCount
	}

	searchParams := &searchParams{
		slot:      slot,
		orderType: req.GetOrderType(),
		count:     resultCount,
	}

	orders, err := m.db.GetOrders(searchParams)
	if err != nil {
		return nil, err
	}

	innerOrders := []*pb.Order{}
	for _, o := range orders {
		innerOrders = append(innerOrders, o.Unwrap())
	}

	return &pb.GetOrdersReply{
		Orders: innerOrders,
	}, nil
}

func (m *Marketplace) GetOrderByID(_ context.Context, req *pb.ID) (*pb.Order, error) {
	order, err := m.db.GetOrderByID(req.Id)
	if err != nil {
		return nil, err
	}
	return order.Unwrap(), nil
}

func (m *Marketplace) CreateOrder(_ context.Context, req *pb.Order) (*pb.Order, error) {
	order, err := structs.NewOrder(req)
	if err != nil {
		return nil, err
	}

	order, err = m.db.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	return order.Unwrap(), nil
}

func (m *Marketplace) CancelOrder(_ context.Context, req *pb.Order) (*pb.Empty, error) {
	err := m.db.DeleteOrder(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (m *Marketplace) Serve() error {
	lis, err := net.Listen("tcp", m.addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterMarketServer(grpcServer, m)
	grpcServer.Serve(lis)
	return nil
}

func NewMarketplace(addr string) *Marketplace {
	return &Marketplace{
		addr: addr,
		db:   NewInMemoryStorage(),
	}
}
