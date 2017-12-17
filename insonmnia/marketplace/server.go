package marketplace

import (
	"crypto/ecdsa"
	"crypto/tls"
	"net"
	"sort"
	"sync"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultResultsCount = 100
)

var (
	errOrderNotFound         = errors.New("order cannot be found")
	errPriceIsZero           = errors.New("order price cannot be less or equal than zero")
	errOrderIsNil            = errors.New("order cannot be nil")
	errSlotIsNil             = errors.New("order slot cannot be nil")
	errResourcesIsNil        = errors.New("slot resources cannot be nil")
	errSearchParamsIsNil     = errors.New("search params cannot be nil")
	errNotEnoughSearchParams = errors.New("not enough search params")
)

// searchParams holds all fields that are used to search on the market.
// Extend this structure instead of increasing amount of params accepted
// by OrderStorage.GetOrders() function.
type searchParams struct {
	order *pb.Order
	count uint64
}

type OrderStorage interface {
	GetOrders(c *searchParams) ([]*structs.Order, error)
	GetOrderByID(id string) (*structs.Order, error)
	CreateOrder(order *structs.Order) (*structs.Order, error)
	DeleteOrder(id string) error
}

type inMemOrderStorage struct {
	sync.RWMutex
	ctx context.Context
	db  map[string]*structs.Order
}

func (in *inMemOrderStorage) generateID() string {
	return uuid.New()
}

func (in *inMemOrderStorage) GetOrders(c *searchParams) ([]*structs.Order, error) {
	if c == nil {
		return nil, errSearchParamsIsNil
	}

	if c.order == nil {
		return nil, errOrderIsNil
	}

	in.RLock()
	defer in.RUnlock()

	var orders []*structs.Order
	for _, order := range in.db {
		if uint64(len(orders)) >= c.count {
			break
		}

		var isMatch = false
		if c.order.Slot != nil {
			slot, _ := structs.NewSlot(c.order.Slot)
			isMatch = compareOrderAndSlot(slot, order, c.order.GetOrderType())
		}

		if c.order.GetSupplierID() != "" || c.order.GetByuerID() != "" {
			supplierMatch := order.Unwrap().GetSupplierID() == c.order.GetSupplierID() && order.Unwrap().GetSupplierID() != ""
			buyerMatch := order.Unwrap().GetByuerID() == c.order.GetByuerID() && order.Unwrap().ByuerID != ""

			isMatch = buyerMatch || supplierMatch
		}

		if isMatch {
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
	ctx         context.Context
	db          OrderStorage
	addr        string
	grpc        *grpc.Server
	certRotator util.HitlessCertRotator
	creds       credentials.TransportCredentials
}

func (m *Marketplace) GetOrders(_ context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	log.G(m.ctx).Info("handling GetOrders request", zap.Any("req", req))

	if req.Order.GetSlot() == nil && req.Order.GetSupplierID() == "" && req.Order.GetByuerID() == "" {
		return nil, errNotEnoughSearchParams
	}

	resultCount := req.GetCount()
	if resultCount == 0 {
		resultCount = defaultResultsCount
	}

	searchParams := &searchParams{
		order: req.GetOrder(),
		count: resultCount,
	}

	orders, err := m.db.GetOrders(searchParams)
	if err != nil {
		return nil, err
	}

	var innerOrders []*pb.Order
	for _, o := range orders {
		innerOrders = append(innerOrders, o.Unwrap())
	}

	return &pb.GetOrdersReply{
		Orders: innerOrders,
	}, nil
}

func (m *Marketplace) GetOrderByID(_ context.Context, req *pb.ID) (*pb.Order, error) {
	log.G(m.ctx).Info("handling GetOrderByID request", zap.Any("req", req))

	order, err := m.db.GetOrderByID(req.Id)
	if err != nil {
		return nil, err
	}
	return order.Unwrap(), nil
}

func (m *Marketplace) CreateOrder(_ context.Context, req *pb.Order) (*pb.Order, error) {
	log.G(m.ctx).Info("handling CreateOrder request", zap.Any("req", req))

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
	log.G(m.ctx).Info("handling CancelOrder request", zap.Any("req", req))

	err := m.db.DeleteOrder(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (m *Marketplace) GetProcessing(ctx context.Context, req *pb.Empty) (*pb.GetProcessingReply, error) {
	// This method exists just to match the Marketplace interface.
	// The Market service itself is unable to know anything about processing orders.
	// This method is implemented for Node in `insonmnia/node/market.go:348`
	return nil, nil
}

func (m *Marketplace) Serve() error {
	lis, err := net.Listen("tcp", m.addr)
	if err != nil {
		return err
	}

	m.grpc.Serve(lis)

	return nil
}

func NewMarketplace(ctx context.Context, cfg *MarketplaceConfig, key *ecdsa.PrivateKey) (m *Marketplace, err error) {
	if key == nil {
		return nil, errors.New("private key should be provided")
	}

	m = &Marketplace{
		ctx:  ctx,
		addr: cfg.ListenAddr,
		db:   NewInMemoryStorage(),
	}

	var TLSConfig *tls.Config
	m.certRotator, TLSConfig, err = util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	m.creds = util.NewTLS(TLSConfig)
	srv := util.MakeGrpcServer(m.creds)
	pb.RegisterMarketServer(srv, m)
	m.grpc = srv

	return m, nil
}
