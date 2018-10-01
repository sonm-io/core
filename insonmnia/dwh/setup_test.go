package dwh

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
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
	"github.com/sonm-io/core/proto"
)

var (
	testDWH           *DWH
	testL1Processor   *L1Processor
	dbUser            = "dwh_tester"
	dbUserPassword    = "dwh_tester"
	globalDBName      = "dwh_test_db"
	monitorDBName     = "dwh_l1_processor_test_db"
	postgresPort      = "15432"
	serviceConnString = fmt.Sprintf("postgresql://localhost:%s/template1?user=postgres&sslmode=disable", postgresPort)
)

func TestAll(t *testing.T) {
	var ctx = context.Background()
	cli, containerID, err := startPostgresContainer(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		tearDownTests(ctx, containerID, cli)
	}()

	if err := setupTestDB(); err != nil {
		t.Error(err)
		return
	}

	testDWH, err = newTestDWH(getConnString(globalDBName, dbUser, dbUserPassword))
	if err != nil {
		t.Error(err)
		return
	}

	testL1Processor, err = newTestL1Processor(getConnString(monitorDBName, dbUser, dbUserPassword))
	if err != nil {
		t.Error(err)
		return
	}

	// This wrapper enables us to insert our own recovery logic _before_ Go's
	// testing module recovery (and execute teardown logic).
	wrapper := func(cb func(*testing.T)) func(*testing.T) {
		return func(t *testing.T) {
			defer func() {
				if err := recover(); err != nil {
					tearDownTests(ctx, containerID, cli)
					panic(err)
				}
			}()
			cb(t)
		}
	}

	tests := []func(*testing.T){
		testDWH_L1Processor,
		testGetStats,
		testGetDeals,
		testGetDealDetails,
		testGetOrders,
		testGetMatchingOrders,
		testGetOrderDetails,
		testGetDealChangeRequests,
		testGetProfiles,
	}
	for _, test := range tests {
		t.Run(GetFunctionName(test), wrapper(test))
	}
}

func GetFunctionName(i interface{}) string {
	split := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), ".")
	return split[len(split)-1]
}

func tearDownTests(ctx context.Context, containerID string, cli *client.Client) {
	if testDWH != nil && testDWH.db != nil {
		if err := testDWH.db.Close(); err != nil {
			fmt.Println(err)
		}
	}
	if testL1Processor != nil && testL1Processor.db != nil {
		if err := testL1Processor.db.Close(); err != nil {
			fmt.Println(err)
		}
	}
	if err := cli.ContainerStop(ctx, containerID, nil); err != nil {
		fmt.Println(err)
	}
	if err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{}); err != nil {
		fmt.Println(err)
	}
	if err := cli.Close(); err != nil {
		fmt.Println(err)
	}
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

