package node

import (
	"sync"
	"time"

	"crypto/ecdsa"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	errNoHandlerWithID    = errors.New("cannot get handler with ID")
	errCannotProposeOrder = errors.New("cannot propose order")
	errNoMatchingOrder    = errors.New("cannot find matching ASK order")
	errNotAnBidOrder      = errors.New("can create only Orders with type BID")
)

const (
	statusNew uint8 = iota
	statusSearching
	statusProposing
	statusDealing
	statusDone
	statusFailed

	orderPollPeriod = 5 * time.Second
)

var statusMap = map[uint8]string{
	statusNew:       "New",
	statusSearching: "Searching",
	statusProposing: "Proposing",
	statusDealing:   "Dealing",
	statusDone:      "Done",
	statusFailed:    "Failed",
}

func HandlerStatusString(status uint8) string {
	s, ok := statusMap[status]
	if !ok {
		return "Unknown"
	}
	return s
}

// orderHandler is wrapper over Order
// allows to keep order execution status
//
// Order handling flow
// 1. Create orderHandler from received BID Order struct
//     - use structHash to generate task id
//     - assign task status as "NEW"
//
// 2. Searching for matching Ask Order -> status = "Searching"
//
// 3. Loop over found orders, try to propose order to Hub -> status = "Proposing"
//
// 4. If Proposing OK then try to create Deal on Etherum
//
// 5. If propose is completed -> status = "Done"
//
// In any internal error -> status = "Failed"

type orderHandler struct {
	id      string
	order   *pb.Order
	status  uint8
	err     error
	ts      time.Time
	ctx     context.Context
	cancel  context.CancelFunc
	locator pb.LocatorClient
}

func newOrderHandler(ctx context.Context, loc string, o *pb.Order) (*orderHandler, error) {
	ctx, cancel := context.WithCancel(ctx)

	cc, err := util.MakeGrpcClient(loc, nil)
	if err != nil {
		log.G(ctx).Debug("cannot create locator client", zap.Error(err))
		return nil, err
	}

	t := &orderHandler{
		ctx:     ctx,
		cancel:  cancel,
		ts:      time.Now(),
		locator: pb.NewLocatorClient(cc),
	}

	order, err := structs.NewOrder(o)
	if err != nil {
		log.G(t.ctx).Debug("cannot convert order to inner order", zap.Error(err))
		t.setError(err)
		return t, err
	}

	t.id = order.GetID()
	t.order = o

	return t, nil
}

// setError keeps error into handler struct and
// changes task status to "failed"
func (h *orderHandler) setError(err error) {
	h.status = statusFailed
	h.err = err
}

// search searches for matching orders on Marketplace
func (h *orderHandler) search(m pb.MarketClient) ([]*pb.Order, error) {
	log.G(h.ctx).Debug("searching for orders")
	h.status = statusSearching

	req := &pb.GetOrdersRequest{
		Slot:      h.order.GetSlot(),
		OrderType: pb.OrderType_ASK,
		Count:     100,
	}

	reply, err := m.GetOrders(h.ctx, req)
	if err != nil {
		return nil, err
	}

	return reply.Orders, nil
}

// resolveHubAddr resolving Hub IP addr from Hub's Eth address
// via Locator service
func (h *orderHandler) resolveHubAddr(ethAddr string) (string, error) {
	log.G(h.ctx).Debug("resolving Hub IP ip", zap.String("eth_addr", ethAddr))
	req := &pb.ResolveRequest{EthAddr: ethAddr}
	reply, err := h.locator.Resolve(h.ctx, req)
	if err != nil {
		return "", err
	}

	ip := reply.IpAddr[0]
	log.G(h.ctx).Debug("hub ip resolved successful", zap.String("ip", ip))
	return ip, nil
}

// propose proposes createDeal to Hub
func (h *orderHandler) propose(askID, supID string) error {
	h.status = statusProposing

	hubIP, err := h.resolveHubAddr(supID)
	if err != nil {
		log.G(h.ctx).Debug("cannot resolve Hub IP", zap.Error(err))
		h.setError(err)
		return err
	}

	cc, err := util.MakeGrpcClient(hubIP, nil)
	if err != nil {
		log.G(h.ctx).Debug("cannot create Hub gRPC client", zap.Error(err))
		h.setError(err)
		return err
	}

	hub := pb.NewHubClient(cc)

	req := &pb.DealRequest{BidId: h.order.Id, AskId: askID, Order: h.order}
	_, err = hub.ProposeDeal(h.ctx, req)
	if err != nil {
		log.G(h.ctx).Debug("cannot propose createDeal to Hub", zap.Error(err))
		// return typed error
		return errCannotProposeOrder
	}

	return nil
}

// createDeal creates deal on Etherum blockchain
func (h *orderHandler) createDeal(askOrder *pb.Order) error {
	log.G(h.ctx).Debug("creating deal on Etherum")
	h.status = statusDealing
	return nil
}

type marketAPI struct {
	conf   Config
	key    *ecdsa.PrivateKey
	market pb.MarketClient
	ctx    context.Context

	taskMux sync.Mutex
	tasks   map[string]*orderHandler
}

func (m *marketAPI) getHandler(id string) (*orderHandler, bool) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	t, ok := m.tasks[id]
	return t, ok
}

