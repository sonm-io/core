package dwh

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

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
	testDBPath = "dwh.db"
)

var (
	w *DWH
)

func TestMain(m *testing.M) {
	var err error
	w, err = getTestDWH()
	if err != nil {
		fmt.Println(err)
		os.Remove(testDBPath)
		os.Exit(1)
	}

	retCode := m.Run()
	w.db.Close()
	os.Remove(testDBPath)
	os.Exit(retCode)
}

func TestDWH_GetDeals(t *testing.T) {
	// Test TEXT columns.
	{
		request := &pb.DealsRequest{
			Status:     pb.DealStatus_DEAL_UNKNOWN,
			SupplierID: "supplier_5",
		}
		reply, err := w.GetDeals(context.Background(), request)

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
		reply, err := w.GetDeals(context.Background(), request)

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
		reply, err := w.GetDeals(context.Background(), request)

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
	deal, err := w.GetDealDetails(context.Background(), &pb.ID{Id: "id_5"})
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
	if string(deal.SupplierCertificates) != string([]byte{1, 2}) {
		t.Errorf("Expected %s, got %s (SupplierCertificates)", string([]byte{1, 2}), deal.SupplierCertificates)
	}
}

func TestDWH_GetOrders(t *testing.T) {
	// Test TEXT columns.
	{
		request := &pb.OrdersRequest{
			Type:   pb.OrderType_ANY,
			DealID: "deal_id_5",
		}
		reply, err := w.GetOrders(context.Background(), request)

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
		reply, err := w.GetOrders(context.Background(), request)

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
		reply, err := w.GetOrders(context.Background(), request)

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
	request := &pb.MatchingOrdersRequest{
		Id: &pb.ID{Id: "ask_id_5"},
	}
	reply, err := w.GetMatchingOrders(context.Background(), request)
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
	order, err := w.GetOrderDetails(context.Background(), &pb.ID{Id: "ask_id_5"})
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
	reply, err := w.GetDealChangeRequests(context.Background(), &pb.ID{Id: "id_0"})
	if err != nil {
		t.Error(err)
		return
	}

	if len(reply.Requests) != 10 {
		t.Errorf("Expected %d DealChangeRequests, got %d", 10, len(reply.Requests))
		return
	}
}

func TestDWH_monitor(t *testing.T) {
	var (
		controller           = gomock.NewController(t)
		mockBlock            = bch.NewMockAPI(controller)
		events               = make(chan *bch.Event, 5)
		commonID             = big.NewInt(0xDEADBEEF)
		commonEventTS uint64 = 5
	)
	defer close(events)

	mockBlock.EXPECT().GetEvents(gomock.Any(), gomock.Any()).AnyTimes().Return(events, nil)

	benchmarks, err := pb.NewBenchmarks([]uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	require.NoError(t, err)

	deal := &pb.Deal{
		Id:             commonID.String(),
		Benchmarks:     benchmarks,
		SupplierID:     "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE",
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
		AuthorID:       "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE",
		CounterpartyID: "counterparty_id",
		Duration:       10020,
		Price:          pb.NewBigInt(big.NewInt(20010)),
		Netflags:       7,
		IdentityLevel:  pb.IdentityLevel_ANONIMOUS,
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
		Id:    "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD",
		Level: 3,
	}
	mockBlock.EXPECT().GetValidator(gomock.Any(), gomock.Any()).AnyTimes().Return(validator, nil)

	certificate := &pb.Certificate{
		ValidatorID:   "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD",
		OwnerID:       "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE",
		Attribute:     CertificateName,
		IdentityLevel: 1,
		Value:         []byte("User Name"),
	}
	mockBlock.EXPECT().GetCertificate(gomock.Any(), gomock.Any()).AnyTimes().Return(
		certificate, nil)

	w.blockchain = mockBlock
	go w.monitorBlockchain()

	// Test onOrderPlaced event handling.
	events <- &bch.Event{Data: &bch.OrderPlacedData{ID: commonID}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if order, err := w.GetOrderDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetOrderDetails: %s", err)
		return
	} else {
		if order.GetOrder().Duration != 10020 {
			t.Errorf("Expected %d, got %d (Order.Duration)", 10020, order.GetOrder().Duration)
		}
	}

	// Test onDealOpened event handling.
	events <- &bch.Event{Data: &bch.DealOpenedData{ID: commonID}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	// Firstly, check that a deal was created.
	if deal, err := w.GetDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if deal.GetDeal().Duration != 10020 {
			t.Errorf("Expected %d, got %d (Deal.Duration)", 10020, deal.GetDeal().Duration)
		}
	}
	// Secondly, check that a DealCondition was created.
	if dealConditions, err := getDealConditions(t); err != nil {
		t.Errorf("Failed to getDealConditions: %s", err)
		return
	} else {
		if dealConditions[0].Duration != 10020 {
			t.Errorf("Expected %d, got %d (DealCondition.Duration)", 10020, deal.Duration)
			return
		}
	}

	// Test that a Validator entry is added after ValidatorCreated event.
	events <- &bch.Event{Data: &bch.ValidatorCreatedData{
		ID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
	}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if validatorsReply, err := w.GetValidators(w.ctx, &pb.ValidatorsRequest{}); err != nil {
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
	events <- &bch.Event{Data: &bch.ValidatorDeletedData{
		ID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD")}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if validatorsReply, err := w.GetValidators(w.ctx, &pb.ValidatorsRequest{}); err != nil {
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
	events <- &bch.Event{Data: &bch.CertificateCreatedData{
		ID: commonID,
	}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if certificateAttrs, err := getCertificates(t); err != nil {
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
	if profiles, err := getProfiles(t); err != nil {
		t.Errorf("Failed to getProfiles: %s", err)
		return
	} else {
		if len(profiles) != 1 {
			t.Errorf("(CertificateCreated) Expected 1 Profile, got %d",
				len(profiles))
			return
		}
		if profiles[0].Name != "User Name" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"User Name", profiles[0].Name)
		}
	}

	certificate.Attribute = CertificateCountry
	certificate.Value = []byte("Country")
	// Test that a  Profile entry is updated after CertificateCreated event.
	events <- &bch.Event{Data: &bch.CertificateCreatedData{
		ID: commonID,
	}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if profiles, err := getProfiles(t); err != nil {
		t.Errorf("Failed to getProfiles: %s", err)
		return
	} else {
		if len(profiles) != 1 {
			t.Errorf("(CertificateCreated) Expected 1 Profile, got %d",
				len(profiles))
			return
		}
		if profiles[0].Country != "Country" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Country)",
				"Country", profiles[0].Name)
		}
		if profiles[0].Name != "User Name" {
			t.Errorf("(CertificateCreated) Expected %s, got %s (Profile.Name)",
				"Name", profiles[0].Name)
		}

		var certificates []*pb.Certificate
		if err := json.Unmarshal(profiles[0].Certificates, &certificates); err != nil {
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
	dwhOrder, err := w.getOrderDetails(context.Background(), &pb.ID{Id: commonID.String()})
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
	if deal, err := w.GetDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if len(deal.SupplierCertificates) == 0 {
			t.Errorf("Expected some SupplierCertificated, got nothing")
		}
	}

	// Test that if order is updated, it is deleted.
	events <- &bch.Event{Data: &bch.OrderUpdatedData{ID: commonID}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if _, err := w.GetOrderDetails(context.Background(), &pb.ID{Id: commonID.String()}); err == nil {
		t.Error("GetOrderDetails returned an order that should have been deleted")
		return
	}

	deal.Duration += 1
	// Test onDealUpdated event handling.
	events <- &bch.Event{Data: &bch.DealUpdatedData{ID: commonID}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if deal, err := w.GetDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if deal.GetDeal().Duration != 10021 {
			t.Errorf("Expected %d, got %d (Deal.Duration)", 10021, deal.GetDeal().Duration)
		}
	}

	// Test creating an ASK DealChangeRequest.
	events <- &bch.Event{Data: &bch.DealChangeRequestSentData{ID: big.NewInt(0)}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if changeRequest, err := getDealChangeRequest(changeRequest.Id); err != nil {
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
	events <- &bch.Event{Data: &bch.DealChangeRequestSentData{ID: big.NewInt(1)}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if changeRequest, err := getDealChangeRequest(changeRequest.Id); err != nil {
		t.Errorf("Failed to getDealChangeRequest: %s", err)
		return
	} else {
		if changeRequest.Duration != 10021 {
			t.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10021, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest("0"); err == nil {
		t.Error("getDealChangeRequest returned a DealChangeRequest that should have been deleted")
		return
	}

	// Test that when a BID DealChangeRequest was created, it was kept (and nothing was deleted).
	changeRequest.Id = "2"
	changeRequest.Duration = 10022
	changeRequest.RequestType = pb.OrderType_BID
	events <- &bch.Event{Data: &bch.DealChangeRequestSentData{ID: big.NewInt(2)}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if changeRequest, err := getDealChangeRequest(changeRequest.Id); err != nil {
		t.Errorf("Failed to getDealChangeRequest: %s", err)
		return
	} else {
		if changeRequest.Duration != 10022 {
			t.Errorf("Expected %d, got %d (DealChangeRequest.Duration)", 10022, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest("1"); err != nil {
		t.Errorf("DealChangeRequest of type ASK was deleted after a BID DealChangeRequest creation: %s", err)
		return
	}

	// Test that when a DealChangeRequest is updated to any status but REJECTED, it is deleted.
	changeRequest.Id = "1"
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_ACCEPTED
	events <- &bch.Event{Data: &bch.DealChangeRequestUpdatedData{ID: big.NewInt(1)}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if _, err := getDealChangeRequest("1"); err == nil {
		t.Error("DealChangeRequest which status was changed to ACCEPTED was not deleted")
		return
	}
	// Also test that a new DealCondition was created, and the old one was updated.
	if conditions, err := getDealConditions(t); err != nil {
		t.Errorf("Failed to getDealConditions: %s", err)
		return
	} else {
		if len(conditions) != 2 {
			t.Errorf("Expected 2 DealConditions, got %d", len(conditions))
			return
		}
		if conditions[0].EndTime.Seconds != 5 {
			t.Errorf("Expected %d, got %d (DealCondition.EndTime)", 5, conditions[0].EndTime.Seconds)
			return
		}
		if conditions[1].StartTime.Seconds != 5 {
			t.Errorf("Expected %d, got %d (DealCondition.StartTime)", 5, conditions[1].StartTime.Seconds)
			return
		}
	}

	// Test that when a DealChangeRequest is updated to REJECTED, it is kept.
	changeRequest.Id = "2"
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_REJECTED
	events <- &bch.Event{Data: &bch.DealChangeRequestUpdatedData{ID: big.NewInt(2)}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if _, err := getDealChangeRequest("2"); err != nil {
		t.Error("DealChangeRequest which status was changed to REJECTED was deleted")
		return
	}

	// Test that after a Billed event last DealCondition.Payout is updated.
	events <- &bch.Event{Data: &bch.BilledData{
		ID: commonID, PayedAmount: big.NewInt(10)}, TS: commonEventTS * 2,
	}
	time.Sleep(time.Millisecond * 200)
	if dealConditions, err := getDealConditions(t); err != nil {
		t.Errorf("Failed to GetDealDetails: %s", err)
		return
	} else {
		if len(dealConditions) != 2 {
			t.Errorf("(Billed) Expected 2 DealConditions, got %d", len(dealConditions))
			return
		}
		if dealConditions[1].TotalPayout.Unwrap().String() != "10" {
			t.Errorf("(Billed) Expected %s, got %s (DealCondition.TotalPayout)",
				"10", dealConditions[1].TotalPayout.String())
		}
	}
	if dealPayments, err := getDealPayments(t); err != nil {
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
	events <- &bch.Event{Data: &bch.DealUpdatedData{ID: commonID}, TS: commonEventTS}
	time.Sleep(time.Millisecond * 200)
	if _, err := w.GetDealDetails(context.Background(), &pb.ID{Id: commonID.String()}); err == nil {
		t.Errorf("Deal was not deleted after status changing to CLOSED")
		return
	}
	if dealConditions, err := getDealConditions(t); err != nil {
		t.Errorf("Failed to getDealConditions: %s", err)
		return
	} else {
		if len(dealConditions) != 0 {
			t.Errorf("(DealUpdated) Expected 0 DealConditions, got %d", len(dealConditions))
			return
		}
	}

	// Test that a worker is added after a WorkerAnnounced event.
	events <- &bch.Event{
		Data: &bch.WorkerAnnouncedData{
			MasterID: common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
			SlaveID:  common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		},
		TS: commonEventTS,
	}
	time.Sleep(time.Millisecond * 200)
	if workersReply, err := w.GetWorkers(w.ctx, &pb.WorkersRequest{}); err != nil {
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
	events <- &bch.Event{
		Data: &bch.WorkerConfirmedData{
			MasterID: common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
			SlaveID:  common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		},
		TS: commonEventTS,
	}
	time.Sleep(time.Millisecond * 200)
	if workersReply, err := w.GetWorkers(w.ctx, &pb.WorkersRequest{}); err != nil {
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
	events <- &bch.Event{
		Data: &bch.WorkerRemovedData{
			MasterID: common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
			SlaveID:  common.StringToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		},
		TS: commonEventTS,
	}
	time.Sleep(time.Millisecond * 200)
	if workersReply, err := w.GetWorkers(w.ctx, &pb.WorkersRequest{}); err != nil {
		t.Errorf("Failed to getWorkers: %s", err)
		return
	} else {
		if len(workersReply.Workers) != 0 {
			t.Errorf("(WorkerRemoved) Expected 0 Workers, got %d", len(workersReply.Workers))
			return
		}
	}

	// Test that a Blacklist entry is added after AddedToBlacklist event.
	events <- &bch.Event{
		Data: &bch.AddedToBlacklistData{
			AdderID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
			AddeeID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		},
		TS: commonEventTS,
	}
	time.Sleep(time.Millisecond * 200)
	if blacklistReply, err := w.GetBlacklist(
		w.ctx, &pb.ID{Id: "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"}); err != nil {
		t.Errorf("Failed to GetBlacklist: %s", err)
		return
	} else {
		if blacklistReply.OwnerID != "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD" {
			t.Errorf("(AddedToBlacklist) Expected %s, got %s (BlacklistReply.AdderID)",
				"0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD", blacklistReply.OwnerID)
		}
	}

	// Test that a Blacklist entry is deleted after RemovedFromBlacklist event.
	events <- &bch.Event{
		Data: &bch.RemovedFromBlacklistData{
			RemoverID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"),
			RemoveeID: common.HexToAddress("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FE"),
		},
		TS: commonEventTS,
	}
	time.Sleep(time.Millisecond * 200)
	if _, err := w.GetBlacklist(
		w.ctx, &pb.ID{Id: "0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD"}); err == nil {
		t.Errorf("GetBlacklist returned a blacklist that should have been deleted: %s", err)
		return
	}
}

func getDealChangeRequest(changeRequestID string) (*pb.DealChangeRequest, error) {
	rows, err := w.db.Query("SELECT * FROM DealChangeRequests WHERE Id=?", changeRequestID)
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return w.decodeDealChangeRequest(rows)
}

func getDealConditions(t *testing.T) ([]*pb.DealCondition, error) {
	rows, err := w.db.Query("SELECT rowid, * FROM DealConditions")
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.DealCondition
	for rows.Next() {
		if dealCondition, err := w.decodeDealCondition(rows); err != nil {
			t.Errorf("decodeDealCondition: %s", err)
		} else {
			out = append(out, dealCondition)
		}
	}

	return out, nil
}

func getProfiles(t *testing.T) ([]*pb.Profile, error) {
	rows, err := w.db.Query("SELECT * FROM Profiles")
	if err != nil {
		return nil, errors.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.Profile
	for rows.Next() {
		if profile, err := w.decodeProfile(rows); err != nil {
			t.Errorf("failed to decode Profile: %s", err)
		} else {
			out = append(out, profile)
		}
	}

	return out, nil
}

func getDealPayments(t *testing.T) ([]*dealPayment, error) {
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
			t.Errorf("failed to decode DealPayment: %s", err)
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

func getCertificates(t *testing.T) ([]*pb.Certificate, error) {
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

func getTestDWH() (*DWH, error) {
	var (
		ctx = context.Background()
		cfg = &Config{
			Storage: &storageConfig{
				Backend:  "sqlite3",
				Endpoint: testDBPath,
			},
		}
		w = &DWH{
			ctx:      ctx,
			cfg:      cfg,
			logger:   log.GetLogger(ctx),
			commands: sqliteCommands,
		}
	)

	w.mu.Lock()
	defer w.mu.Unlock()

	if err := setupSQLite(w); err != nil {
		return nil, err
	}

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
			70010+i,      // LastBillTS
			5,            // Netflags
			3,            // AskIdentityLevel
			4,            // BidIdentityLevel
			[]byte{1, 2}, // SupplierCertificates
			[]byte{3, 4}, // ConsumerCertificates
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
			return nil, err
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
			uint64(pb.IdentityLevel_ANONIMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},          // Tag
			fmt.Sprintf("3001%d", i), // FrozenSum
			uint64(pb.IdentityLevel_PSEUDONYMOUS),
			"CreatorName",
			"CreatorCountry",
			[]byte{}, // CreatorCertificates
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
			return nil, err
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
			uint64(pb.IdentityLevel_ANONIMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},                       // Tag
			fmt.Sprintf("3001%d", i),              // FrozenSum
			uint64(pb.IdentityLevel_PSEUDONYMOUS), // CreatorIdentityLevel
			"CreatorName",
			"CreatorCountry",
			[]byte{}, // CreatorCertificates
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
			return nil, err
		}

		_, err = w.db.Exec(w.commands["insertDealChangeRequest"],
			fmt.Sprintf("changeRequest_%d", i), 0, 0, 0, 0, 0, "id_0")
		if err != nil {
			return nil, err
		}
	}

	if _, err := w.db.Exec(w.commands["updateLastKnownBlockSQLite"], 0); err != nil {
		w.logger.Error("failed to updateLastKnownBlockSQLite", zap.Error(err),
			zap.Uint64("block_number", 0))
	}

	return w, nil
}
