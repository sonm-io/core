package dwh

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	log "github.com/noxiouz/zapctx/ctxlog"
	bch "github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
)

var (
	globalDWH      *DWH
	monitorDWH     *DWH
	dbUser         = "dwh_tester"
	dbUserPassword = "dwh_tester"
	globalDBName   = "dwh_test_global"
	monitorDBName  = "dwh_test_monitor"
)

func TestMain(m *testing.M) {
	var (
		testsReturnCode int
		err             error
	)
	if err := setupDB(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := globalDWH.db.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := monitorDWH.db.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if err := tearDownDB(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(testsReturnCode)
	}()

	globalDWH, err = getTestDWH(getConnString(globalDBName, dbUser, dbUserPassword))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	monitorDWH, err = getTestDWH(getConnString(monitorDBName, dbUser, dbUserPassword))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testsReturnCode = m.Run()
}

func setupDB() error {
	db, err := sql.Open("postgres", "postgresql://localhost:5432/template1?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect to template1: %s", err)
	}
	defer db.Close()

	for _, dbName := range []string{globalDBName, monitorDBName} {
		if _, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)); err != nil {
			return fmt.Errorf("failed to preliminarily drop database: %s", err)
		}
	}

	if _, err := db.Exec(fmt.Sprintf("DROP USER IF EXISTS %s", dbUser)); err != nil {
		return fmt.Errorf("failed to preliminarily drop user: %s", err)
	}
	if _, err = db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", dbUser, dbUserPassword)); err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	for _, dbName := range []string{globalDBName, monitorDBName} {
		if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER dwh_tester", dbName)); err != nil {
			return fmt.Errorf("failed to create database: %s", err)
		}
	}

	return nil
}

func tearDownDB() error {
	db, err := sql.Open("postgres", "postgresql://localhost:5432/template1?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect to template1: %s", err)
	}
	defer db.Close()

	for _, dbName := range []string{globalDBName, monitorDBName} {
		if _, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)); err != nil {
			return fmt.Errorf("failed to drop database: %s", err)
		}
	}

	return nil
}

func getConnString(database, user, password string) string {
	return fmt.Sprintf("postgresql://localhost:5432/%s?user=%s&password=%s&sslmode=disable", database, user, password)
}

func getTestDWH(dbEndpoint string) (*DWH, error) {
	var (
		ctx = context.Background()
		cfg = &Config{
			Storage: &storageConfig{
				Endpoint: dbEndpoint,
			},
		}
		controller     = gomock.NewController(&testing.T{})
		mockBlockchain = bch.NewMockAPI(controller)
		mockMarket     = bch.NewMockMarketAPI(controller)
	)
	mockMarket.EXPECT().GetNumBenchmarks(gomock.Any()).AnyTimes().Return(uint64(12), nil)
	mockBlockchain.EXPECT().Market().AnyTimes().Return(mockMarket)

	db, err := sql.Open("postgres", cfg.Storage.Endpoint)
	if err != nil {
		return nil, err
	}

	w := &DWH{
		blockchain: mockBlockchain,
		ctx:        ctx,
		cfg:        cfg,
		db:         db,
		logger:     log.GetLogger(ctx),
	}

	return w, setupTestDB(w)
}

func TestDWH_GetDeals(t *testing.T) {
	var (
		byAddress            = common.HexToAddress("0x11")
		byMinDuration uint64 = 10011
		byMinPrice           = big.NewInt(20011)
	)

	// Test TEXT columns.
	{
		request := &pb.DealsRequest{SupplierID: pb.NewEthAddress(byAddress)}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Deals) != 1 {
			t.Errorf("Expected 1 deal in reply, got %d", len(reply.Deals))
			return
		}

		var deal = reply.Deals[0].Deal
		if deal.SupplierID.Unwrap() != byAddress {
			t.Errorf("Request `%+v` failed, expected %s, got %s (SupplierID)",
				request, byAddress.Hex(), deal.SupplierID.Unwrap().Hex())
			return
		}
	}
	// Test INTEGER columns.
	{
		request := &pb.DealsRequest{
			Duration: &pb.MaxMinUint64{Min: byMinDuration},
		}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}
		if len(reply.Deals) != 1 {
			t.Errorf("Request `%+v` failed: Expected 1 deal in reply, got %d", request, len(reply.Deals))
			return
		}
	}
	// Test TEXT columns which should be treated as INTEGERS.
	{
		request := &pb.DealsRequest{
			Price: &pb.MaxMinBig{Min: pb.NewBigInt(byMinPrice)},
		}
		reply, err := globalDWH.GetDeals(globalDWH.ctx, request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}
		if len(reply.Deals) != 1 {
			t.Errorf("Request `%+v` failed: Expected 1 deal in reply, got %d", request, len(reply.Deals))
			return
		}
		var deal = reply.Deals[0].Deal
		if deal.Price.Unwrap().String() != byMinPrice.String() {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Price)",
				request, byMinPrice.Int64(), deal.Price.Unwrap().Int64())
			return
		}
	}
}

