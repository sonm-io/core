package node

import (
	"fmt"
	"io"
	"math/big"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain/tsc"
	"github.com/sonm-io/core/insonmnia/dealer"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type HandlerStatus uint8

func (h HandlerStatus) String() string {
	m := map[HandlerStatus]string{
		statusNew:            "New",
		statusSearching:      "Searching",
		statusProposing:      "Proposing",
		statusDealing:        "Dealing",
		statusWaitForApprove: "Waiting for approve",
		statusDone:           "Done",
		statusFailed:         "Failed",
	}

	s, ok := m[h]
	if !ok {
		return "Unknown"
	}
	return s
}

const (
	// todo: make configurable
	orderPollPeriod = 5 * time.Second

	statusNew HandlerStatus = iota
	statusSearching
	statusProposing
	statusDealing
	statusWaitForApprove
	statusDone
	statusFailed
)

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
	sync.Mutex
	id     string
	order  *pb.Order
	status HandlerStatus
	ts     time.Time
	ctx    context.Context
	cancel context.CancelFunc

	err    error
	dealID string
}

func newOrderHandler(ctx context.Context, o *pb.Order) (*orderHandler, error) {
	ctx, cancel := context.WithCancel(ctx)

	order, err := structs.NewOrder(o)
	if err != nil {
		return nil, err
	}

	t := &orderHandler{
		ctx:    ctx,
		cancel: cancel,
		ts:     time.Now(),
		id:     order.GetID(),
		order:  o,
	}

	return t, nil
}

// setError keeps error into handler struct and
// changes task status to "failed"
func (h *orderHandler) setError(err error) {
	h.Lock()
	defer h.Unlock()

	h.status = statusFailed
	h.err = err
}

func (h *orderHandler) setStatus(s HandlerStatus) {
	h.Lock()
	defer h.Unlock()

	h.status = s
}

func (h *orderHandler) getStatus() HandlerStatus {
	h.Lock()
	defer h.Unlock()

	return h.status
}

type marketAPI struct {
	remotes    *remoteOptions
	ctx        context.Context
	hubCreator hubClientCreator

	taskMux sync.Mutex
	tasks   map[string]*orderHandler
}

// resolveHubAddr resolving Hub IP addr from Hub's Eth address
// via Locator service
func (m *marketAPI) resolveHubAddr(ethAddr string) (string, error) {
	req := &pb.ResolveRequest{EthAddr: ethAddr}
	reply, err := m.remotes.locator.Resolve(m.ctx, req)
	if err != nil {
		return "", err
	}

	ip := reply.Endpoints[0]
	log.G(m.ctx).Info("hub ip resolved successful", zap.String("ip", ip))

	return ip, nil
}

func (m *marketAPI) makeHubClient(ethAddr string) (pb.HubClient, io.Closer, error) {
	hubIP, err := m.resolveHubAddr(ethAddr)
	if err != nil {
		log.G(m.ctx).Info("cannot resolve Hub IP", zap.Error(err), zap.String("addr", ethAddr))
		return nil, nil, err
	}

	hub, cc, err := m.hubCreator(hubIP)
	if err != nil {
		log.G(m.ctx).Info("cannot create Hub gRPC client", zap.Error(err))
		return nil, nil, err
	}

	log.G(m.ctx).Info("hub connection built", zap.String("hub_ip", hubIP))
	return hub, cc, nil
}

func (m *marketAPI) getHandler(id string) (*orderHandler, bool) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	t, ok := m.tasks[id]
	return t, ok
}

func (m *marketAPI) registerHandler(id string, t *orderHandler) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	m.tasks[id] = t
}

func (m *marketAPI) deregisterHandler(id string) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	delete(m.tasks, id)
}

func (m *marketAPI) countHandlers() int {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	return len(m.tasks)
}

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	return m.remotes.market.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	return m.remotes.market.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	if req.OrderType != pb.OrderType_BID {
		return nil, errors.New("can create only Orders with type BID")
	}

	if _, err := structs.NewOrder(req); err != nil {
		return nil, err
	}

	req.ByuerID = util.PubKeyToAddr(m.remotes.key.PublicKey).Hex()
	created, err := m.remotes.market.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Marketplace knows nothing about the required duration, we must bypass it by hand.
	// Looks awful, but nevermind, it feels like out timing system is broken by design.
	created.Slot.Duration = req.GetSlot().GetDuration()
	go m.startHandler(created)

	return created, nil
}

