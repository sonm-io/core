package dwh

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	bch "github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
)

const (
	testDBPath        = "test_dwh.db"
	testMonitorDBPath = "test_monitor_dwh.db"
)

var (
	globalDWH  *DWH
	monitorDWH *DWH
)

func TestMain(m *testing.M) {
	var err error
	globalDWH, err = getTestDWH(testDBPath)
	if err != nil {
		fmt.Println(err)
		os.Remove(testDBPath)
		os.Remove(testMonitorDBPath)
		os.Exit(1)
	}

	monitorDWH, err = getTestDWH(testMonitorDBPath)
	if err != nil {
		fmt.Println(err)
		os.Remove(testMonitorDBPath)
		os.Exit(1)
	}

	retCode := m.Run()
	globalDWH.db.Close()
	os.Remove(testDBPath)
	os.Remove(testMonitorDBPath)
	os.Exit(retCode)
}

func getTestDWH(dbPath string) (*DWH, error) {
	var (
		ctx = context.Background()
		cfg = &Config{
			Storage: &storageConfig{
				Backend:  "sqlite3",
				Endpoint: dbPath,
			},
		}
	)

	db, err := sql.Open(cfg.Storage.Backend, cfg.Storage.Endpoint)
	if err != nil {
		return nil, err
	}

	w := &DWH{
		ctx:           ctx,
		cfg:           cfg,
		db:            db,
		logger:        log.GetLogger(ctx),
		numBenchmarks: 12,
	}

	return w, setupTestDB(w)
}

func TestDWH_GetDeals(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	// Test TEXT columns.
	{
		request := &pb.DealsRequest{
			Status:     pb.DealStatus_DEAL_UNKNOWN,
			SupplierID: pb.NewEthAddress(common.HexToAddress("0x15")),
		}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Deals) != 1 {
			t.Errorf("Expected 1 deal in reply, got %d", len(reply.Deals))
			return
		}

		if reply.Deals[0].GetDeal().SupplierID.Unwrap().Hex() != common.HexToAddress("0x15").Hex() {
			t.Errorf("Request `%+v` failed, expected %s, got %s (SupplierID)",
				request, common.HexToAddress("0x15").Hex(), reply.Deals[0].GetDeal().SupplierID)
			return
		}
	}
	// Test INTEGER columns.
	{
		request := &pb.DealsRequest{
			Status: pb.DealStatus_DEAL_UNKNOWN,
			Duration: &pb.MaxMinUint64{
				Min: 10015,
			},
		}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Deals) < 5 {
			t.Errorf("Expected 5 deals in reply, got %d", len(reply.Deals))
			return
		}
	}
	// Test TEXT columns which should be treated as INTEGERS.
	{
		request := &pb.DealsRequest{
			Status: pb.DealStatus_DEAL_UNKNOWN,
			Price: &pb.MaxMinBig{
				Min: pb.NewBigIntFromInt(20015),
			},
		}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Deals) != 5 {
			t.Errorf("Expected 5 deals in reply, got %d", len(reply.Deals))
			return
		}

		if reply.Deals[0].GetDeal().Price.Unwrap().String() != "20015" {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Price)",
				request, 10015, reply.Deals[0].GetDeal().Duration)
			return
		}
	}
}

func TestDWH_GetDealDetails(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	deal, err := globalDWH.storage.GetDealByID(newSimpleConn(globalDWH.db), big.NewInt(40405))
	if err != nil {
		t.Error(err)
		return
	}

	reply := deal.GetDeal()
	if reply.Id.Unwrap().String() != "40405" {
		t.Errorf("Expected %s, got %s (Id)", "40405", reply.Id.Unwrap().String())
	}
	if reply.SupplierID.Unwrap().Hex() != common.HexToAddress("0x15").Hex() {
		t.Errorf("Expected %s, got %s (SupplierID)", common.HexToAddress("0x15").Hex(), reply.SupplierID)
	}
	if reply.ConsumerID.Unwrap().Hex() != common.HexToAddress("0x25").Hex() {
		t.Errorf("Expected %s, got %s (ConsumerID)", common.HexToAddress("0x25").Hex(), reply.ConsumerID)
	}
	if reply.MasterID.Unwrap().Hex() != common.HexToAddress("0x35").Hex() {
		t.Errorf("Expected %s, got %s (MasterID)", common.HexToAddress("0x35").Hex(), reply.MasterID)
	}
	if reply.AskID.Unwrap().String() != "20205" {
		t.Errorf("Expected %s, got %s (AskID)", "20205", reply.AskID.Unwrap().String())
	}
	if reply.BidID.Unwrap().String() != "30305" {
		t.Errorf("Expected %s, got %s (BidID)", "30305", reply.AskID.Unwrap().String())
	}
	if reply.Duration != uint64(10015) {
		t.Errorf("Expected %d, got %d (Duration)", 10015, reply.Duration)
	}
	if reply.Price.Unwrap().String() != "20015" {
		t.Errorf("Expected %s, got %s (Price)", "20015", reply.Price.Unwrap().String())
	}
	if reply.StartTime.Seconds != 30015 {
		t.Errorf("Expected %d, got %d (SatrtTime)", 30015, reply.StartTime.Seconds)
	}
	if reply.EndTime.Seconds != 40015 {
		t.Errorf("Expected %d, got %d (EndTime)", 40015, reply.EndTime.Seconds)
	}
	if reply.BlockedBalance.Unwrap().String() != "50015" {
		t.Errorf("Expected %s, got %s (BlockedBalance)", "50015", reply.BlockedBalance.Unwrap().String())
	}
	if reply.TotalPayout.Unwrap().String() != "60015" {
		t.Errorf("Expected %s, got %s (TotalPayout)", "60015", reply.TotalPayout.Unwrap().String())
	}
	if reply.LastBillTS.Seconds != 70015 {
		t.Errorf("Expected %d, got %d (LastBillTS)", 70015, reply.LastBillTS.Seconds)
	}
	if !deal.ActiveChangeRequest {
		t.Errorf("Expected %t, got %t (ActiveChangeRequest)", true, deal.ActiveChangeRequest)
	}
}