func setupTestDB() error {
	db, err := sql.Open("postgres", serviceConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to template1: %s", err)
	}
	defer db.Close()

	if err := checkPostgresReadiness(db); err != nil {
		return fmt.Errorf("postgres not ready: %v", err)
	}

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

func checkPostgresReadiness(db *sql.DB) error {
	var err error
	for numRetries := 10; numRetries > 0; numRetries-- {
		if _, err := db.Exec("CREATE DATABASE is_ready"); err == nil {
			return nil
		}
		fmt.Printf("postgres container not ready, %d retries left\n", numRetries)
		time.Sleep(time.Second)
	}

	return fmt.Errorf("failed to connect to postgres container: %v", err)
}

func getConnString(database, user, password string) string {
	return fmt.Sprintf("postgresql://localhost:15432/%s?user=%s&password=%s&sslmode=disable", database, user, password)
}

func newTestDWH(dbEndpoint string) (*DWH, error) {
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

	w.storage, err = setupTestData(w.ctx, w.db, w.blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to setupTestData: %v", err)
	}

	return w, nil
}

func newTestL1Processor(dbEndpoint string) (*L1Processor, error) {
	var (
		ctx = context.Background()
		cfg = &L1ProcessorConfig{
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

	p := &L1Processor{
		ctx:        ctx,
		cfg:        cfg,
		blockchain: mockBlockchain,
		db:         db,
		logger:     log.GetLogger(ctx),
	}

	p.storage, err = setupTestData(p.ctx, p.db, p.blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to setupTestData: %v", err)
	}

	return p, nil
}

func setupTestData(ctx context.Context, db *sql.DB, blockchain bch.API) (*sqlStorage, error) {
	storage, err := setupDB(ctx, db, blockchain)
	if err != nil {
		return nil, err
	}

	var certs = []*sonm.Certificate{
		{OwnerID: sonm.NewEthAddress(common.HexToAddress("0xBB")), Value: []byte("Consumer"), Attribute: CertificateName},
	}
	byteCerts, _ := json.Marshal(certs)
	for i := 0; i < 2; i++ {
		insertDeal, args, _ := storage.builder().Insert("Deals").
			Columns(storage.tablesInfo.DealColumns...).
			Values(
				fmt.Sprintf("4040%d", i),
				common.HexToAddress(fmt.Sprintf("0x1%d", i)).Hex(), // Supplier
				common.HexToAddress(fmt.Sprintf("0x2%d", i)).Hex(), // Consumer
				common.HexToAddress(fmt.Sprintf("0x3%d", i)).Hex(), // Master
				fmt.Sprintf("2020%d", i),
				fmt.Sprintf("3030%d", i),
				10010+i,                                              // Duration
				sonm.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
				30010+i, // StartTime
				40010+i, // EndTime
				uint64(sonm.DealStatus_DEAL_ACCEPTED),
				sonm.NewBigIntFromInt(50010+int64(i)).PaddedString(), // BlockedBalance
				sonm.NewBigIntFromInt(60010+int64(i)).PaddedString(), // TotalPayout
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

		_, err := db.Exec(insertDeal, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to insertDeal: %v", err)
		}

		// Create 10 ASK orders.
		_, err = storage.builder().Insert("Orders").
			Columns(storage.tablesInfo.OrderColumns...).Values(
			fmt.Sprintf("2020%d", i),
			common.HexToAddress(fmt.Sprintf("0x9%d", i)).Hex(), // Master
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(sonm.OrderType_ASK),
			uint64(sonm.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xA").Hex(), // AuthorID
			common.Address{}.Hex(),           // CounterpartyID
			10010+i,
			sonm.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			7, // Netflags
			uint64(sonm.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},          // Tag
			fmt.Sprintf("3001%d", i), // FrozenSum
			uint64(sonm.IdentityLevel_REGISTERED),
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
		).RunWith(db).Exec()
		if err != nil {
			return nil, err
		}

		// Create 10 BID orders.
		_, err = storage.builder().Insert("Orders").
			Columns(storage.tablesInfo.OrderColumns...).Values(
			fmt.Sprintf("3030%d", i),
			common.HexToAddress(fmt.Sprintf("0x9%d", i)).Hex(), // Master
			12345, // CreatedTS
			fmt.Sprintf("1010%d", i),
			uint64(sonm.OrderType_BID),
			uint64(sonm.OrderStatus_ORDER_ACTIVE),
			common.HexToAddress("0xB").Hex(), // AuthorID
			common.Address{}.Hex(),           // CounterpartyID
			10010-i,                          // Duration
			sonm.NewBigIntFromInt(20010+int64(i)).PaddedString(), // Price
			5, // Netflags
			uint64(sonm.IdentityLevel_ANONYMOUS),
			fmt.Sprintf("blacklist_%d", i),
			[]byte{1, 2, 3},                       // Tag
			fmt.Sprintf("3001%d", i),              // FrozenSum
			uint64(sonm.IdentityLevel_REGISTERED), // CreatorIdentityLevel
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
		).RunWith(db).Exec()
		if err != nil {
			return nil, err
		}

		_, err = storage.builder().Insert("DealChangeRequests").
			Values(fmt.Sprintf("5050%d", i), 0, 0, 0, 0, 0, "40400").RunWith(db).Exec()
		if err != nil {
			return nil, err
		}

		var identityLevel int
		if (i % 2) == 0 {
			identityLevel = 0
		} else {
			identityLevel = 1
		}
		_, err = storage.builder().Insert("Profiles").Columns(storage.tablesInfo.ProfileColumns[1:]...).Values(
			common.HexToAddress(fmt.Sprintf("0x2%d", i)).Hex(),
			identityLevel,
			"sortedProfile",
			"",
			0,
			0,
			0,
			0,
		).RunWith(db).Exec()
		if err != nil {
			return nil, err
		}
	}

	// Create a couple of profiles for TestDWH_monitor entities.
	_, err = storage.builder().Insert("Profiles").Columns(storage.tablesInfo.ProfileColumns[1:]...).Values(
		fmt.Sprintf(common.HexToAddress("0xBB").Hex()),
		3,
		"Consumer",
		"",
		0,
		0,
		10,
		10,
	).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}
	_, err = storage.builder().Insert("Profiles").Columns(storage.tablesInfo.ProfileColumns[1:]...).Values(
		fmt.Sprintf(common.HexToAddress("0xAA").Hex()),
		3,
		"Supplier",
		"",
		0,
		0,
		10,
		10,
	).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}
	// Blacklist 0xBB for 0xE for testGetProfiles.
	_, err = storage.builder().Insert("Blacklists").Values(
		common.HexToAddress("0xE").Hex(),
		common.HexToAddress("0xBB").Hex(),
	).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	// Add a BID order that will be matched by any of the ASK orders added above and
	// blacklist this BID order's Author for the author of all ASK orders. Then in
	// testGetMatchingOrders we shouldn't get this order.
	_, err = storage.builder().Insert("Blacklists").Values(
		common.HexToAddress("0xA").Hex(),
		common.HexToAddress("0xCC").Hex(),
	).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}
	_, err = storage.builder().Insert("Orders").
		Columns(storage.tablesInfo.OrderColumns...).Values(
		fmt.Sprintf("3050%d", 0),
		common.HexToAddress(fmt.Sprintf("0x9%d", 0)).Hex(), // Master
		12345, // CreatedTS
		fmt.Sprintf("1010%d", 0),
		uint64(sonm.OrderType_BID),
		uint64(sonm.OrderStatus_ORDER_ACTIVE),
		common.HexToAddress("0xCC").Hex(), // AuthorID
		common.HexToAddress("0xA").Hex(),  // CounterpartyID
		10,                                // Duration
		sonm.NewBigIntFromInt(30010+int64(0)).PaddedString(), // Price
		5, // Netflags
		uint64(sonm.IdentityLevel_ANONYMOUS),
		fmt.Sprintf("blacklist_%d", 0),
		[]byte{1, 2, 3},                       // Tag
		fmt.Sprintf("3001%d", 0),              // FrozenSum
		uint64(sonm.IdentityLevel_REGISTERED), // CreatorIdentityLevel
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
	).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	return storage, nil
}
