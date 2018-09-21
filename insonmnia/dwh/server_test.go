package dwh

import (
	"math/big"
	"testing"

	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/proto"
)

func testGetStats(t *testing.T) {
	stats, err := testDWH.storage.getStats(newSimpleConn(testDWH.db))
	if err != nil {
		t.Errorf("Failed to get stats: %v", err)
		return
	}

	if stats.CurrentDeals != 2 {
		t.Errorf("Unexpected stats: %+v", stats)
		return
	}
}

func testGetDeals(t *testing.T) {
	var (
		byAddress            = common.HexToAddress("0x11")
		byMinDuration uint64 = 10011
		byMinPrice           = big.NewInt(20011)
	)

	// Test TEXT columns.
	{
		request := &sonm.DealsRequest{SupplierID: sonm.NewEthAddress(byAddress)}
		reply, err := testDWH.GetDeals(testDWH.ctx, request)
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
		request := &sonm.DealsRequest{
			Duration: &sonm.MaxMinUint64{Min: byMinDuration},
		}
		reply, err := testDWH.GetDeals(testDWH.ctx, request)
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
		request := &sonm.DealsRequest{
			Price: &sonm.MaxMinBig{Min: sonm.NewBigInt(byMinPrice)},
		}
		reply, err := testDWH.GetDeals(testDWH.ctx, request)
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

func testGetDealDetails(t *testing.T) {
	var (
		byDealID = big.NewInt(40400)
	)
	reply, err := testDWH.storage.GetDealByID(newSimpleConn(testDWH.db), byDealID)
	if err != nil {
		t.Error(err)
		return
	}
	deal := reply.GetDeal()
	if deal.Id.Unwrap().Cmp(byDealID) != 0 {
		t.Errorf("Expected %d, got %d (Id)", byDealID.Int64(), deal.Id.Unwrap().Int64())
	}
}

func testGetOrders(t *testing.T) {
	var (
		byDealID              = big.NewInt(10101)
		byMinDuration  uint64 = 10011
		byMinBenchmark uint64 = 11
		byMinPrice            = big.NewInt(20011)
	)
	// Test TEXT columns.
	{
		request := &sonm.OrdersRequest{DealID: sonm.NewBigInt(byDealID)}
		orders, _, err := testDWH.storage.GetOrders(newSimpleConn(testDWH.db), request)
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
		request := &sonm.OrdersRequest{
			Duration:   &sonm.MaxMinUint64{Min: byMinDuration},
			Benchmarks: map[uint64]*sonm.MaxMinUint64{0: {Min: byMinBenchmark}},
		}
		orders, _, err := testDWH.storage.GetOrders(newSimpleConn(testDWH.db), request)
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
		request := &sonm.OrdersRequest{
			Type:  sonm.OrderType_ASK,
			Price: &sonm.MaxMinBig{Min: sonm.NewBigInt(byMinPrice)},
		}
		orders, _, err := testDWH.storage.GetOrders(newSimpleConn(testDWH.db), request)
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

func testGetMatchingOrders(t *testing.T) {
	var byID = big.NewInt(20201)
	request := &sonm.MatchingOrdersRequest{Id: sonm.NewBigInt(byID)}
	orders, _, err := testDWH.storage.GetMatchingOrders(newSimpleConn(testDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: GetMatchingOrders failed: %s", request, err)
		return
	}
	if len(orders) != 1 {
		t.Errorf("Request `%+v` failed: Expected 5 orders in reply, got %d", request, len(orders))
		return
	}
}

func testGetOrderDetails(t *testing.T) {
	var byID = big.NewInt(20201)
	order, err := testDWH.storage.GetOrderByID(newSimpleConn(testDWH.db), byID)
	if err != nil {
		t.Error(err)
		return
	}
	reply := order.GetOrder()
	if reply.Id.Unwrap().Cmp(byID) != 0 {
		t.Errorf("Request `%d` failed: Expected %d, got %d (Id)", byID.Int64(), byID.Int64(), reply.Id.Unwrap().Int64())
	}
}

func testGetDealChangeRequests(t *testing.T) {
	changeRequests, err := testDWH.GetChangeRequests(context.Background(), &sonm.ChangeRequestsRequest{
		DealID: sonm.NewBigIntFromInt(40400),
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(changeRequests.Requests) != 2 {
		t.Errorf("Request `%d` failed: Expected %d DealChangeRequests, got %d", 404000, 2, len(changeRequests.Requests))
		return
	}
}

func testGetProfiles(t *testing.T) {
	var request = &sonm.ProfilesRequest{
		Identifier: "sortedProfile",
		Sortings: []*sonm.SortingOption{
			{
				Field: "UserID",
				Order: sonm.SortingOrder_Asc,
			},
		},
	}
	profiles, _, err := testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), request)
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
	request = &sonm.ProfilesRequest{
		Identifier: "sortedProfile",
		Sortings: []*sonm.SortingOption{
			{
				Field: "UserID",
				Order: sonm.SortingOrder_Desc,
			},
		},
	}
	profiles, _, err = testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), request)
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
	request = &sonm.ProfilesRequest{
		Identifier: "sortedProfile",
		Sortings: []*sonm.SortingOption{
			{
				Field: "IdentityLevel",
				Order: sonm.SortingOrder_Asc,
			},
			{
				Field: "UserID",
				Order: sonm.SortingOrder_Asc,
			},
		},
	}
	profiles, _, err = testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), request)
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
	request = &sonm.ProfilesRequest{
		BlacklistQuery: &sonm.BlacklistQuery{
			OwnerID: sonm.NewEthAddress(common.HexToAddress("0xE")),
			Option:  sonm.BlacklistOption_OnlyMatching,
		},
	}
	profiles, _, err = testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), request)
	if err != nil {
		t.Errorf("Request `%+v` failed: %s", request, err)
		return
	}
	if len(profiles) != 1 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 1, len(profiles))
		return
	}
	request = &sonm.ProfilesRequest{
		BlacklistQuery: &sonm.BlacklistQuery{
			OwnerID: sonm.NewEthAddress(common.HexToAddress("0xE")),
			Option:  sonm.BlacklistOption_WithoutMatching,
		},
	}
	profiles, _, err = testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), request)
	if err != nil {
		t.Error(err)
		return
	}
	if len(profiles) != 3 {
		t.Errorf("Request `%+v` failed: Expected %d Profiles, got %d", request, 3, len(profiles))
		return
	}
	profiles, _, err = testDWH.storage.GetProfiles(newSimpleConn(testDWH.db), &sonm.ProfilesRequest{
		BlacklistQuery: &sonm.BlacklistQuery{
			OwnerID: sonm.NewEthAddress(common.HexToAddress("0xE")),
			Option:  sonm.BlacklistOption_IncludeAndMark,
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
