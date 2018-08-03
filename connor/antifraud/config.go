package antifraud

import "time"

type LogProcessorConfig struct {
	TrackInterval   time.Duration `yaml:"track_interval" default:"10s"`
	TaskWarmupDelay time.Duration `yaml:"warmup_delay" required:"true"`
}

type Config struct {
	TaskQuality          float64            `yaml:"task_quality" required:"true"`
	QualityCheckInterval time.Duration      `yaml:"quality_check_interval" default:"15s"`
	ConnectionTimeout    time.Duration      `yaml:"connection_timeout" default:"30s"`
	LogProcessorConfig   LogProcessorConfig `yaml:"log_processor"`
	PoolProcessorConfig  LogProcessorConfig `yaml:"pool_processor"`
}