func TestDWH_GetOrders(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	// Test TEXT columns.
	{
		request := &pb.OrdersRequest{
			Type:   pb.OrderType_ANY,
			DealID: pb.NewBigIntFromInt(10105),
		}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(orders) != 2 {
			t.Errorf("Expected 2 orders in reply, got %d", len(orders))
			return
		}

		if orders[0].GetOrder().AuthorID.Unwrap().Hex() != common.HexToAddress("0xA").Hex() {
			t.Errorf("Request `%+v` failed, expected %s, got %s (AuthorID)",
				request, common.HexToAddress("0xA").Hex(), orders[0].GetOrder().AuthorID)
			return
		}
	}
	// Test INTEGER columns.
	{
		request := &pb.OrdersRequest{
			Type: pb.OrderType_ASK,
			Duration: &pb.MaxMinUint64{
				Min: 10015,
			},
			Benchmarks: map[uint64]*pb.MaxMinUint64{0: {Min: 15}},
		}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(orders) != 5 {
			t.Errorf("Expected 5 orders in reply, got %d", len(orders))
			return
		}

		if orders[0].GetOrder().Duration != uint64(10015) {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Duration)",
				request, 10015, orders[0].GetOrder().Duration)
			return
		}
	}
	// Test TEXT columns which should be treated as INTEGERS.
	{
		request := &pb.OrdersRequest{
			Type: pb.OrderType_ASK,
			Price: &pb.MaxMinBig{
				Min: pb.NewBigIntFromInt(int64(20015)),
			},
		}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(orders) != 5 {
			t.Errorf("Expected 5 orders in reply, got %d", len(orders))
			return
		}

		if orders[0].GetOrder().Price.Unwrap().String() != "20015" {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Price)",
				request, 10015, orders[0].GetOrder().Duration)
			return
		}
	}
}

func TestDWH_GetMatchingOrders(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	request := &pb.MatchingOrdersRequest{
		Id: pb.NewBigIntFromInt(20205),
	}
	orders, _, err := globalDWH.storage.GetMatchingOrders(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("GetMatchingOrders failed: %s", err)
		return
	}

	if len(orders) != 5 {
		t.Errorf("Expected 5 orders in reply, got %d", len(orders))
		return
	}
}

func TestDWH_GetOrderDetails(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	order, err := globalDWH.storage.GetOrderByID(newSimpleConn(globalDWH.db), big.NewInt(20205))
	if err != nil {
		t.Error(err)
		return
	}

	reply := order.GetOrder()
	if reply.Id.Unwrap().String() != "20205" {
		t.Errorf("Expected %s, got %s (Id)", "20205", reply.Id.Unwrap().String())
	}
	if reply.DealID.Unwrap().String() != "10105" {
		t.Errorf("Expected %s, got %s (DealID)", "10105", reply.DealID)
	}
	if reply.OrderType != 2 {
		t.Errorf("Expected %d, got %d (Type)", 2, reply.OrderType)
	}
	if reply.AuthorID.Unwrap().Hex() != common.HexToAddress("0xA").Hex() {
		t.Errorf("Expected %s, got %s (AuthorID)", common.HexToAddress("0xA").Hex(), reply.AuthorID)
	}
	if reply.CounterpartyID.Unwrap().Hex() != common.HexToAddress("0xB").Hex() {
		t.Errorf("Expected %s, got %s (CounterpartyID)", common.HexToAddress("0xB").Hex(), reply.CounterpartyID)
	}
	if reply.Duration != uint64(10015) {
		t.Errorf("Expected %d, got %d (Duration)", 10015, reply.Duration)
	}
	if reply.Price.Unwrap().String() != "20015" {
		t.Errorf("Expected %s, got %s (Price)", "20015", reply.Price.Unwrap().String())
	}
	if reply.Netflags != 7 {
		t.Errorf("Expected %d, got %d (Netflags)", 7, reply.Netflags)
	}
	if reply.Blacklist != "blacklist_5" {
		t.Errorf("Expected %s, got %s (Blacklist)", "blacklist_5", reply.Blacklist)
	}
	if reply.FrozenSum.Unwrap().String() != "30015" {
		t.Errorf("Expected %s, got %s (FrozenSum)", "30015", reply.FrozenSum.Unwrap().String())
	}
}

func TestDWH_GetDealChangeRequests(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	changeRequests, err := globalDWH.getDealChangeRequests(newSimpleConn(globalDWH.db), pb.NewBigIntFromInt(40400))
	if err != nil {
		t.Error(err)
		return
	}

	if len(changeRequests) != 10 {
		t.Errorf("Expected %d DealChangeRequests, got %d", 10, len(changeRequests))
		return
	}
}

