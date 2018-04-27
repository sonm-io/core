package dwh

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	bch "github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	w := &DWH{
		ctx:      ctx,
		cfg:      cfg,
		logger:   log.GetLogger(ctx),
		commands: sqliteCommands,
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
			SupplierID: "supplier_5",
		}
		reply, err := globalDWH.getDeals(context.Background(), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Deals) != 1 {
			t.Errorf("Expected 1 deal in reply, got %d", len(reply.Deals))
			return
		}

		if reply.Deals[0].GetDeal().SupplierID != "supplier_5" {
			t.Errorf("Request `%+v` failed, expected %d, got %d (SupplierID)",
				request, 10015, reply.Deals[0].GetDeal().Duration)
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
		reply, err := globalDWH.getDeals(context.Background(), request)

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
		reply, err := globalDWH.getDeals(context.Background(), request)

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

	deal, err := globalDWH.getDealDetails(context.Background(), &pb.ID{Id: "id_5"})
	if err != nil {
		t.Error(err)
		return
	}

	reply := deal.GetDeal()
	if reply.Id != "id_5" {
		t.Errorf("Expected %s, got %s (Id)", "id_5", reply.Id)
	}
	if reply.SupplierID != "supplier_5" {
		t.Errorf("Expected %s, got %s (SupplierID)", "supplier_5", reply.SupplierID)
	}
	if reply.ConsumerID != "consumer_5" {
		t.Errorf("Expected %s, got %s (ConsumerID)", "consumer_5", reply.ConsumerID)
	}
	if reply.MasterID != "master_5" {
		t.Errorf("Expected %s, got %s (MasterID)", "master_5", reply.MasterID)
	}
	if reply.AskID != "ask_id_5" {
		t.Errorf("Expected %s, got %s (AskID)", "ask_id_5", reply.AskID)
	}
	if reply.BidID != "bid_id_5" {
		t.Errorf("Expected %s, got %s (BidID)", "bid_id_5", reply.AskID)
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
			DealID: "deal_id_5",
		}
		reply, err := globalDWH.getOrders(context.Background(), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Orders) != 2 {
			t.Errorf("Expected 2 orders in reply, got %d", len(reply.Orders))
			return
		}

		if reply.Orders[0].GetOrder().AuthorID != "ask_author" {
			t.Errorf("Request `%+v` failed, expected %s, got %s (AuthorID)",
				request, "ask_author", reply.Orders[0].GetOrder().AuthorID)
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
			Benchmarks: &pb.DWHBenchmarkConditions{
				CPUSysbenchMulti: &pb.MaxMinUint64{
					Min: 15,
				},
			},
		}
		reply, err := globalDWH.getOrders(context.Background(), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Orders) != 5 {
			t.Errorf("Expected 5 orders in reply, got %d", len(reply.Orders))
			return
		}

		if reply.Orders[0].GetOrder().Duration != uint64(10015) {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Duration)",
				request, 10015, reply.Orders[0].GetOrder().Duration)
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
		reply, err := globalDWH.getOrders(context.Background(), request)

		if err != nil {
			t.Errorf("Request `%+v` failed: %s", request, err)
			return
		}

		if len(reply.Orders) != 5 {
			t.Errorf("Expected 5 orders in reply, got %d", len(reply.Orders))
			return
		}

		if reply.Orders[0].GetOrder().Price.Unwrap().String() != "20015" {
			t.Errorf("Request `%+v` failed, expected %d, got %d (Price)",
				request, 10015, reply.Orders[0].GetOrder().Duration)
			return
		}
	}
}

func TestDWH_GetMatchingOrders(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	request := &pb.MatchingOrdersRequest{
		Id: &pb.ID{Id: "ask_id_5"},
	}
	reply, err := globalDWH.getMatchingOrders(context.Background(), request)
	if err != nil {
		t.Errorf("GetMatchingOrders failed: %s", err)
		return
	}

	if len(reply.Orders) != 5 {
		t.Errorf("Expected 5 orders in reply, got %d", len(reply.Orders))
		return
	}
}

