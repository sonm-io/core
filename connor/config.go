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
	"github.com/sonm-io/core/connor/types"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
)

type priceControlConfig struct {
	Marginality           float64 `yaml:"marginality" required:"true"`
	OrderReplaceThreshold float64 `yaml:"order_replace_threshold" required:"true"`
	DealChangeRequest     float64 `yaml:"deal_change_request" required:"true"`
	DealCancelThreshold   float64 `yaml:"deal_cancel_threshold" required:"true"`
}

type marketConfig struct {
	// todo: (sshaman1101): allow to set multiple subsets for order placing
	Benchmark    string             `yaml:"benchmark" required:"true"`
	From         uint64             `yaml:"from" required:"true"`
	To           uint64             `yaml:"to" required:"true"`
	Step         uint64             `yaml:"step" required:"true"`
	Counterparty common.Address     `yaml:"counterparty"`
	PriceControl priceControlConfig `yaml:"price_control"`
	Benchmarks   map[string]uint64  `yaml:"benchmarks" required:"true"`

	benchmarkID int
}

type nodeConfig struct {
	Endpoint auth.Addr `json:"endpoint"`
}

type engineConfig struct {
	ConnectionTimeout   time.Duration     `yaml:"connection_timeout" default:"60s"`
	OrderWatchInterval  time.Duration     `yaml:"order_watch_interval" default:"10s"`
	TaskStartInterval   time.Duration     `yaml:"task_start_interval" default:"15s"`
	TaskStartTimeout    time.Duration     `yaml:"task_start_timeout" default:"3m"`
	TaskTrackInterval   time.Duration     `yaml:"task_track_interval" default:"15s"`
	TaskRestoreInterval time.Duration     `yaml:"task_restore_interval" default:"10s"`
	ContainerEnv        map[string]string `yaml:"container_env"`
}

type containerConfig struct {
	Image  string            `yaml:"image" required:"true"`
	Tag    string            `yaml:"tag" required:"true"`
	SSHKey string            `yaml:"ssh_key"`
	Env    map[string]string `yaml:"env"`
}

func (m *containerConfig) getTag() *sonm.TaskTag {
	return &sonm.TaskTag{Data: []byte(m.Tag)}
}

type Config struct {
	Node          nodeConfig         `yaml:"node"`
	Eth           accounts.EthConfig `yaml:"ethereum"`
	Market        marketConfig       `yaml:"market"`
	Container     containerConfig    `yaml:"container"`
	Log           logging.Config     `yaml:"log"`
	Engine        engineConfig       `yaml:"engine"`
	BenchmarkList benchmarks.Config  `yaml:"benchmarks"`
	AntiFraud     antifraud.Config   `yaml:"antifraud"`
	PriceSource   price.SourceConfig `yaml:"price_source"`

	Metrics string `yaml:"metrics" default:"127.0.0.1:14005"`
}

func (c *Config) validate() error {
	availablePools := map[string]bool{
		antifraud.PoolFormatDwarf:         true,
		antifraud.ProcessorFormatDisabled: true,
	}
	availableLogs := map[string]bool{
		antifraud.LogFormatClaymore:       true,
		antifraud.LogFormatXMRing:         true,
		antifraud.ProcessorFormatDisabled: true,
	}

	if c.PriceSource.Factory == nil {
		return fmt.Errorf("empty price_source section")
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

	if err := c.AntiFraud.Validate(); err != nil {
		return err
	}

	named, err := reference.ParseNormalizedNamed(c.Container.Image)
	if err != nil {
		return fmt.Errorf("cannot parse image name: %v", err)
	}

	c.Container.Image = named.String()
	return nil
}

func (c *Config) validateBenchmarks(list benchmarks.BenchList) error {
	required := list.MapByCode()
	if len(required) != len(c.Market.Benchmarks) {
		return fmt.Errorf("unexpected count, have %d, want %d", len(c.Market.Benchmarks), len(required))
	}

	if b, ok := required[c.Market.Benchmark]; !ok {
		return fmt.Errorf("unexpected value `%s` for market.benchmark option", c.Market.Benchmark)
	} else {
		c.Market.benchmarkID = int(b.ID)
	}

	for key := range required {
		if _, ok := c.Market.Benchmarks[key]; !ok {
			return fmt.Errorf("missing key %s", key)
		}
	}

	return nil
}

func (c *Config) getBaseBenchmarks() types.Benchmarks {
	return types.Benchmarks{
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

func (c *Config) backends() *backends {
	// todo: do not init backends on each call
	return &backends{
		processorFactory: antifraud.NewProcessorFactory(&c.AntiFraud),
		corderFactory:    types.NewCorderFactory(c.Container.Tag, c.Market.benchmarkID, c.Market.Counterparty),
		dealFactory:      types.NewDealFactory(c.Market.benchmarkID),
		priceProvider:    c.PriceSource.Init(c.Market.PriceControl.Marginality),
	}
}

func applyEnvTemplate(env map[string]string, dealID *sonm.BigInt) map[string]string {
	dealTag := "{DEAL_ID}"

	result := map[string]string{}
	for key, value := range env {
		result[key] = strings.Replace(value, dealTag, dealID.Unwrap().String(), -1)
	}

	return result
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