func TestDWH_GetProfiles(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	profiles, _, err := globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		Name: "sortedProfile",
		Sortings: []*pb.SortingOption{
			{
				Field: "UserID",
				Order: pb.SortingOrder_Asc,
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(profiles))
		return
	}

	if profiles[0].UserID.Unwrap().Hex() != common.HexToAddress("0x20").Hex() {
		t.Errorf("Expected %s, got %s (Profile.UserID)", common.HexToAddress("0x20").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}

	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		Name: "sortedProfile",
		Sortings: []*pb.SortingOption{
			{
				Field: "UserID",
				Order: pb.SortingOrder_Desc,
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(profiles))
		return
	}

	if profiles[0].UserID.Unwrap().Hex() != common.HexToAddress("0x29").Hex() {
		t.Errorf("Expected %s, got %s (Profile.UserID)", common.HexToAddress("0x29").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}

	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		Name: "sortedProfile",
		Sortings: []*pb.SortingOption{
			{
				Field: "IdentityLevel",
				Order: pb.SortingOrder_Asc,
			},
			{
				Field: "UserID",
				Order: pb.SortingOrder_Asc,
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(profiles))
		return
	}

	if profiles[0].UserID.Unwrap().Hex() != common.HexToAddress("0x20").Hex() {
		t.Errorf("Expected %s, got %s (Profile.UserID)", common.HexToAddress("0x20").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}
	if profiles[4].UserID.Unwrap().Hex() != common.HexToAddress("0x28").Hex() {
		t.Errorf("Expected %s, got %s (Profile.UserID)", common.HexToAddress("0x28").Hex(), profiles[4].UserID.Unwrap().Hex())
		return
	}

	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_OnlyMatching,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 1 {
		t.Errorf("Expected %d Profiles, got %d", 1, len(profiles))
		return
	}

	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_WithoutMatching,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 11 {
		t.Errorf("Expected %d Profiles, got %d", 11, len(profiles))
		return
	}

	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_IncludeAndMark,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(profiles) != 12 {
		t.Errorf("Expected %d Profiles, got %d", 12, len(profiles))
		return
	}

	var foundMarkedProfile bool
	for _, profile := range profiles {
		if profile.IsBlacklisted {
			foundMarkedProfile = true
		}
	}

	if !foundMarkedProfile {
		t.Error("failed to find profile marked as blacklisted")
	}
}

func TestDWH_monitor(t *testing.T) {
	var (
		controller           = gomock.NewController(t)
		mockBlock            = bch.NewMockAPI(controller)
		mockMarket           = bch.NewMockMarketAPI(controller)
		mockProfiles         = bch.NewMockProfileRegistryAPI(controller)
		commonID             = big.NewInt(0xDEADBEEF)
		commonEventTS uint64 = 5
	)
	benchmarks, err := pb.NewBenchmarks([]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	require.NoError(t, err)
	deal := &pb.Deal{
		Id:             pb.NewBigInt(commonID),
		Benchmarks:     benchmarks,
		SupplierID:     pb.NewEthAddress(common.HexToAddress("0xAA")),
		ConsumerID:     pb.NewEthAddress(common.HexToAddress("0xBB")),
		MasterID:       pb.NewEthAddress(common.HexToAddress("0xCC")),
		AskID:          pb.NewBigIntFromInt(20205),
		BidID:          pb.NewBigIntFromInt(30305),
		Duration:       10020,
		Price:          pb.NewBigInt(big.NewInt(20010)),
		StartTime:      &pb.Timestamp{Seconds: 30010},
		EndTime:        &pb.Timestamp{Seconds: 40010},
		Status:         pb.DealStatus_DEAL_ACCEPTED,
		BlockedBalance: pb.NewBigInt(big.NewInt(50010)),
		TotalPayout:    pb.NewBigInt(big.NewInt(0)),
		LastBillTS:     &pb.Timestamp{Seconds: int64(commonEventTS)},
	}
	mockMarket.EXPECT().GetDealInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(deal, nil)
	order := &pb.Order{
		Id:             pb.NewBigInt(commonID),
		DealID:         pb.NewBigIntFromInt(0),
		OrderType:      pb.OrderType_ASK,
		OrderStatus:    pb.OrderStatus_ORDER_ACTIVE,
		AuthorID:       pb.NewEthAddress(common.HexToAddress("0xD")),
		CounterpartyID: pb.NewEthAddress(common.HexToAddress("0x0")),
		Duration:       10020,
		Price:          pb.NewBigInt(big.NewInt(20010)),
		Netflags:       7,
		IdentityLevel:  pb.IdentityLevel_ANONYMOUS,
		Blacklist:      "blacklist",
		Tag:            []byte{0, 1},
		Benchmarks:     benchmarks,
		FrozenSum:      pb.NewBigInt(big.NewInt(30010)),
	}
	mockMarket.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(order, nil)
	changeRequest := &pb.DealChangeRequest{
		Id:          pb.NewBigIntFromInt(0),
		DealID:      pb.NewBigInt(commonID),
		RequestType: pb.OrderType_ASK,
		Duration:    10020,
		Price:       pb.NewBigInt(big.NewInt(20010)),
		Status:      pb.ChangeRequestStatus_REQUEST_CREATED,
	}
	mockMarket.EXPECT().GetDealChangeRequestInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(changeRequest, nil)
	mockMarket.EXPECT().GetNumBenchmarks(gomock.Any()).AnyTimes().Return(uint64(12), nil)
	validator := &pb.Validator{
		Id:    pb.NewEthAddress(common.HexToAddress("0xC")),
		Level: 3,
	}
	mockProfiles.EXPECT().GetValidator(gomock.Any(), gomock.Any()).AnyTimes().Return(validator, nil)
	certificate := &pb.Certificate{
		ValidatorID:   pb.NewEthAddress(common.HexToAddress("0xC")),
		OwnerID:       pb.NewEthAddress(common.HexToAddress("0xD")),
		Attribute:     CertificateName,
		IdentityLevel: 1,
		Value:         []byte("User Name"),
	}
	mockProfiles.EXPECT().GetCertificate(gomock.Any(), gomock.Any()).AnyTimes().Return(
		certificate, nil)
	mockBlock.EXPECT().Market().AnyTimes().Return(mockMarket)
	mockBlock.EXPECT().ProfileRegistry().AnyTimes().Return(mockProfiles)

	monitorDWH.blockchain = mockBlock

	if err := testOrderPlaced(commonEventTS, commonID); err != nil {
		t.Errorf("testOrderPlaced: %s", err)
		return
	}
	if err := testDealOpened(deal, commonID); err != nil {
		t.Errorf("testDealOpened: %s", err)
		return
	}
	if err := testValidatorCreatedUpdated(validator); err != nil {
		t.Errorf("testValidatorCreatedUpdated: %s", err)
		return
	}
	if err := testCertificateUpdated(certificate, commonID); err != nil {
		t.Errorf("testCertificateUpdated: %s", err)
		return
	}
	if err := testOrderUpdated(order, commonID); err != nil {
		t.Errorf("testOrderUpdated: %s", err)
		return
	}
	if err := testDealUpdated(deal, commonID); err != nil {
		t.Errorf("testDealUpdated: %s", err)
		return
	}
	if err := testDealChangeRequestSentAccepted(changeRequest, commonEventTS, commonID); err != nil {
		t.Errorf("testDealChangeRequestSentAccepted: %s", err)
		return
	}
	if err := testBilled(commonEventTS, commonID); err != nil {
		t.Errorf("testBilled: %s", err)
		return
	}
	if err := testDealClosed(deal, commonID); err != nil {
		t.Errorf("testDealClosed: %s", err)
		return
	}
	if err := testWorkerAnnouncedConfirmedRemoved(); err != nil {
		t.Errorf("testWorkerAnnouncedConfirmedRemoved: %s", err)
		return
	}
	if err := testBlacklistAddedRemoved(); err != nil {
		t.Errorf("testBlacklistAddedRemoved: %s", err)
		return
	}
}

func testOrderPlaced(commonEventTS uint64, commonID *big.Int) error {
	if err := monitorDWH.onOrderPlaced(commonEventTS, commonID); err != nil {
		return errors.Wrap(err, "onOrderPlaced failed")
	}
	if order, err := monitorDWH.storage.GetOrderByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return errors.Wrap(err, "storage.GetOrderByID failed")
	} else {
		if order.GetOrder().Duration != 10020 {
			return errors.Errorf("Expected %d, got %d (Order.Duration)", 10020, order.GetOrder().Duration)
		}
	}
	return nil
}

func testDealOpened(deal *pb.Deal, commonID *big.Int) error {
	if err := monitorDWH.onDealOpened(commonID); err != nil {
		return errors.Wrap(err, "onDealOpened failed")
	}
	// Firstly, check that a deal was created.
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return errors.Wrap(err, "storage.GetDealByID failed")
	} else {
		if deal.GetDeal().Duration != 10020 {
			errors.Errorf("Expected %d, got %d (Deal.Duration)", 10020, deal.GetDeal().Duration)
		}
	}
	// Secondly, check that a DealCondition was created.
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return errors.Wrap(err, "getDealConditions failed")
	} else {
		if dealConditions[0].Duration != 10020 {
			return errors.Errorf("Expected %d, got %d (DealCondition.Duration)", 10020, deal.Duration)
		}
	}
	return nil
}

func testValidatorCreatedUpdated(validator *pb.Validator) error {
	// Check that a Validator entry is added after ValidatorCreated event.
	if err := monitorDWH.onValidatorCreated(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return errors.Wrap(err, "onValidatorCreated failed")
	}
	if validators, _, err := monitorDWH.storage.GetValidators(newSimpleConn(monitorDWH.db), &pb.ValidatorsRequest{}); err != nil {
		return errors.Wrap(err, "getValidators failed")
	} else {
		if len(validators) != 1 {
			return errors.Errorf("(ValidatorCreated) Expected 1 Validator, got %d", len(validators))
		}
		if validators[0].Level != 3 {
			return errors.Errorf("(ValidatorCreated) Expected %d, got %d (Validator.Level)",
				3, validators[0].Level)
		}
	}
	validator.Level = 0
	// Check that a Validator entry is updated after ValidatorDeleted event.
	if err := monitorDWH.onValidatorDeleted(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return errors.Wrap(err, "onValidatorDeleted failed")
	}
	if validators, _, err := monitorDWH.storage.GetValidators(newSimpleConn(monitorDWH.db), &pb.ValidatorsRequest{}); err != nil {
		return errors.Wrap(err, "getValidators failed")
	} else {
		if len(validators) != 1 {
			return errors.Errorf("(ValidatorDeleted) Expected 1 Validator, got %d", len(validators))
		}
		if validators[0].Level != 0 {
			return errors.Errorf("(ValidatorDeleted) Expected %d, got %d (Validator.Level)",
				0, validators[0].Level)
		}
	}
	return nil
}

func testCertificateUpdated(certificate *pb.Certificate, commonID *big.Int) error {
	// Check that a Certificate entry is created after CertificateCreated event. We create a special certificate,
	// `Name`, that will be recorded directly into profile. There's two such certificate types: `Name` and `Country`.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		return errors.Wrap(err, "onCertificateCreated failed")
	}
	if certificateAttrs, err := getCertificates(monitorDWH); err != nil {
		return errors.Wrap(err, "getCertificates failed")
	} else {
		// Exactly one certificate should be created.
		if len(certificateAttrs) != 1 {
			return errors.Errorf("(CertificateCreated) Expected 1 Certificate, got %d", len(certificateAttrs))
		}
		if string(certificateAttrs[0].Value) != "User Name" {
			return errors.Errorf("(CertificateCreated) Expected %s, got %s (Certificate.Value)",
				"User Name", certificateAttrs[0].Value)
		}
	}
	// After a certificate is created, corresponding profile must be updated (with the value from certificate, in
	// this case with a name). (N.B.: a profile update will also update all orders and deals that reference this
	// profile, which will be checked below.)
	if profiles, _, err := monitorDWH.storage.GetProfiles(newSimpleConn(monitorDWH.db), &pb.ProfilesRequest{
		Sortings: []*pb.SortingOption{{Field: "Id", Order: pb.SortingOrder_Asc}},
	}); err != nil {
		return errors.Errorf("Failed to getProfiles: %s", err)
	} else {
		if len(profiles) != 13 {
			return errors.Errorf("(CertificateCreated) Expected 13 Profiles, got %d",
				len(profiles))
		}
		if profiles[12].Name != "User Name" {
			return errors.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"User Name", profiles[12].Name)
		}
	}
	// Now create a country certificate.
	certificate.Attribute = CertificateCountry
	certificate.Value = []byte("Country")
	// Check that a  Profile entry is updated after CertificateCreated event.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		return errors.Wrap(err, "onCertificateCreated failed")
	}
	if profiles, _, err := monitorDWH.storage.GetProfiles(newSimpleConn(monitorDWH.db), &pb.ProfilesRequest{}); err != nil {
		return errors.Wrap(err, "getProfiles failed")
	} else {
		if len(profiles) != 13 {
			return errors.Errorf("(CertificateCreated) Expected 1 Profile, got %d", len(profiles))
		}
		profiles := profiles
		if profiles[12].Country != "Country" {
			return errors.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Country)",
				"Country", profiles[12].Name)
		}
		if profiles[12].Name != "User Name" {
			return errors.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"Name", profiles[12].Name)
		}

		// All certificates (not only `Name` and `Country`) must be stored as JSON inside a profile entry.
		var certificates []*pb.Certificate
		if err := json.Unmarshal([]byte(profiles[12].Certificates), &certificates); err != nil {
			return errors.Errorf("(CertificateCreated) Failed to unmarshal Profile.Certificates: %s", err)
		} else {
			if len(certificates) != 2 {
				return errors.Errorf("(CertificateCreated) Expected 2 Certificates, got %d",
					len(certificates))
			}
		}
	}
	// Check that profile updates resulted in orders updates.
	dwhOrder, err := monitorDWH.storage.GetOrderByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return errors.Wrap(err, "storage.GetOrderByID failed")
	}
	if dwhOrder.CreatorIdentityLevel != 3 {
		return errors.Errorf("(CertificateCreated) Expected %d, got %d (Order.CreatorIdentityLevel)",
			2, dwhOrder.CreatorIdentityLevel)
	}
	// Check that profile updates resulted in deals updates.
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return errors.Wrap(err, "storage.GetDealByID failed")
	} else {
		if len(deal.SupplierCertificates) == 0 {
			return errors.Errorf("Expected some SupplierCertificates, got nothing")
		}
	}
	return nil
}

