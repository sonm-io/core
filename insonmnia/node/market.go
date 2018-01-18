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
	errCannotProposeOrder  = errors.New("cannot propose order")
	errNoMatchingOrder     = errors.New("cannot find matching ASK order")
	errNotAnBidOrder       = errors.New("can create only Orders with type BID")
	errProposeNotAccepted  = errors.New("no one hub accept proposed deal")
	errLackOfBalance       = errors.New("lack of balance or allowance for order")
	errHubEndpointIsNotSet = errors.New("hub endpoint is not configured, please check Node settings")
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
	sync.Mutex
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
	h.Lock()
	defer h.Unlock()

	h.status = statusFailed
	h.err = err
}

func (h *orderHandler) setStatus(s uint8) {
	h.Lock()
	defer h.Unlock()

	h.status = s
}

func (h *orderHandler) getStatus() uint8 {
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
		h.setError(err)
		return nil, nil, err
	}

	hub, cc, err := h.hubCreator(hubIP)
	if err != nil {
		log.G(h.ctx).Info("cannot create Hub gRPC client", zap.Error(err))
		h.setError(err)
		return nil, nil, err
	}

	log.G(h.ctx).Info("hub connection built", zap.String("hub_ip", hubIP))
	return hub, cc, nil
}

// propose proposes new deal to the Hub
func (h *orderHandler) propose(req *pb.DealRequest, hubClient pb.HubClient) error {
	h.setStatus(statusProposing)

	_, err := hubClient.ProposeDeal(h.ctx, req)
	if err != nil {
		log.G(h.ctx).Info("cannot propose openDeal to Hub", zap.Error(err))
		return errCannotProposeOrder
	}

	log.G(h.ctx).Info("order proposed successfully")
	return nil
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

	txID, err := h.bc.OpenDealPending(h.ctx, key, deal, wait)
	if err != nil {
		log.G(h.ctx).Info("cannot open deal", zap.Error(err))
		h.setError(err)
		return nil, err
	}

	log.G(h.ctx).Info("deal opened", zap.String("tx_id", txID.String()))
	return txID, nil
}

// approveOnHub send deal to the Hub and wait for approval on Hub-side
func (h *orderHandler) approveOnHub(req *pb.ApproveDealRequest, hub pb.HubClient) error {
	log.G(h.ctx).Info("waiting for deal become approved")
	h.setStatus(statusWaitForApprove)

	_, err := hub.ApproveDeal(h.ctx, req)
	if err != nil {
		h.setError(fmt.Errorf("cannot approve deal: %v", err))
		return h.err
	}

	return nil
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

	// process order (search -> propose -> deal)
	err = m.orderLoop(handler)
	if err == nil {
		log.G(handler.ctx).Info("order loop complete at n=1 iteration, exiting")
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
			err := m.orderLoop(handler)
			if err == nil {
				log.G(handler.ctx).Info("order loop complete at n > 1 iteration, exiting")

				_, err := m.remotes.market.CancelOrder(m.ctx, ord)
				if err != nil {
					log.G(handler.ctx).Info("cannot cancel order", zap.String("err", err.Error()),
						zap.String("order_id", handler.id))
				}

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

	var (
		orderToDeal *pb.Order
		hubClient   pb.HubClient
		cc          io.Closer
		dealRequest *pb.DealRequest
	)
	for _, ord := range orders {
		price := structs.CalculateTotalPrice(ord)
		if !m.checkBalanceAndAllowance(price, balance, allowance) {
			log.G(handler.ctx).Info("lack of balance or allowance for order", zap.String("order_id", ord.Id))
			handler.setError(errLackOfBalance)
			return errLackOfBalance
		}

		hubClient, cc, err = handler.makeHubClient(ord.SupplierID)
		if err != nil {
			log.G(handler.ctx).Info("cannot create hub client", zap.Error(err))
			handler.setError(err)
			continue
		}

		dealRequest = &pb.DealRequest{
			AskId:    ord.GetId(),
			BidId:    handler.order.GetId(),
			SpecHash: handler.slotSpecHash(),
		}

		err = handler.propose(dealRequest, hubClient)
		if err != nil {
			if err == errCannotProposeOrder {
				log.G(handler.ctx).Info("cannot propose order, trying next order")
				cc.Close()
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

	defer cc.Close()

	dealID, err := handler.openDeal(orderToDeal, m.remotes.key, m.remotes.dealCreateTimeout)
	if err != nil {
		log.G(handler.ctx).Info("cannot create deal", zap.Error(err))
		handler.setError(err)
		return err
	}

	approveRequest := &pb.ApproveDealRequest{
		DealID: pb.NewBigInt(dealID),
		AskID:  orderToDeal.GetId(),
		BidID:  handler.order.GetId(),
	}

	err = handler.approveOnHub(approveRequest, hubClient)
	if err != nil {
		log.G(handler.ctx).Info("wailed waiting for deal", zap.Error(err))
		handler.setError(err)
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
			handler.setError(errors.New("cancelled by user"))
			handler.cancel()
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
