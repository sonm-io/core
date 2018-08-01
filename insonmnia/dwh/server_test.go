package dwh

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	log "github.com/noxiouz/zapctx/ctxlog"
	bch "github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
)

var (
	globalDWH         *DWH
	monitorDWH        *DWH
	dbUser            = "dwh_tester"
	dbUserPassword    = "dwh_tester"
	globalDBName      = "dwh_test_global"
	monitorDBName     = "dwh_test_monitor"
	postgresPort      = "15432"
	serviceConnString = fmt.Sprintf("postgresql://localhost:%s/template1?user=postgres&sslmode=disable", postgresPort)
)

func TestMain(m *testing.M) {
	var (
		err             error
		testsReturnCode = 1
		ctx             = context.Background()
	)

	cli, containerID, err := startPostgresContainer(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(testsReturnCode)
	}

	if err := checkPostgresReadiness(containerID); err != nil {
		fmt.Println(err)
		os.Exit(testsReturnCode)
	}

	defer func() {
		if globalDWH != nil && globalDWH.db != nil {
			if err := globalDWH.db.Close(); err != nil {
				fmt.Println(err)
			}
		}
		if monitorDWH != nil && monitorDWH.db != nil {
			if err := monitorDWH.db.Close(); err != nil {
				fmt.Println(err)
			}
		}
		if err := tearDownDB(); err != nil {
			fmt.Println(err)
		}
		if err := cli.ContainerStop(ctx, containerID, nil); err != nil {
			fmt.Println(err)
		}
		fmt.Println(1)
		if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
			fmt.Println(err)
		}
		fmt.Println(2)
		if err := cli.Close(); err != nil {
			fmt.Println(err)
		}
		os.Exit(testsReturnCode)
	}()

	if err := setupTestDB(); err != nil {
		fmt.Println(err)
		return
	}

	globalDWH, err = getTestDWH(getConnString(globalDBName, dbUser, dbUserPassword))
	if err != nil {
		fmt.Println(err)
		return
	}

	monitorDWH, err = getTestDWH(getConnString(monitorDBName, dbUser, dbUserPassword))
	if err != nil {
		fmt.Println(err)
		return
	}

	testsReturnCode = m.Run()
}

func startPostgresContainer(ctx context.Context) (cli *client.Client, containerID string, err error) {
	cli, err = client.NewEnvClient()
	if err != nil {
		return nil, "", fmt.Errorf("failed to setup Docker client: %s", err)
	}

	reader, err := cli.ImagePull(ctx, "docker.io/library/postgres", types.ImagePullOptions{})
	if err != nil {
		cli.Close()
		return nil, "", fmt.Errorf("failed to pull postgres image: %s", err)
	}
	io.Copy(os.Stdout, reader)

	containerCfg := &container.Config{
		Image:        "postgres",
		ExposedPorts: nat.PortSet{"5432": struct{}{}},
	}
	hostCfg := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			nat.Port("5432"): {{HostIP: "localhost", HostPort: postgresPort}},
		},
	}
	resp, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, "dwh-postgres-test")
	if err != nil {
		cli.Close()
		return nil, "", fmt.Errorf("failed to create container: %s", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		cli.Close()
		return nil, "", fmt.Errorf("failed to start container: %s", err)
	}

	return cli, resp.ID, nil
}

func checkPostgresReadiness(containerID string) error {
	var (
		err        error
		numRetries = 10
	)
	for ; numRetries > 0; numRetries-- {
		cmd := exec.Command("docker", "exec", containerID, "pg_isready")
		out, err := cmd.CombinedOutput()
		if err == nil {
			if strings.Contains(string(out), "accepting connections") {
				return nil
			}
		}
		fmt.Printf("postgres container not ready, %d retries left\n", numRetries)
		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to postgres container: %v", err)
}

func setupTestDB() error {
	db, err := sql.Open("postgres", serviceConnString)
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
	db, err := sql.Open("postgres", serviceConnString)
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
	return fmt.Sprintf("postgresql://localhost:15432/%s?user=%s&password=%s&sslmode=disable", database, user, password)
}

func getTestDWH(dbEndpoint string) (*DWH, error) {
	var (
		ctx = context.Background()
		cfg = &DWHConfig{
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

	return w, setupTestData(w)
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

func setupTestData(w *DWH) error {
	var err error
	if w.storage, err = setupDB(w.ctx, w.db, w.blockchain); err != nil {
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
	_, err = w.storage.builder().Insert("Profiles").Columns(w.storage.tablesInfo.ProfileColumns[1:]...).Values(
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
