package node

import (
	"crypto/ecdsa"
	"fmt"
	"io"
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
	errProposeNotAccepted = errors.New("no hub accept proposed deal")
	errLackOfBalance      = errors.New("lack of balance or allowance for order")
	errNoAskFound         = errors.New("cannot find matching ASK order")
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

// search searches for matching orders on Marketplace
func (h *orderHandler) search(m pb.MarketClient) ([]*pb.Order, error) {
	log.G(h.ctx).Info("searching for orders")
	h.setStatus(statusSearching)

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

	if len(reply.GetOrders()) == 0 {
		return nil, errNoAskFound
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

	ip := reply.Endpoints[0]
	log.G(h.ctx).Info("hub ip resolved successful", zap.String("ip", ip))

	return ip, nil
}

func (h *orderHandler) makeHubClient(ethAddr string) (pb.HubClient, io.Closer, error) {
	hubIP, err := h.resolveHubAddr(ethAddr)
	if err != nil {
		log.G(h.ctx).Info("cannot resolve Hub IP", zap.Error(err))
		return nil, nil, err
	}

	hub, cc, err := h.hubCreator(hubIP)
	if err != nil {
		log.G(h.ctx).Info("cannot create Hub gRPC client", zap.Error(err))
		return nil, nil, err
	}

	log.G(h.ctx).Info("hub connection built", zap.String("hub_ip", hubIP))
	return hub, cc, nil
}

// openDeal creates deal on Ethereum blockchain
func (h *orderHandler) openDeal(order *pb.Order, key *ecdsa.PrivateKey, wait time.Duration) (*big.Int, error) {
	log.G(h.ctx).Info("creating deal on Etherum")
	h.setStatus(statusDealing)

	deal := &pb.Deal{
		WorkTime:          h.order.GetSlot().GetDuration(),
		SupplierID:        order.GetSupplierID(),
		BuyerID:           util.PubKeyToAddr(key.PublicKey).Hex(),
		Price:             pb.NewBigInt(structs.CalculateTotalPrice(h.order)),
		Status:            pb.DealStatus_PENDING,
		SpecificationHash: h.slotSpecHash(),
	}

	dealID, err := h.bc.OpenDealPending(h.ctx, key, deal, wait)
	if err != nil {
		log.G(h.ctx).Info("cannot open deal", zap.Error(err))
		return nil, err
	}

	log.G(h.ctx).Info("deal opened", zap.String("deal_id", dealID.String()))
	return dealID, nil
}

// approveOnHub send deal to the Hub and wait for approval on Hub-side
func (h *orderHandler) approveOnHub(req *pb.ApproveDealRequest, hub pb.HubClient) error {
	log.G(h.ctx).Info("waiting for deal become approved")
	h.setStatus(statusWaitForApprove)

	_, err := hub.ApproveDeal(h.ctx, req)
	if err != nil {
		return err
	}

	log.G(h.ctx).Info("deal approved")
	return nil
}

func (m *marketAPI) closeUnapprovedDeal(dealID *big.Int) error {
	err := m.remotes.eth.CloseDealPending(m.ctx, m.remotes.key, dealID, time.Duration(180*time.Second))
	if err != nil {
		return err
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

	// process order (search -> propose -> deal)
	if ok := m.executeOrderOnceWithCancel(handler); ok {
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
			if ok := m.executeOrderOnceWithCancel(handler); ok {
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

func (m *marketAPI) checkBalanceAndAllowance(price, balance, allowance *big.Int) bool {
	if balance.Cmp(price) == -1 || allowance.Cmp(price) == -1 {
		return false
	}

	return true
}

func (m *marketAPI) filterOrdersByPriceAndAllowance(ctx context.Context, balance, allowance *big.Int, orders []*pb.Order) ([]*pb.Order, error) {
	var matched []*pb.Order

	for _, ord := range orders {
		price := structs.CalculateTotalPrice(ord)
		if !m.checkBalanceAndAllowance(price, balance, allowance) {
			log.G(ctx).Info("lack of balance or allowance for order, skip",
				zap.String("orderID", ord.Id),
				zap.String("price", price.String()),
				zap.String("balance", balance.String()),
				zap.String("allowance", allowance.String()))
			continue
		}

		matched = append(matched, ord)
	}

	if len(matched) == 0 {
		return nil, errLackOfBalance
	}

	return matched, nil
}

func (m *marketAPI) proposeDeal(h *orderHandler, ord *pb.Order) (*pb.Order, pb.HubClient, io.Closer) {
	h.setStatus(statusProposing)

	hubClient, cc, err := h.makeHubClient(ord.SupplierID)
	if err != nil {
		log.G(h.ctx).Info("cannot create hub client", zap.Error(err))
		return nil, nil, nil
	}

	dealRequest := &pb.DealRequest{
		AskId:    ord.GetId(),
		BidId:    h.order.GetId(),
		SpecHash: h.slotSpecHash(),
	}

	_, err = hubClient.ProposeDeal(h.ctx, dealRequest)
	if err != nil {
		log.G(h.ctx).Info("cannot propose deal to the Hub", zap.Error(err))
		return nil, nil, nil
	}

	// stop proposing orders, now need to create Eth deal
	log.G(h.ctx).Info("finish proposing deal",
		zap.String("ord.id", ord.Id),
		zap.String("sup.id", ord.SupplierID))

	return ord, hubClient, cc
}

func (m *marketAPI) executeOrderOnceWithCancel(handler *orderHandler) bool {
	err := m.executeOrder(handler)

	if err != nil {
		if err != errNoAskFound {
			handler.setError(err)
		}

		return false
	}

	log.G(handler.ctx).Debug("order loop complete at n=1 iteration, exiting")

	if _, err := m.remotes.market.CancelOrder(m.ctx, handler.order); err != nil {
		log.G(handler.ctx).Warn("cannot cancel order on market",
			zap.String("order_id", handler.id),
			zap.Error(err))
	}

	return true
}

// executeOrder searching for orders, iterate found orders and trying to propose deal
func (m *marketAPI) executeOrder(handler *orderHandler) error {
	log.G(handler.ctx).Info("starting executeOrder", zap.String("id", handler.id))

	balance, allowance, err := m.loadBalanceAndAllowance()
	if err != nil {
		log.G(handler.ctx).Error("cannot load balance and allowance", zap.Error(err))
		return err
	}

	orders, err := handler.search(m.remotes.market)
	if err != nil {
		log.G(handler.ctx).Info("cannot get orders", zap.Error(err))
		return err
	}

	// iterate orders #1, check for balance,
	// filter orders if price is too high or allowance is too low
	ordersForProposeDeal, err := m.filterOrdersByPriceAndAllowance(handler.ctx, balance, allowance, orders)
	if err != nil {
		return err
	}

	var (
		orderToDeal *pb.Order
		hubClient   pb.HubClient
		cc          io.Closer
	)

	// iterate orders #2, try to propose order
	for _, ord := range ordersForProposeDeal {
		orderToDeal, hubClient, cc = m.proposeDeal(handler, ord)
		if orderToDeal != nil {
			break
		}
	}

	// order still nil - deal cannot be proposed, failing the handler
	if orderToDeal == nil {
		return errProposeNotAccepted
	}

	defer cc.Close()

	dealID, err := handler.openDeal(orderToDeal, m.remotes.key, m.remotes.dealCreateTimeout)
	if err != nil {
		return err
	}

	approveRequest := &pb.ApproveDealRequest{
		DealID: pb.NewBigInt(dealID),
		AskID:  orderToDeal.GetId(),
		BidID:  handler.order.GetId(),
	}

	err = handler.approveOnHub(approveRequest, hubClient)
	if err != nil {
		log.G(handler.ctx).Info("hub cannot approve deal, need to close deal", zap.Error(err))

		err = m.closeUnapprovedDeal(dealID)
		if err != nil {
			log.G(handler.ctx).Warn("cannot close unapproved deal", zap.Error(err))
			return err
		}

		log.G(handler.ctx).Info("unapproved deal closed")
		return errors.New("deal is not approved on the hub")
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
			extra = fmt.Sprintf("deal Name: %s", task.dealID)
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
	return result
}
