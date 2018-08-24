package connor

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/configor"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/connor/antifraud"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
)

const (
	ethBenchmarkIndex  = 9
	zecBenchmarkIndex  = 10
	nullBenchmarkIndex = 11
	// XMR mining on CPU now uses
	// cpu-sysbench values
	xmrCpuBenchmarkIndex = 0
)

type miningConfig struct {
	Token           string         `yaml:"token" required:"true"`
	Image           string         `yaml:"image" required:"true"`
	Wallet          common.Address `yaml:"wallet" required:"true"`
	PoolReportURL   string         `yaml:"pool_report" required:"false"`
	PoolTrackingURL string         `yaml:"pool_tracking" required:"false"`

	TokenPrice tokenPriceConfig `yaml:"token_price"`
}

func (m *miningConfig) getTag() *sonm.TaskTag {
	return &sonm.TaskTag{Data: []byte(fmt.Sprintf("connor_%s", strings.ToLower(m.Token)))}
}

type priceControlConfig struct {
	Marginality           float64 `yaml:"marginality" required:"true"`
	OrderReplaceThreshold float64 `yaml:"order_replace_threshold" required:"true"`
	DealChangeRequest     float64 `yaml:"deal_change_request" required:"true"`
	DealCancelThreshold   float64 `yaml:"deal_cancel_threshold" required:"true"`
}

type marketConfig struct {
	// todo: (sshaman1101): allow to set multiple subsets for order placing
	FromHashRate uint64             `yaml:"from_hashrate" required:"true"`
	ToHashRate   uint64             `yaml:"to_hashrate" required:"true"`
	Step         uint64             `yaml:"step" required:"true"`
	PriceControl priceControlConfig `yaml:"price_control"`
	Benchmarks   map[string]uint64  `yaml:"benchmarks" required:"true"`
}

type nodeConfig struct {
	Endpoint auth.Addr `json:"endpoint"`
}

type engineConfig struct {
	ConnectionTimeout   time.Duration     `yaml:"connection_timeout" default:"30s"`
	OrderWatchInterval  time.Duration     `yaml:"order_watch_interval" default:"10s"`
	TaskStartInterval   time.Duration     `yaml:"task_start_interval" default:"15s"`
	TaskStartTimeout    time.Duration     `yaml:"task_start_timeout" default:"3m"`
	TaskTrackInterval   time.Duration     `yaml:"task_track_interval" default:"15s"`
	TaskRestoreInterval time.Duration     `yaml:"task_restore_interval" default:"10s"`
	ContainerEnv        map[string]string `yaml:"container_env"`
}

type tokenPriceConfig struct {
	PriceURL       string        `yaml:"price_url"`
	UpdateInterval time.Duration `yaml:"update_interval" default:"60s"`
}

type Config struct {
	Node          nodeConfig         `yaml:"node"`
	Eth           accounts.EthConfig `yaml:"ethereum"`
	Market        marketConfig       `yaml:"market"`
	Mining        miningConfig       `yaml:"mining"`
	Log           logging.Config     `yaml:"log"`
	Engine        engineConfig       `yaml:"engine"`
	BenchmarkList benchmarks.Config  `yaml:"benchmarks"`
	AntiFraud     antifraud.Config   `yaml:"antifraud"`

	Metrics string `yaml:"metrics" default:"127.0.0.1:14005"`
}

func (c *Config) validate() error {
	availableTokens := map[string]bool{
		"ETH":     true,
		"XMR_CPU": true,
		"NULL":    true, // null token is for testing purposes
	}
	availablePools := map[string]bool{
		antifraud.PoolFormatDwarf: true,
	}
	availableLogs := map[string]bool{
		antifraud.LogFormatClaymore: true,
	}

	if _, ok := availableTokens[c.Mining.Token]; !ok {
		return fmt.Errorf("unsupported token \"%s\"", c.Mining.Token)
	}

	if _, ok := availableLogs[c.AntiFraud.LogProcessorConfig.Format]; !ok {
		return fmt.Errorf("unsupported log processor \"%s\"", c.AntiFraud.LogProcessorConfig.Format)
	}

	if _, ok := availablePools[c.AntiFraud.PoolProcessorConfig.Format]; !ok {
		return fmt.Errorf("unsupported pool processor \"%s\"", c.AntiFraud.PoolProcessorConfig.Format)
	}

	if c.Market.PriceControl.Marginality == 0 {
		return fmt.Errorf("market.price_marginality cannot be zero")
	}

	named, err := reference.ParseNormalizedNamed(c.Mining.Image)
	if err != nil {
		return fmt.Errorf("cannot parse image name: %v", err)
	}

	c.Mining.Image = named.String()
	return nil
}

