package node

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLogger(logging.NewNullGRPCLogger())
}

var (
	key  = getTestKey()
	addr = util.PubKeyToAddr(key.PublicKey)
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := ethcrypto.GenerateKey()
	return k
}

func makeSlot() *pb.Slot {
	return &pb.Slot{
		Duration:  uint64(structs.MinSlotDuration.Seconds()),
		Resources: &pb.Resources{},
	}
}

func makeOrder() *pb.Order {
	return &pb.Order{
		Id:             "qwe",
		PricePerSecond: pb.NewBigIntFromInt(100),
		OrderType:      pb.OrderType_BID,
		Slot:           makeSlot(),
	}
}

func getTestEth(ctrl *gomock.Controller) blockchain.Blockchainer {
	deal := &pb.Deal{
		Id:                "1",
		Status:            pb.DealStatus_ACCEPTED,
		SpecificationHash: "217643283185136810854905094570012887099",
	}

	bc := blockchain.NewMockBlockchainer(ctrl)

	bc.EXPECT().BalanceOf(gomock.Any()).AnyTimes().
		Return(big.NewInt(big.MaxPrec), nil)
	bc.EXPECT().AllowanceOf(gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(big.MaxPrec), nil)
	bc.EXPECT().OpenDeal(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&types.Transaction{}, nil)
	bc.EXPECT().GetAcceptedDeal(gomock.Any(), gomock.Any()).AnyTimes().
		Return([]*big.Int{big.NewInt(1)}, nil)
	bc.EXPECT().GetDealInfo(big.NewInt(1)).AnyTimes().
		Return(deal, nil)
	bc.EXPECT().OpenDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(1), nil)

	return bc
}

func getTestMarket(ctrl *gomock.Controller) pb.MarketClient {
	m := pb.NewMockMarketClient(ctrl)
	ord := makeOrder()
	ord.ByuerID = addr.Hex()
	ord.Id = "my-order-id"

	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(ord, nil)
	m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.GetOrdersReply{Orders: []*pb.Order{ord}}, nil)
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.Empty{}, nil)
	return m
}

func getTestLocator(ctrl *gomock.Controller) pb.LocatorClient {
	loc := pb.NewMockLocatorClient(ctrl)
	loc.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.ResolveReply{Endpoints: []string{"127.0.0.1:10001"}}, nil)
	return loc
}

func getTestConfig(ctrl *gomock.Controller) Config {
	cfg := NewMockConfig(ctrl)
	cfg.EXPECT().LocatorEndpoint().AnyTimes().Return("127.0.0.1:9090")
	cfg.EXPECT().MarketEndpoint().AnyTimes().Return("127.0.0.1:9095")
	return cfg
}

func getTestHubClient(ctrl *gomock.Controller) (pb.HubClient, io.Closer) {
	hub := NewMockHubClient(ctrl)
	hub.EXPECT().ProposeDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.Empty{}, nil)
	hub.EXPECT().ApproveDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.Empty{}, nil)

	return hub, &mockConn{}
}

func getTestRemotes(ctx context.Context, ctrl *gomock.Controller) *remoteOptions {
	key := getTestKey()
	conf := getTestConfig(ctrl)

	opts, err := newRemoteOptions(ctx, key, conf, nil)
	if err != nil {
		panic(err)
	}

	opts.eth = getTestEth(ctrl)
	opts.market = getTestMarket(ctrl)
	opts.locator = getTestLocator(ctrl)
	opts.dealApproveTimeout = 3 * time.Second
	opts.hubCreator = func(addr string) (pb.HubClient, io.Closer, error) {
		hub, cc := getTestHubClient(ctrl)
		return hub, cc, nil
	}

	return opts
}

func TestCreateOrder_FullAsyncOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)
	assert.True(t, inner.countHandlers() == 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusDone, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusDone], statusMap[handlr.getStatus()]))
	assert.Equal(t, "1", handlr.dealID)
}

func TestCreateOrder_CannotCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create custom market mock that can fail
	m := pb.NewMockMarketClient(ctrl)
	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.Order{Id: "some-broken-order", Slot: &pb.Slot{}}, nil)
	m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot get orders"))
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot cancel order"))

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)
	opts.market = m

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err, "order must be created on remote market")

	time.Sleep(50 * time.Millisecond)
	assert.True(t, inner.countHandlers() == 1)

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err)
}

func TestCreateOrder_CannotFetchOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create custom market mock that can fail
	m := pb.NewMockMarketClient(ctrl)
	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(makeOrder(), nil)
	m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot get orders"))
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot cancel order"))

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)
	opts.market = m

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)

	created, err := inner.CreateOrder(ctx, makeOrder())
	require.NoError(t, err, "order must be created on remote market")

	time.Sleep(50 * time.Millisecond)
	assert.True(t, inner.countHandlers() == 1)

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err)
	assert.EqualError(t, handlr.err, "TEST: cannot get orders")
}