func TestDWH_GetDealDetails(t *testing.T) {
	var (
		byDealID = big.NewInt(40400)
	)
	reply, err := globalDWH.storage.GetDealByID(newSimpleConn(globalDWH.db), byDealID)
	if err != nil {
		t.Error(err)
		return
	}
	deal := reply.GetDeal()
	if deal.Id.Unwrap().Cmp(byDealID) != 0 {
		t.Errorf("Expected %d, got %d (Id)", byDealID.Int64(), deal.Id.Unwrap().Int64())
	}
}

func TestDWH_GetOrders(t *testing.T) {
	var (
		byDealID              = big.NewInt(10101)
		byMinDuration  uint64 = 10011
		byMinBenchmark uint64 = 11
		byMinPrice            = big.NewInt(20011)
	)
	// Test TEXT columns.
	{
		request := &pb.OrdersRequest{DealID: pb.NewBigInt(byDealID)}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}
		if len(orders) != 2 {
			t.Errorf("Request `%+v` failed: Expected 2 orders in reply (ASK and BID), got %d", request, len(orders))
			return
		}
		var order = orders[0].Order
		if order.AuthorID.Unwrap() != common.HexToAddress("0xA") {
			t.Errorf("Request `%+v` failed, expected %s, got %s (AuthorID)",
				request, common.HexToAddress("0xA").Hex(), order.AuthorID.Unwrap().Hex())
			return
		}
	}
	// Test INTEGER columns.
	{
		request := &pb.OrdersRequest{
			Duration:   &pb.MaxMinUint64{Min: byMinDuration},
			Benchmarks: map[uint64]*pb.MaxMinUint64{0: {Min: byMinBenchmark}},
		}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}
		if len(orders) != 1 {
			t.Errorf("Request `%+v` failed: Expected 1 order in reply, got %d", request, len(orders))
			return
		}
		var order = orders[0].Order
		if order.Duration != byMinDuration {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Duration)", request, byMinDuration, order.Duration)
			return
		}
	}
	// Test TEXT columns which should be treated as INTEGERS.
	{
		request := &pb.OrdersRequest{
			Type:  pb.OrderType_ASK,
			Price: &pb.MaxMinBig{Min: pb.NewBigInt(byMinPrice)},
		}
		orders, _, err := globalDWH.storage.GetOrders(newSimpleConn(globalDWH.db), request)
		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}
		if len(orders) != 1 {
			t.Errorf("Request `%+v` failed: Expected 1 order in reply, got %d", request, len(orders))
			return
		}
		var order = orders[0].Order
		if order.Price.Unwrap().Cmp(byMinPrice) != 0 {
			t.Errorf("Request `%+v` failed: expected %d, got %d (Price)",
				request, byMinPrice.Int64(), order.Price.Unwrap().Int64())
			return
		}
	}
}

func TestDWH_GetMatchingOrders(t *testing.T) {
	var byID = big.NewInt(20201)
	request := &pb.MatchingOrdersRequest{Id: pb.NewBigInt(byID)}
	orders, _, err := globalDWH.storage.GetMatchingOrders(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: GetMatchingOrders failed: %s", request, err)
		return
	}
	if len(orders) != 1 {
		t.Errorf("Request `%+v` failed: Expected 5 orders in reply, got %d", request, len(orders))
		return
	}
}

func TestDWH_GetOrderDetails(t *testing.T) {
	var byID = big.NewInt(20201)
	order, err := globalDWH.storage.GetOrderByID(newSimpleConn(globalDWH.db), byID)
	if err != nil {
		t.Error(err)
		return
	}
	reply := order.GetOrder()
	if reply.Id.Unwrap().Cmp(byID) != 0 {
		t.Errorf("Request `%d` failed: Expected %d, got %d (Id)", byID.Int64(), byID.Int64(), reply.Id.Unwrap().Int64())
	}
}