func testOrderUpdated(order *pb.Order, commonID *big.Int) error {
	// Check that if order is updated, it is deleted. Order should be deleted because its DealID is not set
	// (this means that is has become inactive due to a cancellation and not a match).
	order.OrderStatus = pb.OrderStatus_ORDER_INACTIVE
	if err := monitorDWH.onOrderUpdated(commonID); err != nil {
		return errors.Wrap(err, "onOrderUpdated failed")
	}
	if _, err := monitorDWH.storage.GetOrderByID(newSimpleConn(monitorDWH.db), commonID); err == nil {
		return errors.New("GetOrderDetails returned an order that should have been deleted")
	}
	return nil
}

func testDealUpdated(deal *pb.Deal, commonID *big.Int) error {
	deal.Duration += 1
	// Test onDealUpdated event handling.
	if err := monitorDWH.onDealUpdated(commonID); err != nil {
		return errors.Wrap(err, "onDealUpdated failed")
	}
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return errors.Wrap(err, "storage.GetDealByID failed")
	} else {
		if deal.GetDeal().Duration != 10021 {
			return errors.Errorf("Expected %d, got %d (Deal.Duration)", 10021, deal.GetDeal().Duration)
		}
	}
	return nil
}

func testDealChangeRequestSentAccepted(changeRequest *pb.DealChangeRequest, commonEventTS uint64, commonID *big.Int) error {
	// Test creating an ASK DealChangeRequest.
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(0)); err != nil {
		return errors.Wrap(err, "onDealChangeRequestSent failed")
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return errors.Wrap(err, "getDealChangeRequest failed")
	} else {
		if changeRequest.Duration != 10020 {
			return errors.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10020, changeRequest.Duration)
		}
	}
	// Check that after a second ASK DealChangeRequest was created, the new one was kept and the old one was deleted.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Duration = 10021
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(1)); err != nil {
		return errors.Wrap(err, "onDealChangeRequestSent (2) failed")
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return errors.Wrap(err, "getDealChangeRequest (2) failed")
	} else {
		if changeRequest.Duration != 10021 {
			return errors.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10021, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(0)); err == nil {
		return errors.New("getDealChangeRequest returned a DealChangeRequest that should have been deleted")
	}
	// Check that when a BID DealChangeRequest was created, it was kept (and nothing was deleted).
	changeRequest.Id = pb.NewBigIntFromInt(2)
	changeRequest.Duration = 10022
	changeRequest.RequestType = pb.OrderType_BID
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(2)); err != nil {
		return errors.Wrap(err, "onDealChangeRequestSent (3) failed")
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return errors.Wrap(err, "getDealChangeRequest (3) failed")
	} else {
		if changeRequest.Duration != 10022 {
			return errors.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10022, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(1)); err != nil {
		return errors.Errorf("DealChangeRequest of type ASK was deleted after a BID DealChangeRequest creation: %s", err)
	}
	// Check that when a DealChangeRequest is updated to any status but REJECTED, it is deleted.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_ACCEPTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(1)); err != nil {
		return errors.Wrap(err, "onDealChangeRequestUpdated failed")
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(1)); err == nil {
		return errors.New("DealChangeRequest which status was changed to ACCEPTED was not deleted")
	}
	// Check that when a DealChangeRequest is updated to REJECTED, it is kept.
	changeRequest.Id = pb.NewBigIntFromInt(2)
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_REJECTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(2)); err != nil {
		return errors.Wrap(err, "onDealChangeRequestUpdated (4) failed")
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(2)); err != nil {
		return errors.New("DealChangeRequest which status was changed to REJECTED was deleted")
	}
	// Also test that a new DealCondition was created, and the old one was updated.
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return errors.Errorf("Failed to GetDealConditions: %s", err)
	} else {
		if len(dealConditions) != 2 {
			return errors.Errorf("Expected 2 DealConditions, got %d", len(dealConditions))
		}
		conditions := dealConditions
		if conditions[1].EndTime.Seconds != 5 {
			return errors.Errorf("Expected %d, got %d (DealCondition.EndTime)", 5, conditions[1].EndTime.Seconds)
		}
		if conditions[0].StartTime.Seconds != 5 {
			return errors.Errorf("Expected %d, got %d (DealCondition.StartTime)", 5, conditions[0].StartTime.Seconds)
		}
	}
	return nil
}

