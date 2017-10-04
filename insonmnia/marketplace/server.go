package marketplace

import (
	"errors"
	"net"
	"sync"

	"github.com/pborman/uuid"
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
	GetOrders(slot *pb.Slot) ([]*pb.Order, error)
	GetOrderByID(id string) (*pb.Order, error)
	CreateOrder(order *pb.Order) (*pb.Order, error)
	DeleteOrder(id string) error
}

type Slot pb.Slot

func (one *Slot) compareSupplierRating(two *Slot) bool {
	return two.SupplierRating >= one.SupplierRating
}

func (one *Slot) compareTime(two *Slot) bool {
	startOK := one.StartTime.Seconds > two.StartTime.Seconds
	endOK := one.EndTime.Seconds < two.EndTime.Seconds

	return startOK && endOK
}

func (one *Slot) compareCpuCores(two *Slot) bool {
	return two.Resources.CpuCores >= one.Resources.CpuCores
}

func (one *Slot) compareRamBytes(two *Slot) bool {
	return two.Resources.RamBytes >= one.Resources.RamBytes
}

func (one *Slot) compareGpuCount(two *Slot) bool {
	return two.Resources.GpuCount >= one.Resources.GpuCount
}

func (one *Slot) compareStorage(two *Slot) bool {
	return two.Resources.Storage >= one.Resources.Storage
}

func (one *Slot) compareNetTrafficIn(two *Slot) bool {
	return two.Resources.NetTrafficIn >= one.Resources.NetTrafficIn
}

func (one *Slot) compareNetTrafficOut(two *Slot) bool {
	return two.Resources.NetTrafficOut >= one.Resources.NetTrafficOut
}

func (one *Slot) compareNetworkType(two *Slot) bool {
	return two.Resources.NetworkType >= one.Resources.NetworkType
}

func (one *Slot) Validate() error {
	if one.Resources == nil {
		return errResourcesIsNil
	}

	if one.StartTime == nil {
		return errStartTimeRequired
	}

	if one.EndTime == nil {
		return errEndTimeRequired
	}

	if one.StartTime.Seconds >= one.EndTime.Seconds {
		return errStartTimeAfterEnd
	}

	return nil
}

func (one *Slot) Compare(two *Slot) bool {
	return one.compareSupplierRating(two) &&
		one.compareTime(two) &&
		one.compareCpuCores(two) &&
		one.compareRamBytes(two) &&
		one.compareGpuCount(two) &&
		one.compareStorage(two) &&
		one.compareNetTrafficIn(two) &&
		one.compareNetTrafficOut(two) &&
		one.compareNetworkType(two)
}

type Order pb.Order

func (ord *Order) Validate() error {
	// todo: check that order struct has all required fields
	if ord.Slot == nil {
		return errSlotIsNil
	}

	s := Slot(*ord.Slot)
	err := s.Validate()
	if err != nil {
		return err
	}

	if ord.Price <= 0 {
		return errPriceIsZero
	}

	return nil
}

type inMemOrderStorage struct {
	sync.RWMutex
	db map[string]*pb.Order
}

func (in *inMemOrderStorage) generateID() string {
	return uuid.New()
}

func (in *inMemOrderStorage) GetOrders(s *pb.Slot) ([]*pb.Order, error) {
	if s == nil {
		return nil, errSlotIsNil
	}

	sl := Slot(*s)
	err := sl.Validate()
	if err != nil {
		return nil, err
	}

	in.RLock()
	defer in.RUnlock()

	orders := []*pb.Order{}
	for _, order := range in.db {
		orderSlot := Slot(*order.Slot)
		if !sl.Compare(&orderSlot) {
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
	if order == nil {
		return nil, errOrderIsNil
	}

	ord := Order(*order)

	err := ord.Validate()
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