func (c *Config) validateBenchmarks(list benchmarks.BenchList) error {
	required := list.MapByCode()
	if len(required) != len(c.Market.Benchmarks) {
		return fmt.Errorf("unexpected count, have %d, want %d", len(c.Market.Benchmarks), len(required))
	}

	for key := range required {
		if _, ok := c.Market.Benchmarks[key]; !ok {
			return fmt.Errorf("missing key %s", key)
		}
	}

	return nil
}

func (c *Config) getBaseBenchmarks() Benchmarks {
	return Benchmarks{
		Values: []uint64{
			c.Market.Benchmarks["cpu-sysbench-multi"],
			c.Market.Benchmarks["cpu-sysbench-single"],
			c.Market.Benchmarks["cpu-cores"],
			c.Market.Benchmarks["ram-size"],
			c.Market.Benchmarks["storage-size"],
			c.Market.Benchmarks["net-download"],
			c.Market.Benchmarks["net-upload"],
			c.Market.Benchmarks["gpu-count"],
			c.Market.Benchmarks["gpu-mem"],
			c.Market.Benchmarks["gpu-eth-hashrate"],
			c.Market.Benchmarks["gpu-cash-hashrate"],
			c.Market.Benchmarks["gpu-redshift"],
		},
	}
}

func (c *Config) getTokenParams() *tokenParameters {
	priceProviderConfig := &price.ProviderConfig{
		Margin: c.Market.PriceControl.Marginality,
		URL:    c.Mining.TokenPrice.PriceURL,
	}

	processorFactory := antifraud.NewProcessorFactory(&c.AntiFraud)

	available := map[string]*tokenParameters{
		"ETH": {
			corderFactory:    NewCorderFactory(c.Mining.Token, ethBenchmarkIndex),
			dealFactory:      NewDealFactory(ethBenchmarkIndex),
			priceProvider:    price.NewEthPriceProvider(priceProviderConfig),
			processorFactory: processorFactory,
		},

		"XMR_CPU": {
			dealFactory:      NewDealFactory(xmrCpuBenchmarkIndex),
			corderFactory:    NewCorderFactory(c.Mining.Token, xmrCpuBenchmarkIndex),
			priceProvider:    price.NewXmrPriceProvider(priceProviderConfig),
			processorFactory: processorFactory,
		},

		"NULL": {
			dealFactory:   NewDealFactory(nullBenchmarkIndex),
			corderFactory: NewCorderFactory(c.Mining.Token, nullBenchmarkIndex),
			priceProvider: price.NewNullPriceProvider(priceProviderConfig),
			// todo: use stub factory, then replace with fake mining pool tracker
			// processorFactory: nil,
		},
	}

	return available[c.Mining.Token]
}

// containerEnv returns container's params according
// to required mining pool and mining image
func (c *Config) containerEnv(dealID *sonm.BigInt) map[string]string {
	workerID := "c" + dealID.Unwrap().String()
	ethAddr := strings.ToLower(c.Mining.Wallet.Hex())

	m := make(map[string]string)
	if c.Mining.Token == "ETH" && c.AntiFraud.PoolProcessorConfig.Format == antifraud.PoolFormatDwarf {
		poolAddr := fmt.Sprintf("%s/%s/%s", c.Mining.PoolReportURL, ethAddr, workerID)
		wallet := fmt.Sprintf("%s/%s", ethAddr, workerID)

		m = map[string]string{
			"WALLET": wallet,
			"POOL":   poolAddr,
		}
	}

	// apply extra env params from config
	for k, v := range c.Engine.ContainerEnv {
		m[k] = v
	}

	return m
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	if err := configor.Load(cfg, path); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
