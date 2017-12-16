package node

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/blockchain/tsc"
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
	errProposeNotAccepted = errors.New("no one hub accept proposed deal")
)

const (
	statusNew uint8 = iota
	statusSearching
	statusProposing
	statusDealing
	statusWaitForApprove
	statusDone
	statusFailed

	orderPollPeriod = 5 * time.Second
)

var statusMap = map[uint8]string{
	statusNew:            "New",
	statusSearching:      "Searching",
	statusProposing:      "Proposing",
	statusDealing:        "Dealing",
	statusWaitForApprove: "Waiting for approve",
	statusDone:           "Done",
	statusFailed:         "Failed",
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
	id     string
	order  *pb.Order
	status uint8
	ts     time.Time
	ctx    context.Context
	cancel context.CancelFunc

	err    error
	dealID string

	locator    pb.LocatorClient
	bc         blockchain.Blockchainer
	hubCreator hubClientCreator
}

func newOrderHandler(ctx context.Context, loc pb.LocatorClient, bc blockchain.Blockchainer, hu hubClientCreator, o *pb.Order) (*orderHandler, error) {
	ctx, cancel := context.WithCancel(ctx)

	order, err := structs.NewOrder(o)
	if err != nil {
		return nil, err
	}

	t := &orderHandler{
		ctx:        ctx,
		cancel:     cancel,
		ts:         time.Now(),
		locator:    loc,
		bc:         bc,
		id:         order.GetID(),
		order:      o,
		hubCreator: hu,
	}

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
	log.G(h.ctx).Info("searching for orders")
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
	log.G(h.ctx).Info("resolving Hub IP ip", zap.String("eth_addr", ethAddr))
	req := &pb.ResolveRequest{EthAddr: ethAddr}
	reply, err := h.locator.Resolve(h.ctx, req)
	if err != nil {
		return "", err
	}

	ip := reply.IpAddr[0]
	log.G(h.ctx).Info("hub ip resolved successful", zap.String("ip", ip))
	return ip, nil
}

// propose proposes createDeal to Hub
func (h *orderHandler) propose(askID, supID string) error {
	h.status = statusProposing

	hubIP, err := h.resolveHubAddr(supID)
	if err != nil {
		log.G(h.ctx).Info("cannot resolve Hub IP", zap.Error(err))
		h.setError(err)
		return err
	}

	hub, err := h.hubCreator(hubIP)
	if err != nil {
		log.G(h.ctx).Info("cannot create Hub gRPC client", zap.Error(err))
		h.setError(err)
		return err
	}

	// TODO(sshaman1101): spec hash required here
	req := &pb.DealRequest{BidId: h.order.Id, AskId: askID, Order: h.order, SpecHash: "0"}
	_, err = hub.ProposeDeal(h.ctx, req)
	if err != nil {
		log.G(h.ctx).Info("cannot propose createDeal to Hub", zap.Error(err))
		return errCannotProposeOrder
	}

	log.G(h.ctx).Info("order proposed successfully", zap.String("hub_ip", hubIP))
	return nil
}

// createDeal creates deal on Etherum blockchain
func (h *orderHandler) createDeal(order *pb.Order, key *ecdsa.PrivateKey) error {
	log.G(h.ctx).Info("creating deal on Etherum")
	h.status = statusDealing

	deal := &pb.Deal{
		SupplierID: order.GetSupplierID(),
		BuyerID:    util.PubKeyToAddr(key.PublicKey).Hex(),
		Price:      order.Price,
		Status:     pb.DealStatus_PENDING,
		// TODO(sshaman1101): calculate hash
		SpecificationHash: "0",
	}

	tx, err := h.bc.OpenDeal(key, deal)
	if err != nil {
		log.G(h.ctx).Info("cannot open deal", zap.Error(err))
		h.setError(err)
		return err
	}

	log.G(h.ctx).Info("deal created", zap.String("tx_id", tx.Hash().String()))
	return nil
}