func (m *marketAPI) startHandler(ord *pb.Order) {
	handler, err := newOrderHandler(m.ctx, ord)
	if err != nil {
		// push failed handler too, because we need to show error
		failedHandler := &orderHandler{id: ord.GetId(), err: err, status: statusFailed}
		m.registerHandler(ord.Id, failedHandler)
		log.G(m.ctx).Info("cannot create handler for order",
			zap.Error(err), zap.String("orderID", ord.GetId()))
		return
	}

	m.registerHandler(handler.id, handler)

	// process order (search -> propose -> deal)
	if ok := m.executeOnceWithCancel(handler); ok {
		return
	}

	tk := time.NewTicker(orderPollPeriod)
	defer tk.Stop()

	for {
		select {
		// Cancel context to stop polling for orders.
		case <-handler.ctx.Done():
			log.G(handler.ctx).Info("order handler is cancelled", zap.String("order_id", handler.id))
			return
		case <-tk.C:
			if ok := m.executeOnceWithCancel(handler); ok {
				return
			}
		}
	}
}

func (m *marketAPI) loadBalanceAndAllowance() (*big.Int, *big.Int, error) {
	addr := util.PubKeyToAddr(m.remotes.key.PublicKey).Hex()
	balance, err := m.remotes.eth.BalanceOf(m.ctx, addr)
	if err != nil {
		return nil, nil, err
	}

	allowance, err := m.remotes.eth.AllowanceOf(m.ctx, addr, tsc.DealsAddress)
	if err != nil {
		return nil, nil, err
	}

	return balance, allowance, nil
}

func (m *marketAPI) proposeDeal(h *orderHandler, ask *pb.Order) (*pb.Order, pb.HubClient, io.Closer) {
	log.G(h.ctx).Debug("proposing deal to hub", zap.String("hubEth", ask.GetSupplierID()))
	h.setStatus(statusProposing)

	hubClient, cc, err := m.makeHubClient(ask.SupplierID)
	if err != nil {
		log.G(h.ctx).Info("cannot create hub client", zap.Error(err))
		return nil, nil, nil
	}

	dealRequest := &pb.DealRequest{
		AskId:    ask.GetId(),
		BidId:    h.order.GetId(),
		SpecHash: structs.CalculateSpecHash(h.order),
	}

	_, err = hubClient.ProposeDeal(h.ctx, dealRequest)
	if err != nil {
		log.G(h.ctx).Info("cannot propose deal to the Hub", zap.Error(err))
		return nil, nil, nil
	}

	// stop proposing orders, now need to create Eth deal
	log.G(h.ctx).Info("finish proposing deal",
		zap.String("ord.id", ask.Id),
		zap.String("sup.id", ask.SupplierID))

	return ask, hubClient, cc
}

func (m *marketAPI) executeOnceWithCancel(handler *orderHandler) bool {
	err := m.execute(handler)
	if err != nil {
		if err != dealer.ErrOrdersNotFound {
			handler.setError(err)
		}

		return false
	}

	if _, err := m.remotes.market.CancelOrder(m.ctx, handler.order); err != nil {
		log.G(handler.ctx).Warn("cannot cancel order on market",
			zap.String("order_id", handler.id),
			zap.Error(err))
	}

	return true
}