func TestCreateOrder_CannotNoMatchingOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create custom market mock that can fail
	m := pb.NewMockMarketClient(ctrl)
	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(makeOrder(), nil)
	m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.GetOrdersReply{Orders: []*pb.Order{}}, nil)
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot cancel order"))

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)
	opts.market = m

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)

	created, err := inner.CreateOrder(ctx, makeOrder())
	require.NoError(t, err, "order must be created on remote market")

	time.Sleep(50 * time.Millisecond)
	assert.True(t, inner.countHandlers() == 1)

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusSearching, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusSearching], statusMap[handlr.getStatus()]))
	require.NoError(t, handlr.err)
}

func TestCreateOrder_CannotResolveHubIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	loc := pb.NewMockLocatorClient(ctrl)
	loc.EXPECT().Resolve(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot resolve hub ip"))

	opts := getTestRemotes(ctx, ctrl)
	opts.locator = loc

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(50 * time.Millisecond)

	assert.True(t, inner.countHandlers() == 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err)
	assert.EqualError(t, handlr.err, errProposeNotAccepted.Error())
}

func TestCreateOrder_CannotCreateDeal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	eth := blockchain.NewMockBlockchainer(ctrl)
	eth.EXPECT().BalanceOf(gomock.Any()).AnyTimes().
		Return(big.NewInt(big.MaxPrec), nil)
	eth.EXPECT().AllowanceOf(gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(big.MaxPrec), nil)
	eth.EXPECT().OpenDeal(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot open deal"))
	eth.EXPECT().OpenDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot open deal"))
	eth.EXPECT().GetAcceptedDeal(gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("TEST: cannot get accepted deals"))
	eth.EXPECT().GetDealInfo(big.NewInt(1)).AnyTimes().
		Return(nil, errors.New("TEST: cannot get deal info"))

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)
	opts.eth = eth

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)

	assert.True(t, inner.countHandlers() == 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err)
	assert.EqualError(t, handlr.err, "TEST: cannot open deal")
}

func TestCreateOrder_CannotWaitForApprove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	opts := getTestRemotes(ctx, ctrl)
	opts.hubCreator = func(addr string) (pb.HubClient, io.Closer, error) {
		hub := NewMockHubClient(ctrl)
		hub.EXPECT().ProposeDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.Empty{}, nil)
		hub.EXPECT().ApproveDeal(gomock.Any(), gomock.Any()).AnyTimes().Return(
			nil, errors.New("TEST: cannot approve deal"))
		return hub, &mockConn{}, nil
	}

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)
	assert.True(t, inner.countHandlers() == 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err)
}

func TestCreateOrder_LackAllowanceBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	eth := blockchain.NewMockBlockchainer(ctrl)
	eth.EXPECT().BalanceOf(gomock.Any()).AnyTimes().
		Return(big.NewInt(100), nil)
	eth.EXPECT().AllowanceOf(gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(50), nil)

	opts := getTestRemotes(ctx, ctrl)
	opts.eth = eth

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)

	assert.Equal(t, inner.countHandlers(), 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err, errProposeNotAccepted)
}

func TestCreateOrder_LackAllowance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	eth := blockchain.NewMockBlockchainer(ctrl)
	eth.EXPECT().BalanceOf(gomock.Any()).AnyTimes().
		Return(big.NewInt(10000), nil)
	eth.EXPECT().AllowanceOf(gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(50), nil)

	opts := getTestRemotes(ctx, ctrl)
	opts.eth = eth

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)

	assert.Equal(t, inner.countHandlers(), 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err, errProposeNotAccepted)
}

func TestCreateOrder_LackBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	eth := blockchain.NewMockBlockchainer(ctrl)
	eth.EXPECT().BalanceOf(gomock.Any()).AnyTimes().
		Return(big.NewInt(100), nil)
	eth.EXPECT().AllowanceOf(gomock.Any(), gomock.Any()).AnyTimes().
		Return(big.NewInt(50000), nil)
	eth.EXPECT().OpenDealPending(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().
		Return(nil, errors.New("1"))

	opts := getTestRemotes(ctx, ctrl)
	opts.eth = eth

	server, err := newMarketAPI(opts)
	require.NoError(t, err)

	inner := server.(*marketAPI)
	created, err := inner.CreateOrder(ctx, makeOrder())

	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotEmpty(t, created.Id)

	// wait for async handler is finished
	time.Sleep(1 * time.Second)
	assert.Equal(t, inner.countHandlers(), 1, "Handler must not be removed")

	handlr, ok := inner.getHandler(created.Id)
	require.True(t, ok)

	assert.Equal(t, statusFailed, handlr.getStatus(),
		fmt.Sprintf("Wait for status %s, but has %s", statusMap[statusFailed], statusMap[handlr.getStatus()]))
	assert.Error(t, handlr.err, errProposeNotAccepted)
}

type mockConn struct{}

func (c *mockConn) Close() error { return nil }