func (m *marketAPI) createHandler(id string, t *orderHandler) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	m.tasks[id] = t
}

func (m *marketAPI) removeHandler(id string) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	delete(m.tasks, id)
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	log.G(m.ctx).Info("handling GetOrders request")
	return m.market.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	log.G(m.ctx).Info("handling GetOrderByID request", zap.String("id", req.Id))
	return m.market.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	log.G(m.ctx).Info("handling CreateOrder request")

	if req.OrderType != pb.OrderType_BID {
		return nil, errNotAnBidOrder
	}

	req.ByuerID = util.PubKeyToAddr(m.key.PublicKey)
	created, err := m.market.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	go m.startExecOrderHandler(m.ctx, created)

	return created, nil
}

func (m *marketAPI) startExecOrderHandler(ctx context.Context, ord *pb.Order) {
	log.G(ctx).Debug("starting ExecOrder")
	handler, err := newOrderHandler(ctx, m.conf.LocatorEndpoint(), ord)
	// push handler to a map after error checking
	// we need to store a failed tasks too
	m.createHandler(handler.id, handler)
	if err != nil {
		log.G(handler.ctx).Debug("cannot create new bg handler from order", zap.Error(err))
		return
	}

	// process order (search -> propose -> deal)
	err = m.orderLoop(handler)
	if err == nil {
		log.G(handler.ctx).Debug("order loop complete at n=1 iteration, exiting")
		return
	}

	// remove order from Market if deal was make
	defer func() {
		_, err = m.CancelOrder(ctx, ord)
		if err != nil {
			log.G(handler.ctx).Debug("cannot cancel order", zap.String("err", err.Error()))
		}
	}()

	tk := time.NewTicker(orderPollPeriod)

	for {
		select {
		// cancel context to stop polling for ordrs
		case <-handler.ctx.Done():
			log.G(handler.ctx).Debug("handler is cancelled")
			return
		// retrier for order polling
		case <-tk.C:
			err := m.orderLoop(handler)
			if err == nil {
				log.G(handler.ctx).Debug("order loop complete at n > 1 iteration, exiting")
				return
			}
		}
	}
}

// orderLoop searching for orders, iterate found orders and trying to propose deal
func (m *marketAPI) orderLoop(handler *orderHandler) error {
	log.G(handler.ctx).Info("starting orderLoop", zap.String("id", handler.id))

	orders, err := handler.search(m.market)
	if err != nil {
		log.G(handler.ctx).Debug("cannot get orders", zap.Error(err))
		handler.setError(err)
		return err
	}

	if len(orders) == 0 {
		log.G(handler.ctx).Debug("no matching ASK orders found")
		return errNoMatchingOrder
	} else {
		log.G(handler.ctx).Debug("found order", zap.Int("count", len(orders)))
	}

	orderToDeal := &pb.Order{}
	for _, ord := range orders {
		err = handler.propose(ord.Id, ord.SupplierID)
		if err != nil {
			if err == errCannotProposeOrder {
				log.G(handler.ctx).Debug("cannot propose order, trying next order")
				continue
			}
		} else {
			// stop proposing orders, now need to create Eth deal
			orderToDeal = ord
			break
		}
	}

	err = handler.createDeal(orderToDeal)
	if err != nil {
		log.G(handler.ctx).Debug("cannot create deal, failing handler")
		handler.setError(err)
		return err
	}

	handler.status = statusDone
	log.G(handler.ctx).Info("handler done", zap.String("id", handler.id))
	return nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	log.G(m.ctx).Info("handling CancelOrder request", zap.String("id", req.Id))

	err := m.removeOrderHandler(req.Id)
	if err != nil {
		log.G(m.ctx).Info("cannot remove order handler", zap.String("id", req.Id))
	}

	return m.market.CancelOrder(ctx, req)
}

func (m *marketAPI) GetProcessing(ctx context.Context, req *pb.Empty) (*pb.GetProcessingReply, error) {
	log.G(m.ctx).Info("handling GetProcessing request")

	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	reply := &pb.GetProcessingReply{
		Orders: make(map[string]*pb.GetProcessingReply_ProcessedOrder),
	}

	for id, task := range m.tasks {
		var extra string
		if task.err != nil {
			extra = task.err.Error()
		}

		reply.Orders[id] = &pb.GetProcessingReply_ProcessedOrder{
			Id:        id,
			Status:    uint32(task.status),
			Timestamp: &pb.Timestamp{Seconds: task.ts.Unix()},
			Extra:     extra,
		}
	}

	return reply, nil
}

func (m *marketAPI) removeOrderHandler(id string) error {
	handlr, ok := m.getHandler(id)
	if !ok {
		return errNoHandlerWithID
	}

	handlr.cancel()
	m.removeHandler(id)

	return nil
}

func newMarketAPI(ctx context.Context, conf Config) (pb.MarketServer, error) {
	cc, err := util.MakeGrpcClient(conf.MarketEndpoint(), nil)
	if err != nil {
		return nil, err
	}

	return &marketAPI{
		conf:   conf,
		ctx:    ctx,
		market: pb.NewMarketClient(cc),
		tasks:  make(map[string]*orderHandler),
	}, nil
}