func TestDWH_GetDealChangeRequests(t *testing.T) {
	changeRequests, err := globalDWH.getDealChangeRequests(newSimpleConn(globalDWH.db), pb.NewBigIntFromInt(40400))
	if err != nil {
		t.Error(err)
		return
	}
	if len(changeRequests) != 2 {
		t.Errorf("Request `%d` failed: Expected %d DealChangeRequests, got %d", 404000, 2, len(changeRequests))
		return
	}
}

func TestDWH_GetProfiles(t *testing.T) {
	var request = &pb.ProfilesRequest{
		Identifier: "sortedProfile",
		Sortings: []*pb.SortingOption{
			{
				Field: "UserID",
				Order: pb.SortingOrder_Asc,
			},
		},
	}
	profiles, _, err := globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 2 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 10, len(profiles))
		return
	}
	if profiles[0].UserID.Unwrap() != common.HexToAddress("0x20") {
		t.Errorf("Request `%+v` failed: Expected %s, got %s (Profile.UserID)",
			request, common.HexToAddress("0x20").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}
	request = &pb.ProfilesRequest{
		Identifier: "sortedProfile",
		Sortings: []*pb.SortingOption{
			{
				Field: "UserID",
				Order: pb.SortingOrder_Desc,
			},
		},
	}
	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 2 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 2, len(profiles))
		return
	}
	if profiles[0].UserID.Unwrap() != common.HexToAddress("0x21") {
		t.Errorf("Request `%+v` failed: Expected %s, got %s (Profile.UserID)",
			request, common.HexToAddress("0x21").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}
	request = &pb.ProfilesRequest{
		Identifier: "sortedProfile",
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
	}
	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 2 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 2, len(profiles))
		return
	}
	if profiles[0].UserID.Unwrap().Hex() != common.HexToAddress("0x20").Hex() {
		t.Errorf("Request `%+v` failed: Expected %s, got %s (Profile.UserID)",
			request, common.HexToAddress("0x20").Hex(), profiles[0].UserID.Unwrap().Hex())
		return
	}
	if profiles[1].UserID.Unwrap().Hex() != common.HexToAddress("0x21").Hex() {
		t.Errorf("Request `%+v` failed: Expected %s, got %s (Profile.UserID)",
			request, common.HexToAddress("0x28").Hex(), profiles[1].UserID.Unwrap().Hex())
		return
	}
	request = &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_OnlyMatching,
		},
	}
	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 1 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 1, len(profiles))
		return
	}
	request = &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_WithoutMatching,
		},
	}
	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), request)
	if err != nil {
		t.Error(err)
		return
	}
	if len(profiles) != 3 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 3, len(profiles))
		return
	}
	profiles, _, err = globalDWH.storage.GetProfiles(newSimpleConn(globalDWH.db), &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: pb.NewEthAddress(common.HexToAddress("0xE")),
			Option:  pb.BlacklistOption_IncludeAndMark,
		},
	})
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 4 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 4, len(profiles))
		return
	}
	var foundMarkedProfile bool
	for _, profile := range profiles {
		if profile.IsBlacklisted {
			foundMarkedProfile = true
		}
	}
	if !foundMarkedProfile {
		t.Errorf("Request `%+v` failed: Failed to find profile marked as blacklisted", request)
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
		AskID:          pb.NewBigIntFromInt(20200),
		BidID:          pb.NewBigIntFromInt(30300),
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
		Netflags:       &pb.NetFlags{Flags: uint64(7)},
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
	mockProfiles.EXPECT().GetValidatorLevel(gomock.Any(), gomock.Any()).AnyTimes().Return(int8(0), nil)
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

	err = monitorDWH.storage.InsertWorker(newSimpleConn(monitorDWH.db), common.Address{},
		common.HexToAddress("0x000000000000000000000000000000000000000d"))
	if err != nil {
		t.Error("failed to insert worker (additional)")
	}

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
	err = monitorDWH.storage.DeleteWorker(newSimpleConn(monitorDWH.db), common.Address{},
		common.HexToAddress("0x000000000000000000000000000000000000000d"))
	if err != nil {
		t.Error("failed to delete worker (additional)")
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
		return fmt.Errorf("onOrderPlaced failed: %v", err)
	}
	if order, err := monitorDWH.storage.GetOrderByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return fmt.Errorf("storage.GetOrderByID failed: %v", err)
	} else {
		if order.GetOrder().Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (Order.Duration)", 10020, order.GetOrder().Duration)
		}
	}
	return nil
}

