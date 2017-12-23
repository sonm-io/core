package node

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/cnf/structhash"
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
	errNoHandlerWithID     = errors.New("cannot get handler with ID")
	errCannotProposeOrder  = errors.New("cannot propose order")
	errNoMatchingOrder     = errors.New("cannot find matching ASK order")
	errNotAnBidOrder       = errors.New("can create only Orders with type BID")
	errProposeNotAccepted  = errors.New("no one hub accept proposed deal")
	errLackOfBalance       = errors.New("lack of balance or allowance for order")
	errHubEndpointIsNotSet = errors.New("hub endpoint is not configured, please check Node settings")
)

const (
	statusNew            uint8 = iota
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
		Order: &pb.Order{
			SupplierID: h.order.SupplierID,
			OrderType:  pb.OrderType_ASK,
			Slot:       h.order.GetSlot(),
		},
		Count: 100,
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

	req := &pb.DealRequest{
		BidId:    h.order.Id,
		AskId:    askID,
		Order:    h.order,
		SpecHash: h.slotSpecHash(),
	}

	_, err = hub.ProposeDeal(h.ctx, req)
	if err != nil {
		log.G(h.ctx).Info("cannot propose createDeal to Hub", zap.Error(err))
		return errCannotProposeOrder
	}

	log.G(h.ctx).Info("order proposed successfully", zap.String("hub_ip", hubIP))
	return nil
}

// createDeal creates deal on Ethereum blockchain
func (h *orderHandler) createDeal(order *pb.Order, key *ecdsa.PrivateKey, wait time.Duration) (*big.Int, error) {
	log.G(h.ctx).Info("creating deal on Etherum")
	h.status = statusDealing

	bigPricePerHour, err := util.ParseBigInt(order.Price)
	if err != nil {
		h.setError(err)
		return nil, err
	}
	bigPricePerSecond := bigPricePerHour.Div(bigPricePerHour, big.NewInt(3600))
	bigDuration := big.NewInt(int64(order.Slot.Duration))
	price := big.NewInt(0).Mul(bigPricePerSecond, bigDuration).String()

	deal := &pb.Deal{
		WorkTime:          h.order.GetSlot().GetDuration(),
		SupplierID:        order.GetSupplierID(),
		BuyerID:           util.PubKeyToAddr(key.PublicKey).Hex(),
		Price:             price,
		Status:            pb.DealStatus_PENDING,
		SpecificationHash: h.slotSpecHash(),
	}

	// tx, err := h.bc.OpenDeal(key, deal)
	txID, err := h.bc.OpenDealPending(h.ctx, key, deal, wait)
	if err != nil {
		log.G(h.ctx).Info("cannot open deal", zap.Error(err))
		h.setError(err)
		return nil, err
	}

	log.G(h.ctx).Info("deal opened", zap.String("tx_id", txID.String()))
	return txID, nil
}

func (h *orderHandler) waitForApprove(key *ecdsa.PrivateKey, dealID *big.Int, wait time.Duration) (*pb.Deal, error) {
	log.G(h.ctx).Info("waiting for deal become approved")
	h.status = statusWaitForApprove

	deal, err := h.pollForApprovedStatus(key, dealID, wait)
	if err != nil {
		log.G(h.ctx).Info("cannot find accepted deal", zap.Error(err))
		h.setError(err)
		return nil, err
	}

	if deal == nil {
		h.setError(errors.New("deal was not accepted"))
		log.G(h.ctx).Info("deal was not accepted, fail by timeout")
		return nil, h.err
	}

	log.G(h.ctx).Info("deal approved, ready to allocate task", zap.String("deal_id", deal.Id))
	return deal, nil
}

func (h *orderHandler) pollForApprovedStatus(key *ecdsa.PrivateKey, dealID *big.Int, wait time.Duration) (*pb.Deal, error) {
	ctx, cancel := context.WithTimeout(h.ctx, wait)
	defer cancel()

	tk := time.NewTicker(3 * time.Second)
	defer tk.Stop()

	if deal := h.findDealOnce(key, dealID); deal != nil {
		return deal, nil
	}

	for {
		select {
		case <-tk.C:
			if deal := h.findDealOnce(key, dealID); deal != nil {
				return deal, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (h *orderHandler) findDealOnce(key *ecdsa.PrivateKey, dealID *big.Int) *pb.Deal {
	// get deals opened by our client
	deal, err := h.bc.GetDealInfo(dealID)
	if err == nil && deal.Status == pb.DealStatus_ACCEPTED {
		return deal
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
	return m.remotes.market.GetOrders(ctx, req)
}

func (m *marketAPI) GetOrderByID(ctx context.Context, req *pb.ID) (*pb.Order, error) {
	return m.remotes.market.GetOrderByID(ctx, req)
}

func (m *marketAPI) CreateOrder(ctx context.Context, req *pb.Order) (*pb.Order, error) {
	if req.OrderType != pb.OrderType_BID {
		return nil, errNotAnBidOrder
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
	if balance.Cmp(price) == -1 || allowance.Cmp(price) == -1 {
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
		log.G(handler.ctx).Error("cannot load balance and allowance", zap.Error(err))
		handler.setError(err)
		return err
	}

	var orderToDeal *pb.Order = nil
	for _, ord := range orders {
		price, err := util.ParseBigInt(ord.Price)
		if !m.checkBalanceAndAllowance(price, balance, allowance) {
			log.G(handler.ctx).Info("lack of balance or allowance for order", zap.String("order_id", ord.Id))
			handler.setError(errLackOfBalance)
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
		return errProposeNotAccepted
	}

	dealID, err := handler.createDeal(orderToDeal, m.remotes.key, m.remotes.dealCreateTimeout)
	if err != nil {
		log.G(handler.ctx).Info("cannot create deal", zap.Error(err))
		handler.setError(err)
		return err
	}

	deal, err := handler.waitForApprove(m.remotes.key, dealID, m.remotes.dealApproveTimeout)
	if err != nil {
		log.G(handler.ctx).Info("wailed waiting for deal", zap.Error(err))
		handler.setError(err)
		return err
	}

	handler.dealID = deal.Id
	handler.status = statusDone
	log.G(handler.ctx).Info("handler done",
		zap.String("handle_id", handler.id),
		zap.String("deal_id", deal.GetId()))
	return nil
}

func (m *marketAPI) CancelOrder(ctx context.Context, req *pb.Order) (*pb.Empty, error) {
	return m.remotes.market.CancelOrder(ctx, req)
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

// removeOrderHandler must cancel context and DO NOT REMOVE the handler from tasks map
func (m *marketAPI) removeOrderHandler(id string) error {
	handlr, ok := m.getHandler(id)
	if !ok {
		return errNoHandlerWithID
	}

	handlr.cancel()
	return nil
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
			go m.startExecOrderHandler(o)
		}

		return nil
	}

}

func newMarketAPI(opts *remoteOptions) (pb.MarketServer, error) {
	return &marketAPI{
		remotes: opts,
		ctx:     opts.ctx,
		tasks:   make(map[string]*orderHandler),
	}, nil
}

// slotSpecHash hashes handler's order and convert hash to big.Int
func (h *orderHandler) slotSpecHash() string {
	s := structhash.Md5(h.order.GetSlot().GetResources(), 1)
	result := big.NewInt(0).SetBytes(s).String()
	log.G(h.ctx).Debug("slot hash calculated", zap.String("value", result))
	return result
}
