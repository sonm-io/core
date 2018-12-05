package npp

import (
	"time"

	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
)

// Config represents an NPP (NAT punching protocol) module configuration.
type Config struct {
	Rendezvous         rendezvous.Config `yaml:"rendezvous"`
	Relay              relay.Config      `yaml:"relay"`
	Backlog            int               `yaml:"backlog" default:"128"`
	MinBackoffInterval time.Duration     `yaml:"min_backoff_interval" default:"500ms"`
	MaxBackoffInterval time.Duration     `yaml:"max_backoff_interval" default:"8000ms"`
}