func testDealOpened(deal *pb.Deal, commonID *big.Int) error {
	if err := monitorDWH.onDealOpened(commonID); err != nil {
		return fmt.Errorf("onDealOpened failed: %v", err)
	}
	// Firstly, check that a deal was created.
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if deal.GetDeal().Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (Deal.Duration)", 10020, deal.GetDeal().Duration)
		}
	}
	// Secondly, check that a DealCondition was created.
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return fmt.Errorf("getDealConditions failed: %v", err)
	} else {
		if dealConditions[0].Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (DealCondition.Duration)", 10020, deal.Duration)
		}
	}
	return nil
}

func testValidatorCreatedUpdated(validator *pb.Validator) error {
	// Check that a Validator entry is added after ValidatorCreated event.
	if err := monitorDWH.onValidatorCreated(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return fmt.Errorf("onValidatorCreated failed: %v", err)
	}
	if validators, _, err := monitorDWH.storage.GetValidators(newSimpleConn(monitorDWH.db), &pb.ValidatorsRequest{}); err != nil {
		return fmt.Errorf("getValidators failed: %v", err)
	} else {
		if len(validators) != 1 {
			return fmt.Errorf("(ValidatorCreated) Expected 1 Validator, got %d", len(validators))
		}
		if validators[0].GetValidator().GetLevel() != 3 {
			return fmt.Errorf("(ValidatorCreated) Expected %d, got %d (Validator.Level)",
				3, validators[0].GetValidator().GetLevel())
		}
	}
	validator.Level = 0
	// Check that a Validator entry is updated after ValidatorDeleted event.
	if err := monitorDWH.onValidatorDeleted(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return fmt.Errorf("onValidatorDeleted failed: %v", err)
	}
	if validators, _, err := monitorDWH.storage.GetValidators(newSimpleConn(monitorDWH.db), &pb.ValidatorsRequest{}); err != nil {
		return fmt.Errorf("getValidators failed: %v", err)
	} else {
		if len(validators) != 1 {
			return fmt.Errorf("(ValidatorDeleted) Expected 1 Validator, got %d", len(validators))
		}
		if validators[0].GetValidator().GetLevel() != 0 {
			return fmt.Errorf("(ValidatorDeleted) Expected %d, got %d (Validator.Level)",
				0, validators[0].GetValidator().GetLevel())
		}
	}
	return nil
}

func testCertificateUpdated(certificate *pb.Certificate, commonID *big.Int) error {
	// Check that a Certificate entry is created after CertificateCreated event. We create a special certificate,
	// `Name`, that will be recorded directly into profile. There's two such certificate types: `Name` and `Country`.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		return fmt.Errorf("onCertificateCreated failed: %v", err)
	}
	if certificateAttrs, err := getCertificates(monitorDWH); err != nil {
		return fmt.Errorf("getCertificates failed: %v", err)
	} else {
		// Exactly one certificate should be created.
		if len(certificateAttrs) != 1 {
			return fmt.Errorf("(CertificateCreated) Expected 1 Certificate, got %d", len(certificateAttrs))
		}
		if string(certificateAttrs[0].Value) != "User Name" {
			return fmt.Errorf("(CertificateCreated) Expected %s, got %s (Certificate.Value)",
				"User Name", certificateAttrs[0].Value)
		}
	}
	// After a certificate is created, corresponding profile must be updated (with the value from certificate, in
	// this case with a name). (N.B.: a profile update will also update all orders and deals that reference this
	// profile, which will be checked below.)
	if profiles, _, err := monitorDWH.storage.GetProfiles(newSimpleConn(monitorDWH.db), &pb.ProfilesRequest{
		Sortings: []*pb.SortingOption{{Field: "Id", Order: pb.SortingOrder_Asc}},
	}); err != nil {
		return fmt.Errorf("failed to getProfiles: %s", err)
	} else {
		if len(profiles) != 5 {
			return fmt.Errorf("(CertificateCreated) Expected 5 Profiles, got %d",
				len(profiles))
		}
		if profiles[4].Name != "User Name" {
			return fmt.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"User Name", profiles[4].Name)
		}
	}
	// Now create a country certificate.
	certificate.Attribute = CertificateCountry
	certificate.Value = []byte("Country")
	// Check that a  Profile entry is updated after CertificateCreated event.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		return fmt.Errorf("onCertificateCreated failed: %v", err)
	}
	if profiles, _, err := monitorDWH.storage.GetProfiles(newSimpleConn(monitorDWH.db), &pb.ProfilesRequest{}); err != nil {
		return fmt.Errorf("getProfiles failed: %v", err)
	} else {
		if len(profiles) != 5 {
			return fmt.Errorf("(CertificateCreated) Expected 5 Profiles, got %d", len(profiles))
		}
		profiles := profiles
		if profiles[4].Country != "Country" {
			return fmt.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Country)",
				"Country", profiles[4].Name)
		}
		if profiles[4].Name != "User Name" {
			return fmt.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"Name", profiles[4].Name)
		}

		// All certificates (not only `Name` and `Country`) must be stored as JSON inside a profile entry.
		var certificates []*pb.Certificate
		if err := json.Unmarshal([]byte(profiles[4].Certificates), &certificates); err != nil {
			return fmt.Errorf("(CertificateCreated) Failed to unmarshal Profile.Certificates: %s", err)
		} else {
			if len(certificates) != 2 {
				return fmt.Errorf("(CertificateCreated) Expected 2 Certificates, got %d",
					len(certificates))
			}
		}
	}
	// Check that profile updates resulted in orders updates.
	dwhOrder, err := monitorDWH.storage.GetOrderByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return fmt.Errorf("storage.GetOrderByID failed: %v", err)
	}
	if dwhOrder.CreatorIdentityLevel != 3 {
		return fmt.Errorf("(CertificateCreated) Expected %d, got %d (Order.CreatorIdentityLevel)",
			2, dwhOrder.CreatorIdentityLevel)
	}
	// Check that profile updates resulted in deals updates.
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if len(deal.SupplierCertificates) == 0 {
			return fmt.Errorf("expected some SupplierCertificates, got nothing")
		}
	}
	return nil
}