func testBilled(commonEventTS uint64, commonID *big.Int) error {
	deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return errors.Wrap(err, "GetDealByID failed")
	}

	if deal.Deal.LastBillTS.Seconds != int64(commonEventTS) {
		return errors.Errorf("unexpected LastBillTS (%d)", deal.Deal.LastBillTS)
	}

	// Check that after a Billed event last DealCondition.Payout is updated.
	newBillTS := commonEventTS + 1
	if err := monitorDWH.onBilled(newBillTS, commonID, big.NewInt(10)); err != nil {
		return errors.Wrap(err, "onBilled failed")
	}
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return errors.Errorf("Failed to GetDealDetails: %s", err)
	} else {
		if len(dealConditions) != 2 {
			return errors.Errorf("(Billed) Expected 2 DealConditions, got %d", len(dealConditions))
		}
		conditions := dealConditions
		if conditions[0].TotalPayout.Unwrap().String() != "10" {
			return errors.Errorf("(Billed) Expected %s, got %s (DealCondition.TotalPayout)",
				"10", conditions[0].TotalPayout.Unwrap().String())
		}
	}
	updatedDeal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return errors.Wrap(err, "GetDealByID failed")
	}

	if updatedDeal.Deal.LastBillTS.Seconds != int64(newBillTS) {
		return errors.Errorf("(Billed) Expected %d, got %d (Deal.LastBillTS)",
			newBillTS, updatedDeal.Deal.LastBillTS.Seconds)
	}

	return nil
}

