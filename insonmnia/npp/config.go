package npp

import (
	"github.com/sonm-io/core/insonmnia/npp/relay"
	"github.com/sonm-io/core/insonmnia/npp/rendezvous"
)

type Config struct {
	Rendezvous rendezvous.Config `yaml:"rendezvous"`
	Relay      relay.Config      `yaml:"relay"`
}
