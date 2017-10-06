package marketplace

import (
	"errors"
	"net"
	"sync"

	"github.com/pborman/uuid"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"github.com/sonm-io/core/insonmnia/structs"
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
)

type OrderStorage interface {
	GetOrders(slot *structs.Slot) ([]*Order, error)
	GetOrderByID(id string) (*Order, error)
	CreateOrder(order *Order) (*Order, error)
	DeleteOrder(id string) error
}

type Order struct {
	inner *pb.Order
}

func (o *Order) Unwrap() *pb.Order {
	return o.inner
}

func NewOrder(o *pb.Order) (*Order, error) {
	if err := validateOrder(o); err != nil {
		return nil, err
	} else {
		return &Order{inner: o}, nil
	}
}

func validateOrder(o *pb.Order) error {
	if o == nil {
		return errOrderIsNil
	}

	if o.Price <= 0 {
		return errPriceIsZero
	}

	_, err := structs.NewSlot(o.Slot)
	if err != nil {
		return err
	}

	return nil
}

type inMemOrderStorage struct {
	sync.RWMutex
	db map[string]*Order
}

func (in *inMemOrderStorage) generateID() string {
	return uuid.New()
}

func (in *inMemOrderStorage) GetOrders(s *structs.Slot) ([]*Order, error) {
	if s == nil {
		return nil, errSlotIsNil
	}

	in.RLock()
	defer in.RUnlock()

	orders := []*Order{}
	for _, order := range in.db {
		os, _ := structs.NewSlot(order.inner.Slot)
		if !s.Compare(os) {
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (in *inMemOrderStorage) GetOrderByID(id string) (*Order, error) {
	in.RLock()
	defer in.RUnlock()

	ord, ok := in.db[id]
	if !ok {
		return nil, errOrderNotFound
	}

	return ord, nil
}

func (in *inMemOrderStorage) CreateOrder(o *Order) (*Order, error) {
	id := in.generateID()
	o.inner.Id = id

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
		db: make(map[string]*Order),
	}
}

type Marketplace struct {
	db   OrderStorage
	addr string
}

func (m *Marketplace) GetOrders(_ context.Context, req *pb.Slot) (*pb.GetOrdersReply, error) {
	slot, err := structs.NewSlot(req)
	if err != nil {
		return nil, err
	}

	orders, err := m.db.GetOrders(slot)
	if err != nil {
		return nil, err
	}

	innerOrders := []*pb.Order{}
	for _, o := range orders {
		innerOrders = append(innerOrders, o.inner)
	}

	return &pb.GetOrdersReply{
		Orders: innerOrders,
	}, nil
}

func (m *Marketplace) GetOrderByID(_ context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	order, err := m.db.GetOrderByID(req.Id)
	if err != nil {
		return nil, err
	}
	return order.inner, nil
}

func (m *Marketplace) CreateOrder(_ context.Context, req *pb.Order) (*pb.Order, error) {
	order, err := NewOrder(req)
	if err != nil {
		return nil, err
	}

	order, err = m.db.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	return order.inner, nil
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
