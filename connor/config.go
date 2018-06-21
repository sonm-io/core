package connor

import (
	"time"

	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
)

type marketConfig struct {
	Endpoint string `yaml:"endpoint" required:"true"`
}

type databaseConfig struct {
	Driver     string `yaml:"driver"`
	DataSource string `yaml:"data_source"`
}

type chargeOrdersConfig struct {
	Step        float64 `yaml:"step"`
	Start       float64 `yaml:"start"`
	Destination float64 `yaml:"destination"`
}

type tradeParamConfig struct {
	IdentityForBid      sonm.IdentityLevel `yaml:"identity_for_bid"`
	MarginAccounting    float64            `yaml:"margin_accounting"`
	PartCharge          float64            `yaml:"part_charge"`
	OrdersChangePercent float64            `yaml:"orders_change_percent"`
	DealsChangePercent  float64            `yaml:"deals_change_percent"`
	WaitingTimeCRSec    int64              `yaml:"waiting_time_change_request"`
}

type poolParamConfig struct {
	PoolAccount              string  `yaml:"pool_account"`
	WorkerLimitChangePercent float64 `yaml:"worker_limit_change_percent"`
	BadWorkersPercent        float64 `yaml:"bad_workers_percent"`
	Image                    string  `yaml:"image"`
	EmailForPool             string  `yaml:"email_for_pool"`
}

type typicalBenchmark struct {
	RamSize           uint64 `yaml:"ram_size"`
	CpuCores          uint64 `yaml:"cpu_cores"`
	CpuSysbenchSingle uint64 `yaml:"cpu_sysbench_single"`
	CpuSysbenchMulti  uint64 `yaml:"cpu_sysbench_multi"`
	NetDownload       uint64 `yaml:"net_download"`
	NetUpload         uint64 `yaml:"net_upload"`
	GpuCount          uint64 `yaml:"gpu_count"`
	GpuMem            uint64 `yaml:"gpu_mem"`
}

type tickerConfig struct {
	TradeTicker time.Duration `yaml:"trade_ticker"`
	DataUpdate  time.Duration `yaml:"data_update"`
	PoolInit    time.Duration `yaml:"pool_init"`
	TaskCheck   time.Duration `yaml:"task_check"`
}

type Config struct {
	Market       marketConfig       `yaml:"market" required:"true"`
	Database     databaseConfig     `yaml:"database"`
	UsingToken   string             `yaml:"using_token"`
	ChargeOrders chargeOrdersConfig `yaml:"charge_orders_parameters"`
	Trade        tradeParamConfig   `yaml:"trading_parameters"`
	Pool         poolParamConfig    `yaml:"pool_parameters"`
	Tickers      tickerConfig       `yaml:"tickers"`
	Benchmark    typicalBenchmark   `yaml:"benchmark"`
	Eth          accounts.EthConfig `yaml:"ethereum" required:"true"`
	Log          logging.Config     `yaml:"log"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