func testOrderUpdated(order *pb.Order, commonID *big.Int) error {
	// Check that if order is updated, it is deleted. Order should be deleted because its DealID is not set
	// (this means that is has become inactive due to a cancellation and not a match).
	order.OrderStatus = pb.OrderStatus_ORDER_INACTIVE
	if err := monitorDWH.onOrderUpdated(commonID); err != nil {
		return fmt.Errorf("onOrderUpdated failed: %v", err)
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
		return fmt.Errorf("onDealUpdated failed: %v", err)
	}
	if deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if deal.GetDeal().Duration != 10021 {
			return fmt.Errorf("expected %d, got %d (Deal.Duration)", 10021, deal.GetDeal().Duration)
		}
	}
	return nil
}

func testDealChangeRequestSentAccepted(changeRequest *pb.DealChangeRequest, commonEventTS uint64, commonID *big.Int) error {
	// Test creating an ASK DealChangeRequest.
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(0)); err != nil {
		return fmt.Errorf("onDealChangeRequestSent failed: %v", err)
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return fmt.Errorf("getDealChangeRequest failed: %v", err)
	} else {
		if changeRequest.Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (DealChangeRequest.Duration)", 10020, changeRequest.Duration)
		}
	}
	// Check that after a second ASK DealChangeRequest was created, the new one was kept and the old one was deleted.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Duration = 10021
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(1)); err != nil {
		return fmt.Errorf("onDealChangeRequestSent (2) failed: %v", err)
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return fmt.Errorf("getDealChangeRequest (2) failed: %v", err)
	} else {
		if changeRequest.Duration != 10021 {
			return fmt.Errorf("expected %d, got %d (DealChangeRequest.Duration)", 10021, changeRequest.Duration)
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
		return fmt.Errorf("onDealChangeRequestSent (3) failed: %v", err)
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		return fmt.Errorf("getDealChangeRequest (3) failed: %v", err)
	} else {
		if changeRequest.Duration != 10022 {
			return fmt.Errorf("expected %d, got %d (DealChangeRequest.Duration)", 10022, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(1)); err != nil {
		return fmt.Errorf("dealChangeRequest of type ASK was deleted after a BID DealChangeRequest creation: %s", err)
	}
	// Check that when a DealChangeRequest is updated to any status but REJECTED, it is deleted.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_ACCEPTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(1)); err != nil {
		return fmt.Errorf("onDealChangeRequestUpdated failed: %v", err)
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(1)); err == nil {
		return errors.New("dealChangeRequest which status was changed to ACCEPTED was not deleted")
	}
	// Check that when a DealChangeRequest is updated to REJECTED, it is kept.
	changeRequest.Id = pb.NewBigIntFromInt(2)
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_REJECTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(2)); err != nil {
		return fmt.Errorf("onDealChangeRequestUpdated (4) failed: %v", err)
	}
	if _, err := getDealChangeRequest(monitorDWH, pb.NewBigIntFromInt(2)); err != nil {
		return errors.New("dealChangeRequest which status was changed to REJECTED was deleted")
	}
	// Also test that a new DealCondition was created, and the old one was updated.
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return fmt.Errorf("failed to GetDealConditions: %s", err)
	} else {
		if len(dealConditions) != 2 {
			return fmt.Errorf("expected 2 DealConditions, got %d", len(dealConditions))
		}
		conditions := dealConditions
		if conditions[1].EndTime.Seconds != 5 {
			return fmt.Errorf("expected %d, got %d (DealCondition.EndTime)", 5, conditions[1].EndTime.Seconds)
		}
		if conditions[0].StartTime.Seconds != 5 {
			return fmt.Errorf("expected %d, got %d (DealCondition.StartTime)", 5, conditions[0].StartTime.Seconds)
		}
	}
	return nil
}