func testDealClosed(deal *pb.Deal, commonID *big.Int) error {
	// Check that when a Deal's status is updated to CLOSED, Deal and its DealConditions are deleted.
	deal.Status = pb.DealStatus_DEAL_CLOSED
	// Test onDealUpdated event handling.
	if err := monitorDWH.onDealUpdated(commonID); err != nil {
		return errors.Wrap(err, "onDealUpdated")
	}
	if _, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err == nil {
		return errors.Errorf("Deal was not deleted after status changing to CLOSED")
	}
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return errors.Errorf("Failed to GetDealConditions: %s", err)
	} else {
		if len(dealConditions) != 0 {
			return errors.Errorf("(DealUpdated) Expected 0 DealConditions, got %d", len(dealConditions))
		}
	}
	return nil
}

func testWorkerAnnouncedConfirmedRemoved() error {
	// Check that a worker is added after a WorkerAnnounced event.
	if err := monitorDWH.onWorkerAnnounced(common.HexToAddress("0xC").Hex(),
		common.HexToAddress("0xD").Hex()); err != nil {
		return errors.Wrap(err, "onWorkerAnnounced failed")
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return errors.Wrap(err, "getWorkers failed")
	} else {
		if len(workers) != 1 {
			return errors.Errorf("(WorkerAnnounced) Expected 1 Worker, got %d", len(workers))
		}
		if workers[0].Confirmed {
			return errors.Errorf("(WorkerAnnounced) Expected %t, got %t (Worker.Confirmed)",
				false, workers[0].Confirmed)
		}
	}
	// Check that a worker is confirmed after a WorkerConfirmed event.
	if err := monitorDWH.onWorkerConfirmed(common.HexToAddress("0xC").Hex(),
		common.HexToAddress("0xD").Hex()); err != nil {
		return errors.Wrap(err, "onWorkerConfirmed failed")
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return errors.Wrap(err, "getWorkers failed")
	} else {
		if len(workers) != 1 {
			return errors.Errorf("(WorkerConfirmed) Expected 1 Worker, got %d", len(workers))
		}
		if !workers[0].Confirmed {
			return errors.Errorf("(WorkerConfirmed) Expected %t, got %t (Worker.Confirmed)",
				true, workers[0].Confirmed)
		}
	}
	// Check that a worker is deleted after a WorkerRemoved event.
	if err := monitorDWH.onWorkerRemoved(common.HexToAddress("0xC").Hex(),
		common.HexToAddress("0xD").Hex()); err != nil {
		return errors.Wrap(err, "onWorkerRemoved failed")
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return errors.Wrap(err, "getWorkers failed")
	} else {
		if len(workers) != 0 {
			return errors.Errorf("(WorkerRemoved) Expected 0 Workers, got %d", len(workers))
		}
	}
	return nil
}

