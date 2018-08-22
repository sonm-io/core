package dwh

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	bch "github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
)

func testDWH_L1Processor(t *testing.T) {
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
	mockProfiles.EXPECT().GetProfileLevel(gomock.Any(), gomock.Any()).AnyTimes().Return(pb.IdentityLevel_ANONYMOUS, nil)
	cert := &pb.Certificate{
		ValidatorID:   pb.NewEthAddress(common.HexToAddress("0xC")),
		OwnerID:       pb.NewEthAddress(common.HexToAddress("0xD")),
		Attribute:     CertificateName,
		IdentityLevel: 1,
		Value:         []byte("User Name"),
	}
	mockProfiles.EXPECT().GetCertificate(gomock.Any(), gomock.Any()).AnyTimes().Return(
		cert, nil)
	mockBlock.EXPECT().Market().AnyTimes().Return(mockMarket)
	mockBlock.EXPECT().ProfileRegistry().AnyTimes().Return(mockProfiles)

	testL1Processor.blockchain = mockBlock

	err = testL1Processor.storage.InsertWorker(newSimpleConn(testL1Processor.db), common.Address{},
		common.HexToAddress("0x000000000000000000000000000000000000000d"))
	if err != nil {
		t.Error("failed to insert worker (additional)")
	}

	if err := testOrderPlaced(testL1Processor, commonEventTS, commonID); err != nil {
		t.Errorf("testOrderPlaced: %s", err)
		return
	}
	if err := testDealOpened(testL1Processor, deal, commonID); err != nil {
		t.Errorf("testDealOpened: %s", err)
		return
	}
	if err := testValidatorCreatedUpdated(testL1Processor, validator); err != nil {
		t.Errorf("testValidatorCreatedUpdated: %s", err)
		return
	}
	if err := testCertificateCreated(testL1Processor, cert, commonID); err != nil {
		t.Errorf("testCertificateCreated: %s", err)
		return
	}
	if err := testCertificateUpdated(testL1Processor, cert); err != nil {
		t.Errorf("testCertificateUpdated: %s", err)
		return
	}
	if err := testOrderUpdated(testL1Processor, order, commonID); err != nil {
		t.Errorf("testOrderUpdated: %s", err)
		return
	}
	err = testL1Processor.storage.DeleteWorker(newSimpleConn(testL1Processor.db), common.Address{},
		common.HexToAddress("0x000000000000000000000000000000000000000d"))
	if err != nil {
		t.Error("failed to delete worker (additional)")
	}
	if err := testDealUpdated(testL1Processor, deal, commonID); err != nil {
		t.Errorf("testDealUpdated: %s", err)
		return
	}
	if err := testDealChangeRequestSentAccepted(testL1Processor, changeRequest, commonEventTS, commonID); err != nil {
		t.Errorf("testDealChangeRequestSentAccepted: %s", err)
		return
	}
	if err := testBilled(testL1Processor, commonEventTS, commonID); err != nil {
		t.Errorf("testBilled: %s", err)
		return
	}
	if err := testDealClosed(testL1Processor, deal, commonID); err != nil {
		t.Errorf("testDealClosed: %s", err)
		return
	}
	if err := testWorkerAnnouncedConfirmedRemoved(testL1Processor); err != nil {
		t.Errorf("testWorkerAnnouncedConfirmedRemoved: %s", err)
		return
	}
	if err := testBlacklistAddedRemoved(testL1Processor); err != nil {
		t.Errorf("testBlacklistAddedRemoved: %s", err)
		return
	}
}

func testOrderPlaced(p *L1Processor, commonEventTS uint64, commonID *big.Int) error {
	if err := p.onOrderPlaced(commonEventTS, commonID); err != nil {
		return fmt.Errorf("onOrderPlaced failed: %v", err)
	}
	if order, err := p.storage.GetOrderByID(newSimpleConn(p.db), commonID); err != nil {
		return fmt.Errorf("storage.GetOrderByID failed: %v", err)
	} else {
		if order.GetOrder().Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (Order.Duration)", 10020, order.GetOrder().Duration)
		}
	}
	return nil
}