// execute searching for orders, iterate found orders and trying to propose deal
func (m *marketAPI) execute(handler *orderHandler) error {
	log.G(handler.ctx).Info("starting execute", zap.String("id", handler.id))

	balance, allowance, err := m.loadBalanceAndAllowance()
	if err != nil {
		log.G(handler.ctx).Warn("cannot load balance and allowance", zap.Error(err))
		return err
	}

	log.G(handler.ctx).Debug("balance and allowance loaded",
		zap.String("balance", balance.String()),
		zap.String("allowance", allowance.String()))

	filter, err := dealer.NewSearchFilter(handler.order.GetSlot(), pb.OrderType_ASK,
		balance, allowance, handler.order.GetSupplierID())
	if err != nil {
		return err
	}

	handler.setStatus(statusSearching)
	searcher := dealer.NewAskSearcher(m.remotes.market)
	orders, err := searcher.Search(handler.ctx, filter)
	if err != nil {
		log.G(m.ctx).Debug("no order found on market", zap.String("bidID", handler.order.GetId()))
		return err
	}

	matcher := dealer.NewAskSelector()
	askToPropose, err := matcher.Select(orders)
	if err != nil {
		log.G(m.ctx).Debug("no matching selected for dealing", zap.String("bidID", handler.order.GetId()))
		return err
	}

	askToDealWith, hubClient, cc := m.proposeDeal(handler, askToPropose)
	if askToDealWith == nil {
		log.G(m.ctx).Debug("no hub accept proposed deal", zap.String("bidID", handler.order.GetId()))
		return errors.New("no hub accept proposed deal")
	}
	defer cc.Close()

	d := dealer.NewDealer(m.remotes.key, hubClient, m.remotes.eth, m.remotes.dealCreateTimeout)
	dealID, err := d.Deal(m.ctx, handler.order, askToDealWith)
	if err != nil {
		return err
	}

	handler.Lock()
	handler.err = nil
	handler.dealID = dealID.String()
	handler.status = statusDone
	handler.Unlock()

	log.G(handler.ctx).Info("handler done",
		zap.String("order_id", handler.id),
		zap.String("deal_id", dealID.String()))

	return nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, order *pb.Order) (*pb.Empty, error) {
	order, err := m.GetOrderByID(ctx, &pb.ID{Id: order.Id})
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve order type")
	}

	if order.OrderType != pb.OrderType_BID {
		return nil, errors.New(
			"can only remove bids via Market API; please use Hub ask-plan API to manage asks")
	}

	repl, err := m.remotes.market.CancelOrder(ctx, order)
	if err == nil {
		handler, ok := m.getHandler(order.Id)
		if ok {
			handler.cancel()
			m.deregisterHandler(order.Id)
		} else {
			log.G(m.ctx).Info("no order handler found", zap.String("order_id", order.Id))
		}
	}

	return repl, err
}

func (m *marketAPI) TouchOrders(ctx context.Context, req *pb.TouchOrdersRequest) (*pb.Empty, error) {
	return m.remotes.market.TouchOrders(ctx, req)
}

func (m *marketAPI) GetProcessing(ctx context.Context, req *pb.Empty) (*pb.GetProcessingReply, error) {
	m.taskMux.Lock()
	defer m.taskMux.Unlock()

	reply := &pb.GetProcessingReply{
		Orders: make(map[string]*pb.GetProcessingReply_ProcessedOrder),
	}

	for id, task := range m.tasks {
		var extra string
		if task.err != nil {
			extra = fmt.Sprintf("error: %s", task.err.Error())
		} else if task.dealID != "" {
			extra = fmt.Sprintf("deal ID: %s", task.dealID)
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

// getMyOrders query Marketplace service for orders
// with type == BID and that placed with current eth address
func (m *marketAPI) getMyOrders() (*pb.GetOrdersReply, error) {
	req := &pb.GetOrdersRequest{
		Order: &pb.Order{
			ByuerID:   util.PubKeyToAddr(m.remotes.key.PublicKey).Hex(),
			OrderType: pb.OrderType_BID,
		},
	}

	return m.remotes.market.GetOrders(m.ctx, req)
}

// restartOrdersProcessing loads BIDs for current account
// and restarts background processing for that orders
func (m *marketAPI) restartOrdersProcessing() func() error {
	return func() error {
		orders, err := m.getMyOrders()
		if err != nil {
			return err
		}

		log.G(m.ctx).Info("restart order processing",
			zap.Int("order_count", len(orders.GetOrders())))

		for _, o := range orders.GetOrders() {
			go m.startHandler(o)
		}

		return nil
	}

}

func newMarketAPI(opts *remoteOptions) (pb.MarketServer, error) {
	return &marketAPI{
		remotes:    opts,
		ctx:        opts.ctx,
		tasks:      make(map[string]*orderHandler),
		hubCreator: opts.hubCreator,
	}, nil
}