func testBilled(commonEventTS uint64, commonID *big.Int) error {
	deal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return fmt.Errorf("GetDealByID failed: %v", err)
	}

	if deal.Deal.LastBillTS.Seconds != int64(commonEventTS) {
		return fmt.Errorf("unexpected LastBillTS (%d)", deal.Deal.LastBillTS)
	}

	// Check that after a Billed event last DealCondition.Payout is updated.
	newBillTS := commonEventTS + 1
	if err := monitorDWH.onBilled(newBillTS, commonID, big.NewInt(10)); err != nil {
		return fmt.Errorf("onBilled failed: %v", err)
	}
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return fmt.Errorf("failed to GetDealDetails: %s", err)
	} else {
		if len(dealConditions) != 2 {
			return fmt.Errorf("(Billed) Expected 2 DealConditions, got %d", len(dealConditions))
		}
		conditions := dealConditions
		if conditions[0].TotalPayout.Unwrap().String() != "10" {
			return fmt.Errorf("(Billed) Expected %s, got %s (DealCondition.TotalPayout)",
				"10", conditions[0].TotalPayout.Unwrap().String())
		}
	}
	updatedDeal, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID)
	if err != nil {
		return fmt.Errorf("GetDealByID failed: %v", err)
	}

	if updatedDeal.Deal.LastBillTS.Seconds != int64(newBillTS) {
		return fmt.Errorf("(Billed) Expected %d, got %d (Deal.LastBillTS)",
			newBillTS, updatedDeal.Deal.LastBillTS.Seconds)
	}

	return nil
}

