package antifraud

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type ProcessorConfig struct {
	Format          string        `yaml:"format" required:""`
	TrackInterval   time.Duration `yaml:"track_interval" default:"10s"`
	TaskWarmupDelay time.Duration `yaml:"warmup_delay" required:"true"`
}

type Config struct {
	TaskQuality            float64          `yaml:"task_quality" required:"true"`
	QualityCheckInterval   time.Duration    `yaml:"quality_check_interval" default:"15s"`
	BlacklistCheckInterval time.Duration    `yaml:"blacklist_check_interval" default:"5m"`
	ConnectionTimeout      time.Duration    `yaml:"connection_timeout" default:"60s"`
	LogProcessorConfig     ProcessorConfig  `yaml:"log_processor"`
	PoolProcessorConfig    ProcessorConfig  `yaml:"pool_processor"`
	Whitelist              []common.Address `yaml:"whitelist"`
}