func testDealOpened(p *L1Processor, deal *pb.Deal, commonID *big.Int) error {
	if err := p.onDealOpened(commonID); err != nil {
		return fmt.Errorf("onDealOpened failed: %v", err)
	}
	// Firstly, check that a deal was created.
	if deal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if deal.GetDeal().Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (Deal.Duration)", 10020, deal.GetDeal().Duration)
		}
	}
	// Secondly, check that a DealCondition was created.
	if dealConditions, _, err := p.storage.GetDealConditions(
		newSimpleConn(p.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
		return fmt.Errorf("getDealConditions failed: %v", err)
	} else {
		if dealConditions[0].Duration != 10020 {
			return fmt.Errorf("expected %d, got %d (DealCondition.Duration)", 10020, deal.Duration)
		}
	}
	return nil
}

func testValidatorCreatedUpdated(p *L1Processor, validator *pb.Validator) error {
	// Check that a Validator entry is added after ValidatorCreated event.
	if err := p.onValidatorCreated(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return fmt.Errorf("onValidatorCreated failed: %v", err)
	}
	if validators, _, err := p.storage.GetValidators(newSimpleConn(p.db), &pb.ValidatorsRequest{}); err != nil {
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
	if err := p.onValidatorDeleted(common.HexToAddress(common.HexToAddress("0xC").Hex())); err != nil {
		return fmt.Errorf("onValidatorDeleted failed: %v", err)
	}
	if validators, _, err := p.storage.GetValidators(newSimpleConn(p.db), &pb.ValidatorsRequest{}); err != nil {
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

func testCertificateCreated(p *L1Processor, certificate *pb.Certificate, commonID *big.Int) error {
	// Check that a Certificate entry is created after CertificateCreated event. We create a special certificate,
	// `Name`, that will be recorded directly into profile. There's two such certificate types: `Name` and `Country`.
	if err := p.onCertificateCreated(commonID); err != nil {
		return fmt.Errorf("onCertificateCreated failed: %v", err)
	}
	if certificateAttrs, err := getCertificates(testL1Processor); err != nil {
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
	if profiles, _, err := p.storage.GetProfiles(newSimpleConn(p.db), &pb.ProfilesRequest{
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
	// Check that a Profile entry is updated after CertificateCreated event.
	if err := p.onCertificateCreated(commonID); err != nil {
		return fmt.Errorf("onCertificateCreated failed: %v", err)
	}
	if profiles, _, err := p.storage.GetProfiles(newSimpleConn(p.db), &pb.ProfilesRequest{}); err != nil {
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
	dwhOrder, err := p.storage.GetOrderByID(newSimpleConn(p.db), commonID)
	if err != nil {
		return fmt.Errorf("storage.GetOrderByID failed: %v", err)
	}
	if dwhOrder.CreatorIdentityLevel != 3 {
		return fmt.Errorf("(CertificateCreated) Expected %d, got %d (Order.CreatorIdentityLevel)",
			3, dwhOrder.CreatorIdentityLevel)
	}
	// Check that profile updates resulted in deals updates.
	if deal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if len(deal.SupplierCertificates) == 0 {
			return fmt.Errorf("expected some SupplierCertificates, got nothing")
		}
	}
	return nil
}

func testCertificateUpdated(p *L1Processor, cert *pb.Certificate) error {
	if err := p.onCertificateUpdated(newSimpleConn(p.db), cert.Id.Unwrap()); err != nil {
		return fmt.Errorf("(CertificateUpdated) %v", err)
	}

	profile, err := p.storage.GetProfileByID(newSimpleConn(p.db), cert.OwnerID.Unwrap())
	if err != nil {
		return err
	}

	if profile.IdentityLevel != uint64(pb.IdentityLevel_ANONYMOUS) {
		return fmt.Errorf("(CertificateUpdated) Expected %d, got %d (Profile.IdentityLevel)",
			pb.IdentityLevel_ANONYMOUS, profile.IdentityLevel)
	}

	return nil
}

func testOrderUpdated(p *L1Processor, order *pb.Order, commonID *big.Int) error {
	// Check that if order is updated, it is deleted. Order should be deleted because its DealID is not set
	// (this means that is has become inactive due to a cancellation and not a match).
	order.OrderStatus = pb.OrderStatus_ORDER_INACTIVE
	if err := p.onOrderUpdated(commonID); err != nil {
		return fmt.Errorf("onOrderUpdated failed: %v", err)
	}
	dwhOrder, err := p.storage.GetOrderByID(newSimpleConn(p.db), commonID)
	if err != nil {
		return fmt.Errorf("GetOrderByID failed: %v", err)
	}

	if dwhOrder.Order.OrderStatus != pb.OrderStatus_ORDER_INACTIVE {
		return errors.New("order was not deactivated")
	}

	return nil
}

func testDealUpdated(p *L1Processor, deal *pb.Deal, commonID *big.Int) error {
	deal.Duration += 1
	// Test onDealUpdated event handling.
	if err := p.onDealUpdated(commonID); err != nil {
		return fmt.Errorf("onDealUpdated failed: %v", err)
	}
	if deal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID); err != nil {
		return fmt.Errorf("storage.GetDealByID failed: %v", err)
	} else {
		if deal.GetDeal().Duration != 10021 {
			return fmt.Errorf("expected %d, got %d (Deal.Duration)", 10021, deal.GetDeal().Duration)
		}
	}
	return nil
}

func testDealChangeRequestSentAccepted(p *L1Processor, changeRequest *pb.DealChangeRequest, commonEventTS uint64, commonID *big.Int) error {
	// Test creating an ASK DealChangeRequest.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Duration = 10021
	if err := p.onDealChangeRequestSent(commonEventTS, big.NewInt(1)); err != nil {
		return fmt.Errorf("onDealChangeRequestSent (2) failed: %v", err)
	}
	if changeRequest, err := getDealChangeRequest(testL1Processor, changeRequest.Id); err != nil {
		return fmt.Errorf("getDealChangeRequest (2) failed: %v", err)
	} else {
		if changeRequest.Duration != 10021 {
			return fmt.Errorf("expected %d, got %d (DealChangeRequest.Duration)", 10021, changeRequest.Duration)
		}
	}
	if _, err := getDealChangeRequest(testL1Processor, pb.NewBigIntFromInt(0)); err == nil {
		return errors.New("getDealChangeRequest returned a DealChangeRequest that should have been deleted")
	}
	// Check that when a DealChangeRequest is updated to any status but REJECTED, it is deleted.
	changeRequest.Id = pb.NewBigIntFromInt(1)
	changeRequest.Status = pb.ChangeRequestStatus_REQUEST_ACCEPTED
	if err := p.onDealChangeRequestUpdated(commonEventTS, big.NewInt(1)); err != nil {
		return fmt.Errorf("onDealChangeRequestUpdated failed: %v", err)
	}
	// Also test that a new DealCondition was created, and the old one was updated.
	if dealConditions, _, err := p.storage.GetDealConditions(
		newSimpleConn(p.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
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

func testBilled(p *L1Processor, commonEventTS uint64, commonID *big.Int) error {
	deal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID)
	if err != nil {
		return fmt.Errorf("GetDealByID failed: %v", err)
	}

	if deal.Deal.LastBillTS.Seconds != int64(commonEventTS) {
		return fmt.Errorf("unexpected LastBillTS (%d)", deal.Deal.LastBillTS)
	}

	// Check that after a Billed event last DealCondition.Payout is updated.
	newBillTS := commonEventTS + 1
	if err := p.onBilled(newBillTS, commonID, big.NewInt(10)); err != nil {
		return fmt.Errorf("onBilled failed: %v", err)
	}
	if dealConditions, _, err := p.storage.GetDealConditions(
		newSimpleConn(p.db), &pb.DealConditionsRequest{DealID: pb.NewBigInt(commonID)}); err != nil {
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
	updatedDeal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID)
	if err != nil {
		return fmt.Errorf("GetDealByID failed: %v", err)
	}

	if updatedDeal.Deal.LastBillTS.Seconds != int64(newBillTS) {
		return fmt.Errorf("(Billed) Expected %d, got %d (Deal.LastBillTS)",
			newBillTS, updatedDeal.Deal.LastBillTS.Seconds)
	}

	return nil
}

func testDealClosed(p *L1Processor, deal *pb.Deal, commonID *big.Int) error {
	// Check that when a Deal's status is updated to CLOSED, Deal and its DealConditions are deleted.
	deal.Status = pb.DealStatus_DEAL_CLOSED
	// Test onDealUpdated event handling.
	if err := p.onDealUpdated(commonID); err != nil {
		return fmt.Errorf("onDealUpdated failed: %v", err)
	}
	dwhDeal, err := p.storage.GetDealByID(newSimpleConn(p.db), commonID)
	if err != nil {
		return fmt.Errorf("GetDealByID failed: %v", err)
	}

	if dwhDeal.Deal.Status != pb.DealStatus_DEAL_CLOSED {
		return errors.New("failed to deactivate closed deal")
	}

	return nil
}

func testWorkerAnnouncedConfirmedRemoved(p *L1Processor) error {
	// Check that a worker is added after a WorkerAnnounced event.
	if err := p.onWorkerAnnounced(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerAnnounced failed: %v", err)
	}
	if workers, _, err := p.storage.GetWorkers(newSimpleConn(p.db), &pb.WorkersRequest{}); err != nil {
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
	if err := p.onWorkerConfirmed(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerConfirmed failed: %v", err)
	}
	if workers, _, err := p.storage.GetWorkers(newSimpleConn(p.db), &pb.WorkersRequest{}); err != nil {
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
	if err := p.onWorkerRemoved(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onWorkerRemoved failed: %v", err)
	}
	if workers, _, err := p.storage.GetWorkers(newSimpleConn(p.db), &pb.WorkersRequest{}); err != nil {
		return fmt.Errorf("getWorkers failed: %v", err)
	} else {
		if len(workers) != 0 {
			return fmt.Errorf("(WorkerRemoved) Expected 0 Workers, got %d", len(workers))
		}
	}
	return nil
}

func testBlacklistAddedRemoved(p *L1Processor) error {
	// Check that a Blacklist entry is added after AddedToBlacklist event.
	if err := p.onAddedToBlacklist(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onAddedToBlacklist failed: %v", err)
	}
	if blacklistReply, err := p.storage.GetBlacklist(
		newSimpleConn(p.db), &pb.BlacklistRequest{UserID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return fmt.Errorf("getBlacklist failed: %v", err)
	} else {
		if blacklistReply.OwnerID.Unwrap().Hex() != common.HexToAddress("0xC").Hex() {
			return fmt.Errorf("(AddedToBlacklist) Expected %s, got %s (BlacklistReply.AdderID)",
				common.HexToAddress("0xC").Hex(), blacklistReply.OwnerID)
		}
	}
	// Check that a Blacklist entry is deleted after RemovedFromBlacklist event.
	if err := p.onRemovedFromBlacklist(common.HexToAddress("0xC"), common.HexToAddress("0xD")); err != nil {
		return fmt.Errorf("onRemovedFromBlacklist failed: %v", err)
	}
	if repl, err := p.storage.GetBlacklist(
		newSimpleConn(p.db), &pb.BlacklistRequest{UserID: pb.NewEthAddress(common.HexToAddress("0xC"))}); err != nil {
		return fmt.Errorf("getBlacklist (2) failed: %v", err)
	} else {
		if len(repl.Addresses) > 0 {
			return fmt.Errorf("getBlacklist returned a blacklist that should have been deleted: %+v", repl.Addresses)
		}
	}
	return nil
}

func getDealChangeRequest(p *L1Processor, changeRequestID *pb.BigInt) (*pb.DealChangeRequest, error) {
	rows, err := p.storage.builder().Select("*").From("DealChangeRequests").
		Where("Id = ?", changeRequestID.Unwrap().String()).RunWith(p.db).Query()
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.New("no rows returned")
	}

	return p.storage.decodeDealChangeRequest(rows)
}

func getCertificates(p *L1Processor) ([]*pb.Certificate, error) {
	rows, err := p.storage.builder().Select("*").From("Certificates").RunWith(p.db).Query()
	if err != nil {
		return nil, fmt.Errorf("query failed: %s", err)
	}
	defer rows.Close()

	var out []*pb.Certificate
	for rows.Next() {
		if certificate, err := p.storage.decodeCertificate(rows); err != nil {
			return nil, fmt.Errorf("failed to decodeCertificate: %v", err)
		} else {
			out = append(out, certificate)
		}
	}

	return out, nil
}
