package marketplace

import (
	"errors"
	"net"
	"sync"

	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
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
	GetOrders(slot *structs.Slot) ([]*structs.Order, error)
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

func (in *inMemOrderStorage) GetOrders(s *structs.Slot) ([]*structs.Order, error) {
	if s == nil {
		return nil, errSlotIsNil
	}

	in.RLock()
	defer in.RUnlock()

	orders := []*structs.Order{}
	for _, order := range in.db {
		os, _ := structs.NewSlot(order.Unwrap().Slot)
		if !s.Compare(os) {
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
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
		innerOrders = append(innerOrders, o.Unwrap())
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
