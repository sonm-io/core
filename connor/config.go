package connor

import (
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
)

type marketConfig struct {
	Endpoint string `yaml:"endpoint" required:"true"`
}

type usingToken struct {
	Token string `yaml:"token"`
}
type poolAddressesConfig struct {
	ZecPoolAddr string `yaml:"zec_pool_addr"`
	XmrPoolAddr string `yaml:"xmr_pool_addr"`
	EthPoolAddr string `yaml:"eth_pool_addr"`
}
type stepsConfig struct {
	StepForETH float64 `yaml:"stepETH"`
	StepForZEC float64 `yaml:"stepZEC"`
	StepForXMR float64 `yaml:"stepXMR"`
}
type chargeOrdersETHConfig struct {
	Start       float64 `yaml:"start"`
	Destination float64 `yaml:"destination"`
}
type chargeOrdersZECConfig struct {
	Start       float64 `yaml:"start"`
	Destination float64 `yaml:"destination"`
}
type chargeOrdersXMRConfig struct {
	Start       float64 `yaml:"start"`
	Destination float64 `yaml:"destination"`
}
type imageConfig struct {
	Image string `yaml:"image"`
}
type databaseConfig struct {
	Driver     string `yaml:"driver"`
	DataSource string `yaml:"data_source"`
}
type sensitivityConfig struct {
	SensitivityForOrders     int64   `yaml:"reaction_to_aging_of_orders"`
	MarginAccounting         float64 `yaml:"margin_accounting"`
	PartCharge               float64 `yaml:"part_charge"`
	PartResale               float64 `yaml:"part_resale"`
	PartBuffer               float64 `yaml:"part_buffer"`
	OrdersChangePercent      float64 `yaml:"orders_change_percent"`
	DealsChangePercent       float64 `yaml:"deals_change_percent"`
	WorkerLimitChangePercent float64 `yaml:"worker_limit_change_percent"`
	BadWorkersPercent        float64 `yaml:"bad_workers_percent"`
	WaitingTimeCRSec         int64   `yaml:"waiting_time_change_request"`
}
type otherParameters struct {
	IdentityForBid sonm.IdentityLevel `yaml:"identity_for_bid"`
	EmailForPool   string             `yaml:"email_for_pool"`
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
	// TODO: лучше сделать время
	TradeTicker int64 `yaml:"trade_tickers"`
	DataUpdate  int64 `yaml:"data_update"`
	PoolInit    int64 `yaml:"pool_init"`
}

type Config struct {
	Market            marketConfig          `yaml:"market" required:"true"`
	Database          databaseConfig        `yaml:"database"`
	PoolAddress       poolAddressesConfig   `yaml:"pool_accounts" required:"false"`
	UsingToken        usingToken            `yaml:"using_token"`
	Distances         stepsConfig           `yaml:"step_for_token"`
	ChargeIntervalETH chargeOrdersETHConfig `yaml:"charge_orders_interval"`
	ChargeIntervalZEC chargeOrdersZECConfig `yaml:"charge_orders_ZEC_interval"`
	ChargeIntervalXMR chargeOrdersXMRConfig `yaml:"charge_orders_XMR_interval"`
	Sensitivity       sensitivityConfig     `yaml:"sensitivity"`
	Images            imageConfig           `yaml:"images"`
	Tickers           tickerConfig          `yaml:"tickers"`
	OtherParameters   otherParameters       `yaml:"other_parameters"`
	Benchmark         typicalBenchmark      `yaml:"benchmark"`
	Eth               accounts.EthConfig    `yaml:"ethereum" required:"true"`
	Log               logging.Config        `yaml:"log"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := configor.Load(cfg, path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