func testDealClosed(deal *pb.Deal, commonID *big.Int) error {
	// Check that when a Deal's status is updated to CLOSED, Deal and its DealConditions are deleted.
	deal.Status = pb.DealStatus_DEAL_CLOSED
	// Test onDealUpdated event handling.
	if err := monitorDWH.onDealUpdated(commonID); err != nil {
		return fmt.Errorf("onDealUpdated failed: %v", err)
	}
	if _, err := monitorDWH.storage.GetDealByID(newSimpleConn(monitorDWH.db), commonID); err == nil {
		return fmt.Errorf("deal was not deleted after status changing to CLOSED")
	}
	if dealConditions, _, err := monitorDWH.storage.GetDealConditions(
		newSimpleConn(monitorDWH.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return fmt.Errorf("failed to GetDealConditions: %s", err)
	} else {
		if len(dealConditions) != 0 {
			return fmt.Errorf("(DealUpdated) Expected 0 DealConditions, got %d", len(dealConditions))
		}
	}
	return nil
}

func testWorkerAnnouncedConfirmedRemoved() error {
	// Check that a worker is added after a WorkerAnnounced event.
	if err := monitorDWH.onWorkerAnnounced(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerAnnounced failed: %v", err)
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return fmt.Errorf("getWorkers failed: %v", err)
	} else {
		if len(workers) != 1 {
			return fmt.Errorf("(WorkerAnnounced) Expected 1 Worker, got %d", len(workers))
		}
		if workers[0].Confirmed {
			return fmt.Errorf("(WorkerAnnounced) Expected %t, got %t (Worker.Confirmed)",
				false, workers[0].Confirmed)
		}
	}
	// Check that a worker is confirmed after a WorkerConfirmed event.
	if err := monitorDWH.onWorkerConfirmed(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerConfirmed failed: %v", err)
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return fmt.Errorf("getWorkers failed: %v", err)
	} else {
		if len(workers) != 1 {
			return fmt.Errorf("(WorkerConfirmed) Expected 1 Worker, got %d", len(workers))
		}
		if !workers[0].Confirmed {
			return fmt.Errorf("(WorkerConfirmed) Expected %t, got %t (Worker.Confirmed)",
				true, workers[0].Confirmed)
		}
	}
	// Check that a worker is deleted after a WorkerRemoved event.
	if err := monitorDWH.onWorkerRemoved(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerRemoved failed: %v", err)
	}
	if workers, _, err := monitorDWH.storage.GetWorkers(newSimpleConn(monitorDWH.db), &pb.WorkersRequest{}); err != nil {
		return fmt.Errorf("getWorkers failed: %v", err)
	} else {
		if len(workers) != 0 {
			return fmt.Errorf("(WorkerRemoved) Expected 0 Workers, got %d", len(workers))
		}
	}
	return nil
}

func testBlacklistAddedRemoved() error {
	// Check that a Blacklist entry is added after AddedToBlacklist event.
	if err := monitorDWH.onAddedToBlacklist(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onAddedToBlacklist failed: %v", err)
	}
	if blacklistReply, err := monitorDWH.storage.GetBlacklist(
		newSimpleConn(monitorDWH.db), &pb.BlacklistRequest{UserID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return fmt.Errorf("getBlacklist failed: %v", err)
	} else {
		if blacklistReply.OwnerID.Unwrap().Hex() != common.HexToAddress("0xC").Hex() {
			return fmt.Errorf("(AddedToBlacklist) Expected %s, got %s (BlacklistReply.AdderID)",
				common.HexToAddress("0xC").Hex(), blacklistReply.OwnerID)
		}
	}
	// Check that a Blacklist entry is deleted after RemovedFromBlacklist event.
	if err := monitorDWH.onRemovedFromBlacklist(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onRemovedFromBlacklist failed: %v", err)
	}
	if repl, err := monitorDWH.storage.GetBlacklist(
		newSimpleConn(monitorDWH.db), &pb.BlacklistRequest{UserID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return fmt.Errorf("getBlacklist (2) failed: %v", err)
	} else {
		if len(repl.Addresses) > 0 {
			return fmt.Errorf("getBlacklist returned a blacklist that should have been deleted: %+v", repl.Addresses)
		}
	}
	return nil
}

func getDealChangeRequest(w *DWH, changeRequestID *pb.BigInt) (*pb.DealChangeRequest, error) {
	rows, err := w.storage.builder().Select("*").From("DealChangeRequests").
		Where("Id = ?", changeRequestID.Unwrap().String()).RunWith(w.db).Query()
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return w.storage.decodeDealChangeRequest(rows)
}

func getCertificates(w *DWH) ([]*pb.Certificate, error) {
	rows, err := w.storage.builder().Select("*").From("Certificates").RunWith(w.db).Query()
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.Certificate
	for rows.Next() {
		if certificate, err := w.storage.decodeCertificate(rows); err != nil {
			return nil, fmt.Errorf("failed to decodeCertificate: %v", err)
		} else {
			out = append(out, certificate)
		}
	}

	return out, nil
}

func setupTestDB(w *DWH) error {
	if err := w.setupDB(); err != nil {
		return err
	}

	var certs = []*pb.Certificate{
		{OwnerID: pb.NewEthAddress(common.HexToAddress("0xBB")), Value: []byte("Consumer"), Attribute: CertificateName},
	}
	byteCerts, _ := json.Marshal(certs)
	for i := 0; i < 2; i++ {
		insertDeal, args, _ := w.storage.builder().Insert("Deals").
			Columns(w.storage.tablesInfo.DealColumns...).
			Values(
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
			).ToSql()

		_, err := w.db.Exec(insertDeal, args...)
		if err != nil {
			return fmt.Errorf("failed to insertDeal: %v", err)
		}

		// Create 10 ASK orders.
		_, err = w.storage.builder().Insert("Orders").
			Columns(w.storage.tablesInfo.OrderColumns...).Values(
			fmt.Sprintf("2020%d", i),
			common.HexToAddress(fmt.Sprintf("0x9%d", i)).Hex(), // Master
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(pb.OrderType_ASK),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xA").Hex(), // AuthorID
			common.Address{}.Hex(),           // CounterpartyID
			10010+i,
			pb.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			7, // Netflags
			uint64(pb.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},          // Tag
			fmt.Sprintf("3001%d", i), // FrozenSum
			uint64(pb.IdentityLevel_REGISTERED),
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
		).RunWith(w.db).Exec()
		if err != nil {
			return err
		}

		// Create 10 BID orders.
		_, err = w.storage.builder().Insert("Orders").
			Columns(w.storage.tablesInfo.OrderColumns...).Values(
			fmt.Sprintf("3030%d", i),
			common.HexToAddress(fmt.Sprintf("0x9%d", i)).Hex(), // Master
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(pb.OrderType_BID),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xB").Hex(), // AuthorID
			common.Address{}.Hex(),           // CounterpartyID
			10010-i,                          // Duration
			pb.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			5, // Netflags
			uint64(pb.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},                     // Tag
			fmt.Sprintf("3001%d", i),            // FrozenSum
			uint64(pb.IdentityLevel_REGISTERED), // CreatorIdentityLevel
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
		).RunWith(w.db).Exec()
		if err != nil {
			return err
		}

		_, err = w.storage.builder().Insert("DealChangeRequests").
			Values(fmt.Sprintf("5050%d", i), 0, 0, 0, 0, 0, "40400").RunWith(w.db).Exec()
		if err != nil {
			return err
		}

		var identityLevel int
		if (i % 2) == 0 {
			identityLevel = 0
		} else {
			identityLevel = 1
		}
		_, err = w.storage.builder().Insert("Profiles").Columns(w.storage.tablesInfo.ProfileColumns[1:]...).Values(
			common.HexToAddress(fmt.Sprintf("0x2%d", i)).Hex(),
			identityLevel,
			"sortedProfile",
			"",
			0,
			0,
			[]byte{},
			0,
			0,
		).RunWith(w.db).Exec()
		if err != nil {
			return err
		}
	}

	// Create a couple of profiles for TestDWH_monitor entities.
	_, err := w.storage.builder().Insert("Profiles").Columns(w.storage.tablesInfo.ProfileColumns[1:]...).Values(
		fmt.Sprintf(common.HexToAddress("0xBB").Hex()),
		3,
		"Consumer",
		"",
		0,
		0,
		byteCerts,
		10,
		10,
	).RunWith(w.db).Exec()
	if err != nil {
		return err
	}
	_, err = w.storage.builder().Insert("Profiles").Columns(w.storage.tablesInfo.ProfileColumns[1:]...).Values(
		fmt.Sprintf(common.HexToAddress("0xAA").Hex()),
		3,
		"Supplier",
		"",
		0,
		0,
		byteCerts,
		10,
		10,
	).RunWith(w.db).Exec()
	if err != nil {
		return err
	}
	// Blacklist 0xBB for 0xE for TestDWH_GetProfiles.
	_, err = w.storage.builder().Insert("Blacklists").Values(
		common.HexToAddress("0xE").Hex(),
		common.HexToAddress("0xBB").Hex(),
	).RunWith(w.db).Exec()
	if err != nil {
		return err
	}

	// Add a BID order that will be matched by any of the ASK orders added above and
	// blacklist this BID order's Author for the author of all ASK orders. Then in
	// TestDWH_GetMatchingOrders we shouldn't get this order.
	_, err = w.storage.builder().Insert("Blacklists").Values(
		common.HexToAddress("0xA").Hex(),
		common.HexToAddress("0xCC").Hex(),
	).RunWith(w.db).Exec()
	if err != nil {
		return err
	}
	_, err = w.storage.builder().Insert("Orders").
		Columns(w.storage.tablesInfo.OrderColumns...).Values(
		fmt.Sprintf("3050%d", 0),
		common.HexToAddress(fmt.Sprintf("0x9%d", 0)).Hex(), // Master
		12345, // CreatedTS
		fmt.Sprintf("1010%d", 0),
		uint64(pb.OrderType_BID),
		uint64(pb.OrderStatus_ORDER_ACTIVE),
		common.HexToAddress("0xCC").Hex(), // AuthorID
		common.HexToAddress("0xA").Hex(),  // CounterpartyID
		10, // Duration
		pb.NewBigIntFromInt(30010+int64(0)).PaddedString(), // Price
		5, // Netflags
		uint64(pb.IdentityLevel_ANONYMOUS),
		fmt.Sprintf("blacklist_%d", 0),
		[]byte{1, 2, 3},                     // Tag
		fmt.Sprintf("3001%d", 0),            // FrozenSum
		uint64(pb.IdentityLevel_REGISTERED), // CreatorIdentityLevel
		"CreatorName",
		"CreatorCountry",
		byteCerts, // CreatorCertificates
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
	).RunWith(w.db).Exec()
	if err != nil {
		return err
	}

	return nil
}