func testBlacklistAddedRemoved() error {
	// Check that a Blacklist entry is added after AddedToBlacklist event.
	if err := monitorDWH.onAddedToBlacklist(common.HexToAddress("0xC").Hex(),
		common.HexToAddress("0xD").Hex()); err != nil {
		return errors.Wrap(err, "onAddedToBlacklist failed")
	}
	if blacklistReply, err := monitorDWH.storage.GetBlacklist(
		newSimpleConn(monitorDWH.db), &pb.BlacklistRequest{OwnerID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return errors.Wrap(err, "getBlacklist failed")
	} else {
		if blacklistReply.OwnerID.Unwrap().Hex() != common.HexToAddress("0xC").Hex() {
			return errors.Errorf("(AddedToBlacklist) Expected %s, got %s (BlacklistReply.AdderID)",
				common.HexToAddress("0xC").Hex(), blacklistReply.OwnerID)
		}
	}
	// Check that a Blacklist entry is deleted after RemovedFromBlacklist event.
	if err := monitorDWH.onRemovedFromBlacklist(common.HexToAddress("0xC").Hex(),
		common.HexToAddress("0xD").Hex()); err != nil {
		return errors.Wrap(err, "onRemovedFromBlacklist failed")
	}
	if repl, err := monitorDWH.storage.GetBlacklist(
		newSimpleConn(monitorDWH.db), &pb.BlacklistRequest{OwnerID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return errors.Wrap(err, "getBlacklist (2) failed")
	} else {
		if len(repl.Addresses) > 0 {
			return errors.Errorf("GetBlacklist returned a blacklist that should have been deleted: %+v", repl.Addresses)
		}
	}
	return nil
}

func getDealChangeRequest(w *DWH, changeRequestID *pb.BigInt) (*pb.DealChangeRequest, error) {
	rows, err := w.db.Query("SELECT * FROM DealChangeRequests WHERE Id=?", changeRequestID.Unwrap().String())
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return globalDWH.storage.(*sqlStorage).decodeDealChangeRequest(rows)
}

func getCertificates(w *DWH) ([]*pb.Certificate, error) {
	rows, err := w.db.Query("SELECT * FROM Certificates")
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.Certificate
	for rows.Next() {
		if certificate, err := w.storage.(*sqlStorage).decodeCertificate(rows); err != nil {
			return nil, errors.Wrap(err, "failed to decodeCertificate")
		} else {
			out = append(out, certificate)
		}
	}

	return out, nil
}

type dealPayment struct {
	BilledTS   uint64
	PaidAmount string
	DealID     string
}

func setupTestDB(w *DWH) error {
	if err := w.setupSQLite(w.db, 12); err != nil {
		return err
	}

	var certs = []*pb.Certificate{
		{OwnerID: pb.NewEthAddress(common.HexToAddress("0xBB")), Value: []byte("Consumer"), Attribute: CertificateName},
	}
	byteCerts, _ := json.Marshal(certs)

	insertDeal := `INSERT INTO Deals(Id, SupplierID, ConsumerID, MasterID, AskID, BidID, Duration, Price, StartTime,
		EndTime, Status, BlockedBalance, TotalPayout, LastBillTS, Netflags, AskIdentityLevel, BidIdentityLevel,
		SupplierCertificates, ConsumerCertificates, ActiveChangeRequest, Benchmark0, Benchmark1, Benchmark2,
		Benchmark3, Benchmark4, Benchmark5, Benchmark6, Benchmark7, Benchmark8, Benchmark9, Benchmark10, Benchmark11)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertOrder := `INSERT INTO Orders(Id, CreatedTS, DealID, Type, Status, AuthorID, CounterpartyID, Duration,
		Price, Netflags, IdentityLevel, Blacklist, Tag, FrozenSum, CreatorIdentityLevel, CreatorName, CreatorCountry,
		CreatorCertificates, Benchmark0, Benchmark1, Benchmark2, Benchmark3, Benchmark4, Benchmark5, Benchmark6,
		Benchmark7, Benchmark8, Benchmark9, Benchmark10, Benchmark11) VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertDealChangeRequest := `INSERT INTO DealChangeRequests VALUES ($1, $2, $3, $4, $5, $6, $7)`
	for i := 0; i < 10; i++ {
		_, err := w.db.Exec(
			insertDeal,
			fmt.Sprintf("4040%d", i),
			common.HexToAddress(fmt.Sprintf("0x1%d", i)).Hex(), // Supplier
			common.HexToAddress(fmt.Sprintf("0x2%d", i)).Hex(), // Consumer
			common.HexToAddress(fmt.Sprintf("0x3%d", i)).Hex(), // Master
			fmt.Sprintf("2020%d", i),
			fmt.Sprintf("3030%d", i),
			10010+i, // Duration
			pb.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			30010+i, // StartTime
			40010+i, // EndTime
			uint64(pb.DealStatus_DEAL_ACCEPTED),
			pb.NewBigIntFromInt(50010+int64(i)).PaddedString(), // BlockedBalance
			pb.NewBigIntFromInt(60010+int64(i)).PaddedString(), // TotalPayout
			70010+i,   // LastBillTS
			5,         // Netflags
			3,         // AskIdentityLevel
			4,         // BidIdentityLevel
			byteCerts, // SupplierCertificates
			byteCerts, // ConsumerCertificates
			true,
			10, // CPUSysbenchMulti
			20,
			30,
			40,
			50,
			60,
			70,
			80,
			90,
			100, // GPUEthHashrate
			110,
			120,
		)
		if err != nil {
			return errors.Wrap(err, "failed to insertDeal")
		}

		_, err = w.db.Exec(
			insertOrder,
			fmt.Sprintf("2020%d", i),
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(pb.OrderType_ASK),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xA").Hex(), // AuthorID
			common.HexToAddress("0xB").Hex(), // CounterpartyID
			10010+i,
			pb.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			7, // Netflags
			uint64(pb.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},          // Tag
			fmt.Sprintf("3001%d", i), // FrozenSum
			uint64(pb.IdentityLevel_PSEUDONYMOUS),
			"CreatorName",
			"CreatorCountry",
			byteCerts, // CreatorCertificates
			10+i,
			20+i,
			30+i,
			40+i,
			50+i,
			60+i,
			70+i,
			80+i,
			90+i,
			100+i,
			110+i,
			120+i,
		)
		if err != nil {
			return err
		}

		_, err = w.db.Exec(
			insertOrder,
			fmt.Sprintf("3030%d", i),
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(pb.OrderType_BID),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xB").Hex(), // AuthorID
			common.HexToAddress("0xA").Hex(), // CounterpartyID
			10010-i,                          // Duration
			pb.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			5, // Netflags
			uint64(pb.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},                       // Tag
			fmt.Sprintf("3001%d", i),              // FrozenSum
			uint64(pb.IdentityLevel_PSEUDONYMOUS), // CreatorIdentityLevel
			"CreatorName",
			"CreatorCountry",
			byteCerts, // CreatorCertificates
			10-i,
			20-i,
			30-i,
			40-i,
			50-i,
			60-i,
			70-i,
			80-i,
			90-i,
			100-i,
			110-i,
			120-i,
		)
		if err != nil {
			return err
		}

		_, err = w.db.Exec(insertDealChangeRequest,
			fmt.Sprintf("5050%d", i), 0, 0, 0, 0, 0, "40400")
		if err != nil {
			return err
		}

		var identityLevel int
		if (i % 2) == 0 {
			identityLevel = 0
		} else {
			identityLevel = 1
		}
		_, err = w.db.Exec("INSERT INTO Profiles VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			common.HexToAddress(fmt.Sprintf("0x2%d", i)).Hex(), identityLevel, "sortedProfile", "", 0, 0, []byte{}, 0, 0)
		if err != nil {
			return err
		}
	}

	_, err := w.db.Exec("INSERT INTO Profiles VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fmt.Sprintf(common.HexToAddress("0xBB").Hex()), 3, "Consumer", "", 0, 0, byteCerts, 10, 10)
	if err != nil {
		return err
	}

	_, err = w.db.Exec("INSERT INTO Profiles VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fmt.Sprintf(common.HexToAddress("0xAA").Hex()), 3, "Supplier", "", 0, 0, byteCerts, 10, 10)
	if err != nil {
		return err
	}

	_, err = w.db.Exec("INSERT INTO Blacklists VALUES (?, ?)", common.HexToAddress("0xE").Hex(), common.HexToAddress("0xBB").Hex())
	if err != nil {
		return err
	}

	return nil
}