func TestDWH_GetOrderDetails(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	order, err := globalDWH.getOrderDetails(context.Background(), &pb.ID{Id: "ask_id_5"})
	if err != nil {
		t.Error(err)
		return
	}

	reply := order.GetOrder()
	if reply.Id != "ask_id_5" {
		t.Errorf("Expected %s, got %s (Id)", "ask_id_5", reply.Id)
	}
	if reply.DealID != "deal_id_5" {
		t.Errorf("Expected %s, got %s (DealID)", "deal_id_5", reply.DealID)
	}
	if reply.OrderType != 2 {
		t.Errorf("Expected %d, got %d (Type)", 2, reply.OrderType)
	}
	if reply.AuthorID != "ask_author" {
		t.Errorf("Expected %s, got %s (AuthorID)", "ask_author", reply.AuthorID)
	}
	if reply.CounterpartyID != "bid_author" {
		t.Errorf("Expected %s, got %s (CounterpartyID)", "bid_author", reply.CounterpartyID)
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

	reply, err := globalDWH.getDealChangeRequests(context.Background(), &pb.ID{Id: "id_0"})
	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Requests) != 10 {
		t.Errorf("Expected %d DealChangeRequests, got %d", 10, len(reply.Requests))
		return
	}
}

func TestDWH_GetProfiles(t *testing.T) {
	globalDWH.mu.Lock()
	defer globalDWH.mu.Unlock()

	reply, err := globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
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

	if len(reply.Profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(reply.Profiles))
		return
	}

	if reply.Profiles[0].UserID != "test_profile_0" {
		t.Errorf("Expected %s, got %s (Profile.UserID)", "test_profile_0", reply.Profiles[0].UserID)
		return
	}

	reply, err = globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
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

	if len(reply.Profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(reply.Profiles))
		return
	}

	if reply.Profiles[0].UserID != "test_profile_9" {
		t.Errorf("Expected %s, got %s (Profile.UserID)", "test_profile_9", reply.Profiles[0].UserID)
		return
	}

	reply, err = globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
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

	if len(reply.Profiles) != 10 {
		t.Errorf("Expected %d Profiles, got %d", 10, len(reply.Profiles))
		return
	}

	if reply.Profiles[0].UserID != "test_profile_0" {
		t.Errorf("Expected %s, got %s (Profile.UserID)", "test_profile_0", reply.Profiles[0].UserID)
		return
	}
	if reply.Profiles[4].UserID != "test_profile_8" {
		t.Errorf("Expected %s, got %s (Profile.UserID)", "test_profile_8", reply.Profiles[4].UserID)
		return
	}

	reply, err = globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: "blacklisting_user",
			Option:  pb.BlacklistOption_OnlyMatching,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Profiles) != 1 {
		t.Errorf("Expected %d Profiles, got %d", 1, len(reply.Profiles))
		return
	}

	reply, err = globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: "blacklisting_user",
			Option:  pb.BlacklistOption_WithoutMatching,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Profiles) != 11 {
		t.Errorf("Expected %d Profiles, got %d", 11, len(reply.Profiles))
		return
	}

	reply, err = globalDWH.getProfiles(globalDWH.ctx, &pb.ProfilesRequest{
		BlacklistQuery: &pb.BlacklistQuery{
			OwnerID: "blacklisting_user",
			Option:  pb.BlacklistOption_IncludeAndMark,
		},
	})
	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Profiles) != 12 {
		t.Errorf("Expected %d Profiles, got %d", 12, len(reply.Profiles))
		return
	}

	var foundMarkedProfile bool
	for _, profile := range reply.Profiles {
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
		commonID             = big.NewInt(0xDEADBEEF)
		commonEventTS uint64 = 5
	)

	benchmarks, err := pb.NewBenchmarks([]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	require.NoError(t, err)

	deal := &pb.Deal{
		Id:             commonID.String(),
		Benchmarks:     benchmarks,
		SupplierID:     "supplier_id",
		ConsumerID:     "consumer_id",
		MasterID:       "master_id",
		AskID:          "ask_id_5",
		BidID:          "bid_id_5",
		Duration:       10020,
		Price:          pb.NewBigInt(big.NewInt(20010)),
		StartTime:      &pb.Timestamp{Seconds: 30010},
		EndTime:        &pb.Timestamp{Seconds: 40010},
		Status:         pb.DealStatus_DEAL_ACCEPTED,
		BlockedBalance: pb.NewBigInt(big.NewInt(50010)),
		TotalPayout:    pb.NewBigInt(big.NewInt(0)),
		LastBillTS:     &pb.Timestamp{Seconds: 70010},
	}
	mockBlock.EXPECT().GetDealInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(deal, nil)

	order := &pb.Order{
		Id:             commonID.String(),
		DealID:         "",
		OrderType:      pb.OrderType_ASK,
		OrderStatus:    pb.OrderStatus_ORDER_ACTIVE,
		AuthorID:       strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		CounterpartyID: "counterparty_id",
		Duration:       10020,
		Price:          pb.NewBigInt(big.NewInt(20010)),
		Netflags:       7,
		IdentityLevel:  pb.IdentityLevel_ANONYMOUS,
		Blacklist:      "blacklist",
		Tag:            []byte{0, 1},
		Benchmarks:     benchmarks,
		FrozenSum:      pb.NewBigInt(big.NewInt(30010)),
	}
	mockBlock.EXPECT().GetOrderInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(order, nil)

	changeRequest := &pb.DealChangeRequest{
		Id:          "0",
		DealID:      commonID.String(),
		RequestType: pb.OrderType_ASK,
		Duration:    10020,
		Price:       pb.NewBigInt(big.NewInt(20010)),
		Status:      pb.ChangeRequestStatus_REQUEST_CREATED,
	}
	mockBlock.EXPECT().GetDealChangeRequestInfo(gomock.Any(), gomock.Any()).AnyTimes().Return(changeRequest, nil)

	validator := &pb.Validator{
		Id:    strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		Level: 3,
	}
	mockBlock.EXPECT().GetValidator(gomock.Any(), gomock.Any()).AnyTimes().Return(validator, nil)

	certificate := &pb.Certificate{
		ValidatorID:   strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		OwnerID:       strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		Attribute:     CertificateName,
		IdentityLevel: 1,
		Value:         []byte("User Name"),
	}
	mockBlock.EXPECT().GetCertificate(gomock.Any(), gomock.Any()).AnyTimes().Return(
		certificate, nil)

	monitorDWH.blockchain = mockBlock

	// Test onOrderPlaced event handling.
	if err := monitorDWH.onOrderPlaced(commonEventTS, commonID); err != nil {
		t.Error(err)
		return
	}
	if order, err := monitorDWH.getOrderDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetOrderDetails: %s", err)
		return
	} else {
		if order.GetOrder().Duration != 10020 {
			t.Errorf("Expected %d, got %d (Order.Duration)", 10020, order.GetOrder().Duration)
		}
	}

	// Test onDealOpened event handling.
	if err := monitorDWH.onDealOpened(commonID); err != nil {
		t.Error(err)
		return
	}
	// Firstly, check that a deal was created.
	if deal, err := monitorDWH.getDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if deal.GetDeal().Duration != 10020 {
			t.Errorf("Expected %d, got %d (Deal.Duration)", 10020, deal.GetDeal().Duration)
		}
	}
	// Secondly, check that a DealCondition was created.
	if dealConditionsReply, err := monitorDWH.getDealConditions(
		monitorDWH.ctx, &pb.DealConditionsRequest{DealID: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealConditions: %s", err)
		return
	} else {
		if dealConditionsReply.Conditions[0].Duration != 10020 {
			t.Errorf("Expected %d, got %d (DealCondition.Duration)", 10020, deal.Duration)
			return
		}
	}

	// Test that a Validator entry is added after ValidatorCreated event.
	if err := monitorDWH.onValidatorCreated(common.HexToAddress(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"))); err != nil {
		t.Error(err)
		return
	}
	if validatorsReply, err := monitorDWH.getValidators(monitorDWH.ctx, &pb.ValidatorsRequest{}); err != nil {
		t.Errorf("Failed to GetValidators: %s", err)
		return
	} else {
		if len(validatorsReply.Validators) != 1 {
			t.Errorf("(ValidatorCreated) Expected 1 Validator, got %d", len(validatorsReply.Validators))
			return
		}
		if validatorsReply.Validators[0].Level != 3 {
			t.Errorf("(ValidatorCreated) Expected %d, got %d (Validator.Level)",
				3, validatorsReply.Validators[0].Level)
		}
	}

	validator.Level = 0
	// Test that a Validator entry is updated after ValidatorDeleted event.
	if err := monitorDWH.onValidatorDeleted(common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD")); err != nil {
		t.Error(err)
		return
	}
	if validatorsReply, err := monitorDWH.getValidators(monitorDWH.ctx, &pb.ValidatorsRequest{}); err != nil {
		t.Errorf("Failed to GetValidators: %s", err)
		return
	} else {
		if len(validatorsReply.Validators) != 1 {
			t.Errorf("(ValidatorDeleted) Expected 1 Validator, got %d", len(validatorsReply.Validators))
			return
		}
		if validatorsReply.Validators[0].Level != 0 {
			t.Errorf("(ValidatorDeleted) Expected %d, got %d (Validator.Level)",
				0, validatorsReply.Validators[0].Level)
		}
	}

	// Test that a Certificate entry is updated after CertificateCreated event.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		t.Error(err)
		return
	}
	if certificateAttrs, err := getCertificates(monitorDWH); err != nil {
		t.Errorf("Failed to getValidators: %s", err)
		return
	} else {
		if len(certificateAttrs) != 1 {
			t.Errorf("(CertificateCreated) Expected 1 Certificate, got %d",
				len(certificateAttrs))
			return
		}
		if string(certificateAttrs[0].Value) != "User Name" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Certificate.Value)",
				"User Name", certificateAttrs[0].Value)
		}
	}
	if profilesReply, err := monitorDWH.getProfiles(monitorDWH.ctx, &pb.ProfilesRequest{}); err != nil {
		t.Errorf("Failed to getProfiles: %s", err)
		return
	} else {
		if len(profilesReply.Profiles) != 13 {
			t.Errorf("(CertificateCreated) Expected 1 Profile, got %d",
				len(profilesReply.Profiles))
			return
		}
		if profilesReply.Profiles[12].Name != "User Name" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"User Name", profilesReply.Profiles[12].Name)
		}
	}

	certificate.Attribute = CertificateCountry
	certificate.Value = []byte("Country")
	// Test that a  Profile entry is updated after CertificateCreated event.
	if err := monitorDWH.onCertificateCreated(commonID); err != nil {
		t.Error(err)
		return
	}
	if profilesReply, err := monitorDWH.getProfiles(monitorDWH.ctx, &pb.ProfilesRequest{}); err != nil {
		t.Errorf("Failed to getProfiles: %s", err)
		return
	} else {
		if len(profilesReply.Profiles) != 13 {
			t.Errorf("(CertificateCreated) Expected 1 Profile, got %d",
				len(profilesReply.Profiles))
			return
		}

		profiles := profilesReply.Profiles

		if profiles[12].Country != "Country" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Country)",
				"Country", profiles[12].Name)
		}
		if profiles[12].Name != "User Name" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"Name", profiles[12].Name)
		}

		var certificates []*pb.Certificate
		if err := json.Unmarshal(profiles[12].Certificates, &certificates); err != nil {
			t.Errorf("(CertificateCreated) Failed to unmarshal Profile.Certificates: %s", err)
			return
		} else {
			if len(certificates) != 2 {
				t.Errorf("(CertificateCreated) Expected 2 Certificates, got %d",
					len(certificates))
				return
			}
		}
	}
	// Check that profile updates resulted in orders updates.
	dwhOrder, err := monitorDWH.getOrderDetails(context.Background(), &pb.ID{Id: commonID.String()})
	if err != nil {
		t.Errorf("failed to getOrderDetails (`%s`): %s", commonID.String(), err)
		return
	}
	if dwhOrder.CreatorIdentityLevel != 3 {
		t.Errorf("(CertificateCreated) Expected %d, got %d (Order.CreatorIdentityLevel)",
			2, dwhOrder.CreatorIdentityLevel)
		return
	}

	// Check that profile updates resulted in orders updates.
	if deal, err := monitorDWH.getDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if len(deal.SupplierCertificates) == 0 {
			t.Errorf("Expected some SupplierCertificates, got nothing")
		}
	}

	// Test that if order is updated, it is deleted.
	if err := monitorDWH.onOrderUpdated(commonID); err != nil {
		t.Error(err)
		return
	}
	if _, err := monitorDWH.getOrderDetails(context.Background(), &pb.ID{Id: commonID.String()}); err == nil {
		t.Error("GetOrderDetails returned an order that should have been deleted")
		return
	}

	deal.Duration += 1
	// Test onDealUpdated event handling.
	if err := monitorDWH.onDealUpdated(commonID); err != nil {
		t.Error(err)
		return
	}
	if deal, err := monitorDWH.getDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if deal.GetDeal().Duration != 10021 {
			t.Errorf("Expected %d, got %d (Deal.Duration)", 10021, deal.GetDeal().Duration)
		}
	}

	// Test creating an ASK DealChangeRequest.
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(0)); err != nil {
		t.Error(err)
		return
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		t.Errorf("Failed to getDealChangeRequest: %s", err)
		return
	} else {
		if changeRequest.Duration != 10020 {
			t.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10020, changeRequest.Duration)
		}
	}

	// Test that after a second ASK DealChangeRequest was created, the new one was kept and the old one was deleted.
	changeRequest.Id = "1"
	changeRequest.Duration = 10021
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(1)); err != nil {
		t.Error(err)
		return
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		t.Errorf("Failed to getDealChangeRequest: %s", err)
		return
	} else {
		if changeRequest.Duration != 10021 {
			t.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10021, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(monitorDWH, "0"); err == nil {
		t.Error("getDealChangeRequest returned a DealChangeRequest that should have been deleted")
		return
	}

	// Test that when a BID DealChangeRequest was created, it was kept (and nothing was deleted).
	changeRequest.Id = "2"
	changeRequest.Duration = 10022
	changeRequest.RequestType = pb.OrderType_BID
	if err := monitorDWH.onDealChangeRequestSent(commonEventTS, big.NewInt(2)); err != nil {
		t.Error(err)
		return
	}
	if changeRequest, err := getDealChangeRequest(monitorDWH, changeRequest.Id); err != nil {
		t.Errorf("Failed to getDealChangeRequest: %s", err)
		return
	} else {
		if changeRequest.Duration != 10022 {
			t.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10022, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(monitorDWH, "1"); err != nil {
		t.Errorf("DealChangeRequest of type ASK was deleted after a BID DealChangeRequest creation: %s", err)
		return
	}

	// Test that when a DealChangeRequest is updated to any status but REJECTED, it is deleted.
	changeRequest.Id = "1"
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_ACCEPTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(1)); err != nil {
		t.Error(err)
		return
	}
	if _, err := getDealChangeRequest(monitorDWH, "1"); err == nil {
		t.Error("DealChangeRequest which status was changed to ACCEPTED was not deleted")
		return
	}
	// Also test that a new DealCondition was created, and the old one was updated.
	if dealConditionsReply, err := monitorDWH.getDealConditions(
		monitorDWH.ctx, &pb.DealConditionsRequest{DealID: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealConditions: %s", err)
		return
	} else {
		if len(dealConditionsReply.Conditions) != 2 {
			t.Errorf("Expected 2 DealConditions, got %d", len(dealConditionsReply.Conditions))
			return
		}
		conditions := dealConditionsReply.Conditions
		if conditions[1].EndTime.Seconds != 5 {
			t.Errorf("Expected %d, got %d (DealCondition.EndTime)", 5, conditions[0].EndTime.Seconds)
			return
		}
		if conditions[0].StartTime.Seconds != 5 {
			t.Errorf("Expected %d, got %d (DealCondition.StartTime)", 5, conditions[1].StartTime.Seconds)
			return
		}
	}

	// Test that when a DealChangeRequest is updated to REJECTED, it is kept.
	changeRequest.Id = "2"
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_REJECTED
	if err := monitorDWH.onDealChangeRequestUpdated(commonEventTS, big.NewInt(2)); err != nil {
		t.Error(err)
		return
	}
	if _, err := getDealChangeRequest(monitorDWH, "2"); err != nil {
		t.Error("DealChangeRequest which status was changed to REJECTED was deleted")
		return
	}

	// Test that after a Billed event last DealCondition.Payout is updated.
	if err := monitorDWH.onBilled(commonEventTS, commonID, big.NewInt(10)); err != nil {
		t.Error(err)
		return
	}
	if dealConditionsReply, err := monitorDWH.getDealConditions(
		monitorDWH.ctx, &pb.DealConditionsRequest{DealID: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if len(dealConditionsReply.Conditions) != 2 {
			t.Errorf("(Billed) Expected 2 DealConditions, got %d", len(dealConditionsReply.Conditions))
			return
		}
		conditions := dealConditionsReply.Conditions
		if conditions[0].TotalPayout.Unwrap().String() != "10" {
			t.Errorf("(Billed) Expected %s, got %s (DealCondition.TotalPayout)",
				"10", conditions[0].TotalPayout.Unwrap().String())
		}
	}
	if dealPayments, err := getDealPayments(monitorDWH); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if len(dealPayments) != 1 {
			t.Errorf("(Billed) Expected 1 DealPayment, got %d", len(dealPayments))
			return
		}
		if !strings.HasSuffix(dealPayments[0].PaidAmount, "10") {
			t.Errorf("(Billed) Expected %s, got %s (DealPayment.PaidAmount)",
				"10", dealPayments[0].PaidAmount)
		}
	}

	// Test that when a Deal's status is updated to CLOSED, Deal and its DealConditions are deleted.
	deal.Status = pb.DealStatus_DEAL_CLOSED
	// Test onDealUpdated event handling.
	if err := monitorDWH.onDealUpdated(commonID); err != nil {
		t.Error(err)
		return
	}
	if _, err := monitorDWH.getDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err == nil {
		t.Errorf("Deal was not deleted after status changing to CLOSED")
		return
	}
	if dealConditions, err := monitorDWH.getDealConditions(
		monitorDWH.ctx, &pb.DealConditionsRequest{DealID: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealConditions: %s", err)
		return
	} else {
		if len(dealConditions.Conditions) != 0 {
			t.Errorf("(DealUpdated) Expected 0 DealConditions, got %d", len(dealConditions.Conditions))
			return
		}
	}

	if profile, err := monitorDWH.getProfileInfo(monitorDWH.ctx, &pb.ID{Id: "consumer_id"}, true); err != nil {
		t.Errorf("Failed to GetProfileInfo: %s", err)
		return
	} else {
		if profile.ActiveBids != 9 {
			t.Errorf("(DealUpdated) Expected 9 ActiveBids, got %d", profile.ActiveBids)
			return
		}
	}

	// Test that a worker is added after a WorkerAnnounced event.
	if err := monitorDWH.onWorkerAnnounced(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE")); err != nil {
		t.Error(err)
		return
	}
	if workersReply, err := monitorDWH.getWorkers(monitorDWH.ctx, &pb.WorkersRequest{}); err != nil {
		t.Errorf("Failed to GetWorkers: %s", err)
		return
	} else {
		if len(workersReply.Workers) != 1 {
			t.Errorf("(WorkerAnnounced) Expected 1 Worker, got %d", len(workersReply.Workers))
			return
		}
		if workersReply.Workers[0].Confirmed {
			t.Errorf("(WorkerAnnounced) Expected %t, got %t (Worker.Confirmed)",
				false, workersReply.Workers[0].Confirmed)
		}
	}
	// Test that a worker is confirmed after a WorkerConfirmed event.
	if err := monitorDWH.onWorkerConfirmed(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE")); err != nil {
		t.Error(err)
		return
	}
	if workersReply, err := monitorDWH.getWorkers(monitorDWH.ctx, &pb.WorkersRequest{}); err != nil {
		t.Errorf("Failed to GetWorkers: %s", err)
		return
	} else {
		if len(workersReply.Workers) != 1 {
			t.Errorf("(WorkerConfirmed) Expected 1 Worker, got %d", len(workersReply.Workers))
			return
		}
		if !workersReply.Workers[0].Confirmed {
			t.Errorf("(WorkerConfirmed) Expected %t, got %t (Worker.Confirmed)",
				true, workersReply.Workers[0].Confirmed)
		}
	}
	// Test that a worker is deleted after a WorkerRemoved event.
	if err := monitorDWH.onWorkerRemoved(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE")); err != nil {
		t.Error(err)
		return
	}
	if workersReply, err := monitorDWH.getWorkers(monitorDWH.ctx, &pb.WorkersRequest{}); err != nil {
		t.Errorf("Failed to getWorkers: %s", err)
		return
	} else {
		if len(workersReply.Workers) != 0 {
			t.Errorf("(WorkerRemoved) Expected 0 Workers, got %d", len(workersReply.Workers))
			return
		}
	}

	// Test that a Blacklist entry is added after AddedToBlacklist event.
	if err := monitorDWH.onAddedToBlacklist(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE")); err != nil {
		t.Error(err)
		return
	}
	if blacklistReply, err := monitorDWH.getBlacklist(
		monitorDWH.ctx, &pb.BlacklistRequest{OwnerID: strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD")}); err != nil {
		t.Errorf("Failed to GetBlacklist: %s", err)
		return
	} else {
		if blacklistReply.OwnerID != strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD") {
			t.Errorf("(AddedToBlacklist) Expected %s, got %s (BlacklistReply.AdderID)",
				strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"), blacklistReply.OwnerID)
		}
	}

	// Test that a Blacklist entry is deleted after RemovedFromBlacklist event.
	if err := monitorDWH.onRemovedFromBlacklist(strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
		strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE")); err != nil {
		t.Error(err)
		return
	}
	if repl, err := monitorDWH.getBlacklist(
		monitorDWH.ctx, &pb.BlacklistRequest{OwnerID: strings.ToLower("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD")}); err != nil {
		t.Error(err)
		return
	} else {
		if len(repl.Addresses) > 0 {
			t.Errorf("GetBlacklist returned a blacklist that should have been deleted: %s", err)
		}
	}
}

func getDealChangeRequest(w *DWH, changeRequestID string) (*pb.DealChangeRequest, error) {
	rows, err := w.db.Query("SELECT * FROM DealChangeRequests WHERE Id=?", changeRequestID)
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return globalDWH.decodeDealChangeRequest(rows)
}

func getDealPayments(w *DWH) ([]*dealPayment, error) {
	rows, err := w.db.Query("SELECT * FROM DealPayments")
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*dealPayment
	for rows.Next() {
		var (
			billedTS   uint64
			paidAmount string
			dealID     string
		)
		if err := rows.Scan(&billedTS, &paidAmount, &dealID); err != nil {
			return nil, err
		} else {
			out = append(out, &dealPayment{
				BilledTS:   billedTS,
				PaidAmount: paidAmount,
				DealID:     dealID,
			})
		}
	}

	return out, nil
}

func getCertificates(w *DWH) ([]*pb.Certificate, error) {
	rows, err := w.db.Query("SELECT * FROM Certificates")
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.Certificate
	for rows.Next() {
		if certificate, err := w.decodeCertificate(rows); err != nil {
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
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := setupSQLite(w); err != nil {
		return err
	}

	var certs = []*pb.Certificate{
		{OwnerID: "consumer_id", Value: []byte("Consumer"), Attribute: CertificateName},
	}
	byteCerts, _ := json.Marshal(certs)

	for i := 0; i < 10; i++ {
		_, err := w.db.Exec(
			w.commands["insertDeal"],
			fmt.Sprintf("id_%d", i),
			fmt.Sprintf("supplier_%d", i),
			fmt.Sprintf("consumer_%d", i),
			fmt.Sprintf("master_%d", i),
			fmt.Sprintf("ask_id_%d", i),
			fmt.Sprintf("bid_id_%d", i),
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
			return err
		}

		_, err = w.db.Exec(
			w.commands["insertOrder"],
			fmt.Sprintf("ask_id_%d", i),
			12345, // CreatedTS
			fmt.Sprintf("deal_id_%d", i),
			uint64(pb.OrderType_ASK),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			"ask_author", // AuthorID
			"bid_author", // CounterpartyID
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
			w.commands["insertOrder"],
			fmt.Sprintf("bid_id_%d", i),
			12345, // CreatedTS
			fmt.Sprintf("deal_id_%d", i),
			uint64(pb.OrderType_BID),
			uint64(pb.OrderStatus_ORDER_ACTIVE),
			"bid_author", // AuthorID
			"ask_author", // CounterpartyID
			10010-i,      // Duration
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

		_, err = w.db.Exec(w.commands["insertDealChangeRequest"],
			fmt.Sprintf("changeRequest_%d", i), 0, 0, 0, 0, 0, "id_0")
		if err != nil {
			return err
		}

		var identityLevel int
		if (i % 2) == 0 {
			identityLevel = 0
		} else {
			identityLevel = 1
		}
		_, err = w.db.Exec("INSERT INTO Profiles VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			fmt.Sprintf("test_profile_%d", i), identityLevel, "sortedProfile", "", 0, 0, []byte{}, 0, 0)
		if err != nil {
			return err
		}
	}

	_, err := w.db.Exec("INSERT INTO Profiles VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fmt.Sprintf("consumer_id"), 3, "Consumer", "", 0, 0, byteCerts, 10, 10)
	if err != nil {
		return err
	}

	_, err = w.db.Exec("INSERT INTO Profiles VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		fmt.Sprintf("supplier_id"), 3, "Supplier", "", 0, 0, byteCerts, 10, 10)
	if err != nil {
		return err
	}

	_, err = w.db.Exec("INSERT INTO Blacklists VALUES (?, ?)", "blacklisting_user", "consumer_id")
	if err != nil {
		return err
	}

	if _, err := w.db.Exec(w.commands["updateLastKnownBlockSQLite"], 0); err != nil {
		w.logger.Error("failed to updateLastKnownBlockSQLite", zap.Error(err),
			zap.Uint64("block_number", 0))
	}

	return nil
}
