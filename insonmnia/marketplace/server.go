package marketplace

import (
	"errors"
	"sync"

	"fmt"
	"github.com/pborman/uuid"
	pb "github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
)

var (
	errOrderNotFound  = errors.New("Order cannot be found")
	errPriceIsZero    = errors.New("Order price cannot be less or equal than zero")
	errOrderIsNil     = errors.New("Order cannot be nil")
	errSlotIsNil      = errors.New("Order slot cannot be nil")
	errResourcesIsNil = errors.New("Slot resources cannot be nil")
)

type OrderStorage interface {
	GetOrders(slot *pb.Slot) ([]*pb.Order, error)
	GetOrderByID(id string) (*pb.Order, error)
	CreateOrder(order *pb.Order) (*pb.Order, error)
	DeleteOrder(id string) error
}

type inMemOrderStorage struct {
	sync.RWMutex
	db map[string]*pb.Order
}

func (in *inMemOrderStorage) generateID() string {
	return uuid.New()
}

func (in *inMemOrderStorage) validateOrder(order *pb.Order) error {
	// todo: check that order struct has all required fields
	if order == nil {
		return errOrderIsNil
	}

	err := in.validateSlot(order.Slot)
	if err != nil {
		return err
	}

	if order.Price <= 0 {
		return errPriceIsZero
	}

	return nil
}

func (in *inMemOrderStorage) validateSlot(slot *pb.Slot) error {
	if slot == nil {
		return errSlotIsNil
	}

	if slot.Resources == nil {
		return errResourcesIsNil
	}
	return nil
}

func (in *inMemOrderStorage) compareTime(slot *pb.Slot, order *pb.Order) bool {
	startOK := slot.StartTime.Seconds > order.Slot.StartTime.Seconds
	endOK := slot.EndTime.Seconds < order.Slot.EndTime.Seconds

	return startOK && endOK
}

func (in *inMemOrderStorage) compareSupplierRating(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.SupplierRating >= slot.SupplierRating
}

func (in *inMemOrderStorage) compareCpuCores(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.CpuCores >= slot.Resources.CpuCores
}

func (in *inMemOrderStorage) compareRamBytes(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.RamBytes >= slot.Resources.RamBytes
}

func (in *inMemOrderStorage) compareGpuCount(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.GpuCount >= slot.Resources.GpuCount
}

func (in *inMemOrderStorage) compareStorage(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.Storage >= slot.Resources.Storage
}

func (in *inMemOrderStorage) compareNetTrafficIn(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.NetTrafficIn >= slot.Resources.NetTrafficIn
}

func (in *inMemOrderStorage) compareNetTrafficOut(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.NetTrafficOut >= slot.Resources.NetTrafficOut
}

func (in *inMemOrderStorage) compareNetType(slot *pb.Slot, order *pb.Order) bool {
	return order.Slot.Resources.NetworkType >= slot.Resources.NetworkType
}

func (in *inMemOrderStorage) GetOrders(slot *pb.Slot) ([]*pb.Order, error) {
	err := in.validateSlot(slot)
	if err != nil {
		return nil, err
	}

	in.RLock()
	defer in.RUnlock()

	orders := []*pb.Order{}
	for _, order := range in.db {
		if !in.compareSupplierRating(slot, order) {
			continue
		}

		if !in.compareTime(slot, order) {
			continue
		}

		if !in.compareCpuCores(slot, order) {
			continue
		}

		if !in.compareRamBytes(slot, order) {
			continue
		}

		if !in.compareGpuCount(slot, order) {
			continue
		}

		if !in.compareStorage(slot, order) {
			continue
		}

		if !in.compareNetTrafficIn(slot, order) {
			continue
		}

		if !in.compareNetTrafficOut(slot, order) {
			continue
		}

		if !in.compareNetType(slot, order) {
			continue
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (in *inMemOrderStorage) GetOrderByID(id string) (*pb.Order, error) {
	in.RLock()
	defer in.RUnlock()

	ord, ok := in.db[id]
	if !ok {
		return nil, errOrderNotFound
	}

	return ord, nil
}

func (in *inMemOrderStorage) CreateOrder(order *pb.Order) (*pb.Order, error) {
	err := in.validateOrder(order)
	if err != nil {
		return nil, err
	}

	order.Id = in.generateID()

	in.Lock()
	defer in.Unlock()

	in.db[order.Id] = order
	return order, nil
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
		db: make(map[string]*pb.Order),
	}
}

type Marketplace struct {
	db   OrderStorage
	addr string
}

func (m *Marketplace) GetOrders(_ context.Context, req *pb.Slot) (*pb.GetOrdersReply, error) {
	orders, err := m.db.GetOrders(req)
	if err != nil {
		return nil, err
	}
	return &pb.GetOrdersReply{
		Orders: orders,
	}, nil
}

func (m *Marketplace) GetOrderByID(_ context.Context, req *pb.GetOrderRequest) (*pb.Order, error) {
	order, err := m.db.GetOrderByID(req.Id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (m *Marketplace) CreateOrder(_ context.Context, req *pb.Order) (*pb.Order, error) {
	order, err := m.db.CreateOrder(req)
	if err != nil {
		return nil, err
	}

	return order, nil
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
