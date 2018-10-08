package antifraud

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type ProcessorConfig struct {
	Format          string        `yaml:"format" required:"true"`
	TrackInterval   time.Duration `yaml:"track_interval" default:"10s"`
	TaskWarmupDelay time.Duration `yaml:"warmup_delay" required:"true"`
	DecayTime       float64       `yaml:"decay_time" required:"true"`
}

type LogProcessorConfig struct {
	ProcessorConfig `yaml:",inline"`
	Pattern         string  `yaml:"pattern" required:"true"`
	Field           int     `yaml:"field"`
	Multiplier      float64 `yaml:"multiplier" required:"true"`
	LogDir          string  `yaml:"log_dir"`
}

type PoolProcessorConfig struct {
	ProcessorConfig `yaml:",inline"`
}

type Config struct {
	TaskQuality            float64             `yaml:"task_quality" required:"true"`
	QualityCheckInterval   time.Duration       `yaml:"quality_check_interval" default:"15s"`
	BlacklistCheckInterval time.Duration       `yaml:"blacklist_check_interval" default:"5m"`
	ConnectionTimeout      time.Duration       `yaml:"connection_timeout" default:"60s"`
	LogProcessorConfig     LogProcessorConfig  `yaml:"log_processor"`
	PoolProcessorConfig    PoolProcessorConfig `yaml:"pool_processor"`
	Whitelist              []common.Address    `yaml:"whitelist"`
}

func (c Config) Validate() error {
	if c.LogProcessorConfig.DecayTime <= 0 {
		return fmt.Errorf("antifraud config: log_processor.decay_time value must be positive")
	}

	if c.PoolProcessorConfig.DecayTime <= 0 {
		return fmt.Errorf("antifraud config: pool_processor.decay_time value must be positive")
	}

	return nil
}