func (h *orderHandler) waitForApprove(order *pb.Order, key *ecdsa.PrivateKey, wait time.Duration) (*pb.Deal, error) {
	log.G(h.ctx).Info("waiting for deal become approved")
	h.status = statusWaitForApprove

	localCtx := context.Background()
	// TODO(sshaman1101): calculate hash
	deal, err := h.findDeals(localCtx, key, order.GetSupplierID(), "0", wait)
	if err != nil {
		log.G(h.ctx).Info("cannot find accepted deal", zap.Error(err))
		h.setError(err)
		return nil, err
	}

	if deal == nil {
		log.G(h.ctx).Info("deal was not accepted, fail by timeout")
		err = errors.New("deal was not accepted")
		h.setError(err)
		return nil, err
	}

	h.dealID = deal.Id
	log.G(h.ctx).Info("deal approved, ready to allocate task", zap.String("deal_id", deal.Id))
	return deal, nil
}

func (h *orderHandler) findDeals(ctx context.Context, key *ecdsa.PrivateKey, addr, hash string, wait time.Duration) (*pb.Deal, error) {
	ctx, cancel := context.WithTimeout(h.ctx, wait)
	defer cancel()

	tk := time.NewTicker(3 * time.Second)
	defer tk.Stop()

	if deal := h.findDealOnce(key, addr, hash); deal != nil {
		return deal, nil
	}

	for {
		select {
		case <-tk.C:
			if deal := h.findDealOnce(key, addr, hash); deal != nil {
				return deal, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (h *orderHandler) findDealOnce(key *ecdsa.PrivateKey, addr, hash string) *pb.Deal {
	// get deals opened by our client
	IDs, err := h.bc.GetAcceptedDeal(util.PubKeyToAddr(key.PublicKey).Hex(), addr)
	if err != nil {
		return nil
	}

	for _, id := range IDs {
		// then get extended info
		deal, err := h.bc.GetDealInfo(id)
		if err != nil {
			continue
		}

		// then check for status
		// and check if task hash is equal with request's one
		if deal.GetStatus() == pb.DealStatus_ACCEPTED && deal.GetSpecificationHash() == hash {
			return deal
		}
	}

	return nil
}

type marketAPI struct {
	remotes *remoteOptions
	ctx     context.Context
	taskMux sync.Mutex
	tasks   map[string]*orderHandler
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

func (m *marketAPI) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersReply, error) {
	log.G(m.ctx).Info("handling GetOrders request")
	return m.remotes.market.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	log.G(m.ctx).Info("handling GetOrderByID request", zap.String("id", req.Id))
	return m.remotes.market.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	log.G(m.ctx).Info("handling CreateOrder request")

	if req.OrderType != pb.OrderType_BID {
		return nil, errNotAnBidOrder
	}

	req.ByuerID = util.PubKeyToAddr(m.remotes.key.PublicKey).Hex()
	created, err := m.remotes.market.CreateOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	go m.startExecOrderHandler(created)

	return created, nil
}

func (m *marketAPI) startExecOrderHandler(ord *pb.Order) {
	log.G(m.ctx).Info("starting ExecOrder")

	handler, err := newOrderHandler(m.ctx, m.remotes.locator, m.remotes.eth, m.remotes.hubCreator, ord)
	if err != nil {
		// push failed handler too, because we need to show error
		failedHandler := &orderHandler{id: ord.GetId(), err: err, status: statusFailed}
		m.registerHandler(ord.Id, failedHandler)
		log.G(m.ctx).Info("cannot create new bg handler from order", zap.Error(err))
		return
	}

	m.registerHandler(handler.id, handler)

	// remove order from Market if deal was make
	defer func() {
		err := m.removeOrderHandler(handler.id)
		if err != nil {
			log.G(m.ctx).Info("cannot remove order handler", zap.String("handler_id", handler.id))
		}

		_, err = m.CancelOrder(m.ctx, ord)
		if err != nil {
			log.G(handler.ctx).Info("cannot cancel order", zap.String("err", err.Error()))
		}
	}()

	// process order (search -> propose -> deal)
	err = m.orderLoop(handler)
	if err == nil {
		log.G(handler.ctx).Info("order loop complete at n=1 iteration, exiting")
		return
	}

	tk := time.NewTicker(orderPollPeriod)

	for {
		select {
		// cancel context to stop polling for ordrs
		case <-handler.ctx.Done():
			log.G(handler.ctx).Info("handler is cancelled")
			return
			// retrier for order polling
		case <-tk.C:
			err := m.orderLoop(handler)
			if err == nil {
				log.G(handler.ctx).Info("order loop complete at n > 1 iteration, exiting")
				return
			}
		}
	}
}

func (m *marketAPI) loadBalanceAndAllowance() (*big.Int, *big.Int, error) {
	addr := util.PubKeyToAddr(m.remotes.key.PublicKey).Hex()
	balance, err := m.remotes.eth.BalanceOf(addr)
	if err != nil {
		return nil, nil, err
	}
	allowance, err := m.remotes.eth.AllowanceOf(addr, tsc.DealsAddress)
	if err != nil {
		return nil, nil, err
	}
	return balance, allowance, nil
}

func (m *marketAPI) checkBalanceAndAllowance(price, balance, allowance *big.Int) bool {
	if balance.Cmp(price) == -1 && allowance.Cmp(price) == -1 {
		return false
	}
	return true
}

// orderLoop searching for orders, iterate found orders and trying to propose deal
func (m *marketAPI) orderLoop(handler *orderHandler) error {
	log.G(handler.ctx).Info("starting orderLoop", zap.String("id", handler.id))

	orders, err := handler.search(m.remotes.market)
	if err != nil {
		log.G(handler.ctx).Info("cannot get orders", zap.Error(err))
		handler.setError(err)
		return err
	}

	if len(orders) == 0 {
		log.G(handler.ctx).Info("no matching ASK orders found")
		return errNoMatchingOrder
	} else {
		log.G(handler.ctx).Info("found order", zap.Int("count", len(orders)))
	}

	balance, allowance, err := m.loadBalanceAndAllowance()
	if err != nil {
		log.G(handler.ctx).Error("cannot get orders", zap.Error(err))
		handler.setError(err)
		return err
	}

	var orderToDeal *pb.Order = nil
	for _, ord := range orders {
		price, err := util.ParseBigInt(ord.Price)
		if !m.checkBalanceAndAllowance(price, balance, allowance) {
			log.G(handler.ctx).Info("lack of balance and allowance for order", zap.String("order_id", ord.Id))
			continue
		}

		err = handler.propose(ord.Id, ord.SupplierID)
		if err != nil {
			if err == errCannotProposeOrder {
				log.G(handler.ctx).Info("cannot propose order, trying next order")
				continue
			}
		} else {
			// stop proposing orders, now need to create Eth deal
			log.G(handler.ctx).Info("finish proposing deal",
				zap.String("ord.id", ord.Id),
				zap.String("sup.id", ord.SupplierID))
			orderToDeal = ord
			break
		}
	}

	if orderToDeal == nil {
		// order still nil - proposeDeal failed for each order for each hub
		log.G(handler.ctx).Info("no one hub accept proposed deal")
		handler.setError(errProposeNotAccepted)
		return err
	}

	err = handler.createDeal(orderToDeal, m.remotes.key)
	if err != nil {
		log.G(handler.ctx).Info("cannot create deal", zap.Error(err))
		handler.setError(err)
		return err
	}

	deal, err := handler.waitForApprove(orderToDeal, m.remotes.key, m.remotes.approveTimeout)
	if err != nil {
		log.G(handler.ctx).Info("wailed waiting for deal", zap.Error(err))
		handler.setError(err)
		return err
	}

	handler.status = statusDone
	log.G(handler.ctx).Info("handler done",
		zap.String("handle_id", handler.id),
		zap.String("deal_id", deal.GetId()))
	return nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	log.G(m.ctx).Info("handling CancelOrder request", zap.String("id", req.Id))
	return m.remotes.market.CancelOrder(ctx, req)
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

// removeOrderHandler must cancel context and DO NOT REMOVE the handler from tasks map
func (m *marketAPI) removeOrderHandler(id string) error {
	handlr, ok := m.getHandler(id)
	if !ok {
		return errNoHandlerWithID
	}

	handlr.cancel()
	return nil
}

func newMarketAPI(opts *remoteOptions) (pb.MarketServer, error) {
	return &marketAPI{
		remotes: opts,
		ctx:     opts.ctx,
		tasks:   make(map[string]*orderHandler),
	}, nil
}
